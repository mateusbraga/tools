package main

import (
	"log"
	"os"
	"runtime"
	"sync"

	"github.com/mateusbraga/tools/newmusic/music"
)

var (
	fixMusicWorkerTotal = runtime.NumCPU() // Number of concurrent musics being processed
)

var (
	fixMusicWorkerCh        = make(chan string, 20)
	fixMusicWorkerWaitGroup sync.WaitGroup
)

func main() {
	var rootDir string
	var isRecursive bool
	if len(os.Args) == 1 { // get rootDir from cmdline args or current dir
		pwd, err := os.Getwd()
		if err != nil {
			log.Fatal("Getwd", err)
		}
		rootDir = pwd
	} else {
		rootDir = os.Args[1]
		if rootDir == "./..." {
			pwd, err := os.Getwd()
			if err != nil {
				log.Fatal("Getwd", err)
			}
			rootDir = pwd
			isRecursive = true
		}
	}

	done := make(chan struct{})
	defer close(done)

	// findMusicFiles will produce filenames that processMusic will consume
	paths, errc := music.WalkFiles(done, rootDir, isRecursive)

	//startFixMusicWorkers(done, paths)
	fixMusicWorkerWaitGroup.Add(fixMusicWorkerTotal)
	for i := 0; i < fixMusicWorkerTotal; i++ {
		go fixMusicWorker(done, paths)
	}

	// wait for all fixMusicWorkers to complete
	fixMusicWorkerWaitGroup.Wait()
	if err := <-errc; err != nil {
		log.Fatalln("WalkFiles:", err)
	}
}

func fixMusicWorker(done <-chan struct{}, paths <-chan string) {
	defer fixMusicWorkerWaitGroup.Done()

	for path := range paths {
		music.FixMusic(path)
	}
}
