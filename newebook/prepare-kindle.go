package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/mateusbraga/tools/executil"
)

var (
	workersTotal = runtime.NumCPU() // Number of concurrent musics being processed
)

var (
	fixMusicWorkerCh = make(chan string, 20)
	workersWaitGroup sync.WaitGroup
)

var filetypesSupported = map[string]bool{
	".html": true,
	".epub": true,
}

func main() {
	var rootDir string
	var isRecursive bool
	if len(os.Args) == 1 { // get rootDir from cmdline args or current dir
		pwd, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
		rootDir = pwd
	} else {
		rootDir = os.Args[1]
		if rootDir == "./..." {
			pwd, err := os.Getwd()
			if err != nil {
				log.Fatal(err)
			}
			rootDir = pwd
			isRecursive = true
		}
	}

	done := make(chan struct{})
	defer close(done)

	// walkFiles will produce filenames that prepareForKindleWorker will consume
	paths, errc := walkFiles(done, rootDir, isRecursive)

	// start workers
	workersWaitGroup.Add(workersTotal)
	for i := 0; i < workersTotal; i++ {
		go prepareForKindleWorker(done, paths)
	}

	// wait for all workers to complete:
	//   when everything from paths channel is processed
	workersWaitGroup.Wait()
	if err := <-errc; err != nil {
		log.Fatalln(err)
	}
}

func prepareForKindleWorker(done <-chan struct{}, paths <-chan string) {
	defer workersWaitGroup.Done()

	for path := range paths {
		prepareForKindle(path)
	}
}

func callEbookConvert(inputFile, outputFile string) {
	// ebook-convert file.html file2.mobi --filter-css 'font-family,color,margin-left,margin-right' --mobi-ignore-margins
	args := []string{inputFile, outputFile, "--filter-css", "font-family,color,margin-left,margin-right", "--mobi-ignore-margins"}

	ebookConvert := exec.Command("ebook-convert", args...)
	executil.MustRun(ebookConvert)
}

func prepareForKindle(path string) error {
	err := executil.HasExecutables("ebook-convert")
	if err != nil {
		return err
	}

	switch ext := filepath.Ext(path); ext {
	case ".html":
		newFile := path + ".mobi"

		callEbookConvert(path, newFile)

		htmlFolder := strings.TrimSuffix(path, ".html") + "_files"
		os.RemoveAll(htmlFolder)
		os.Remove(path)

		log.Printf("Derived '%v' from '%v'", newFile, path)
		return nil
	case ".epub":
		newFile := path + ".mobi"

		callEbookConvert(path, newFile)

		log.Printf("Derived '%v' from '%v'", newFile, path)
		return nil
	default:
		return fmt.Errorf("'%v' filetype '%v' is not supported", path, ext)
	}
}

func walkFiles(done <-chan struct{}, root string, isRecursive bool) (<-chan string, <-chan error) {
	paths := make(chan string)
	errc := make(chan error, 1)

	go func() {
		defer close(paths)
		errc <- filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				fmt.Println(err)
				return nil
			}

			if info.IsDir() {
				if isRecursive || path == root {
					fmt.Printf("Walk in '%v'\n", path)
					return nil
				} else {
					return filepath.SkipDir
				}
			}

			if ext := filepath.Ext(path); filetypesSupported[ext] {
				abs, _ := filepath.Abs(path)
				select {
				case paths <- abs:
				case <-done:
					return errors.New("walk canceled")
				}
			}
			return nil
		})
	}()
	return paths, errc
}
