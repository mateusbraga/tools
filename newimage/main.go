package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rwcarlsen/goexif/exif"
	"github.com/rwcarlsen/goexif/mknote"
)

var filetypesSupported = map[string]bool{
	".jpg": true,
	".mp4": true,
}

const (
	timestampLayout = "2006-01-02 15.04.05"
)

var (
	seqNumber = 1
)

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

	// walkFiles will produce filenames in lexical order
	paths, errc := walkFiles(done, rootDir, isRecursive)

	for path := range paths {
		err := prepareImage(path)
		if err != nil {
			fmt.Println(err)
		}
	}

	if err := <-errc; err != nil {
		log.Fatalln(err)
	}
}

func getExifDateTime(path string) (time.Time, error) {
	f, err := os.Open(path)
	if err != nil {
		return time.Time{}, err
	}
	defer f.Close()

	x, err := exif.Decode(f)
	if err != nil {
		return time.Time{}, fmt.Errorf("exif.Decode: %v", err)
	}

	tm, err := x.DateTime()
	if err != nil {
		return time.Time{}, err
	}

	return tm, nil
}

// newNameWithDate returns the name with the date from exif, if that does not exist it uses ModTime.
func newNameWithDate(path string) (string, error) {
	ext := filepath.Ext(path)
	dir := filepath.Dir(path)

	tm, err := getExifDateTime(path)
	if err != nil {
		fstat, err := os.Stat(path)
		if err != nil {
			return "", err
		}
		tm = fstat.ModTime()
	}

	newFile := fmt.Sprintf("%s/%s%s", dir, tm.Format(timestampLayout), ext)

	if path == newFile {
		// skip
		return newFile, nil
	}

	_, err = os.Stat(newFile)
	if err == nil {
		// already exists
		newFile = fmt.Sprintf("%s/%s_%d%s", dir, tm.Format(timestampLayout), seqNumber, ext)
		seqNumber++
	}

	return newFile, nil
}

func prepareImage(path string) error {
	switch ext := strings.ToLower(filepath.Ext(path)); ext {
	case ".jpg":
		newFile, err := newNameWithDate(path)

		if path == newFile {
			// skip
			return nil
		}

		_, err = os.Stat(newFile)
		if err == nil {
			// don't overwrite anything
			return err
		}

		err = os.Rename(path, newFile)
		if err != nil {
			return err
		}

		fmt.Printf("%v -> %v\n", filepath.Base(path), filepath.Base(newFile))
		return nil
	case ".mp4":
		newFile, err := newNameWithDate(path)

		if path == newFile {
			// skip
			return nil
		}

		_, err = os.Stat(newFile)
		if err == nil {
			// don't overwrite anything
			return err
		}

		err = os.Rename(path, newFile)
		if err != nil {
			return err
		}

		fmt.Printf("%v -> %v\n", filepath.Base(path), filepath.Base(newFile))
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

			if ext := filepath.Ext(path); filetypesSupported[strings.ToLower(ext)] {
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

func init() {
	exif.RegisterParsers(mknote.All...)
}
