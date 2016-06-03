package music

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/mateusbraga/tools/executil"
)

const backupCopyExtension = ".newmusic_backup"

var musicFiletypesSupported = map[string]bool{
	".mp3":              true,
	".wma":              true,
	".flac":             true,
	".flv":              true,
	".mp4":              true,
	".webm":             true,
	".m4a":              true,
	".mkv":              true,
	".ogg":              true,
	backupCopyExtension: true,
}

func callLame(inputFile string, outputFile string) error {
	lame := exec.Command("lame", "-v", inputFile, outputFile)
	_, err := executil.RunWithVerboseError(lame)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func callMp3Gain(file string) error {
	mp3gain := exec.Command("mp3gain", "-r", "-k", "-T", file)
	_, err := executil.RunWithVerboseError(mp3gain)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func callCopy(inputFile string, outputFile string) error {
	cp := exec.Command("cp", inputFile, outputFile)
	_, err := executil.RunWithVerboseError(cp)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func callMove(inputFile string, outputFile string) error {
	move := exec.Command("mv", inputFile, outputFile)
	_, err := executil.RunWithVerboseError(move)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func atomicReplaceFile(newFile string, oldFile string) {
	fileWorkingCopy := oldFile + backupCopyExtension
	callMove(oldFile, fileWorkingCopy)

	callMove(newFile, oldFile)

	err := os.Remove(fileWorkingCopy)
	if err != nil {
		log.Println("os.Remove '%v': %v\n", fileWorkingCopy, err)
	}
}

func processMp3(fileWorkingCopy string) {
	// lame
	tempLameOutput := fileWorkingCopy + ".mp3"
	callLame(fileWorkingCopy, tempLameOutput)
	callMove(tempLameOutput, fileWorkingCopy)

	// mp3gain
	callMp3Gain(fileWorkingCopy)
}

func FixMusic(path string) (err error) {
	err = executil.HasExecutables("lame", "mp3gain", "mv", "cp")
	if err != nil {
		return err
	}

	// Get temp dir to process file
	tempDir, err := ioutil.TempDir("", "gomusic")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)

	// Create a working copy at temp dir
	originalFilename := filepath.Base(path)
	fileWorkingCopy := filepath.Join(tempDir, originalFilename)
	callCopy(path, fileWorkingCopy)

	newFile := path[:len(path)-len(filepath.Ext(path))] + ".mp3"

	switch ext := filepath.Ext(path); ext {
	case ".mp3":
		processMp3(fileWorkingCopy)

		//commit
		atomicReplaceFile(fileWorkingCopy, path)

		log.Printf("Derived '%v' from '%v'", path, path)
		return nil
	case ".wma":
		newWorkingCopy := fileWorkingCopy[:len(fileWorkingCopy)-len(".wma")] + ".mp3"

		// convert wma -> mp3
		ffmpeg := exec.Command("ffmpeg", "-i", fileWorkingCopy, "-map_metadata", "0:s:0", "-acodec", "libmp3lame", newWorkingCopy)
		_, err := executil.RunWithVerboseError(ffmpeg)
		if err != nil {
			return err
		}

		processMp3(newWorkingCopy)

		//commit
		callMove(newWorkingCopy, newFile)
		os.Remove(path)

		log.Printf("Derived '%v' from '%v'", newFile, path)
		return nil
	case ".flac":
		newWorkingCopy := fileWorkingCopy[:len(fileWorkingCopy)-len(".flac")] + ".mp3"
		tempNewFileWav := fileWorkingCopy[:len(fileWorkingCopy)-len(".flac")] + ".wav"

		// convert flac -> wav
		flac := exec.Command("flac", "-d", fileWorkingCopy, "-o", tempNewFileWav)
		_, err := executil.RunWithVerboseError(flac)
		if err != nil {
			return err
		}

		// convert wav -> mp3
		callLame(tempNewFileWav, newWorkingCopy)

		processMp3(newWorkingCopy)

		//commit
		callMove(newWorkingCopy, newFile)
		os.Remove(path)

		log.Printf("Derived '%v' from '%v'", newFile, path)
		return nil
	case ".flv":
		newWorkingCopy := fileWorkingCopy[:len(fileWorkingCopy)-len(".flv")] + ".mp3"

		// convert flv -> mp3
		ffmpeg := exec.Command("ffmpeg", "-i", fileWorkingCopy, "-map_metadata", "0:s:0", newWorkingCopy)
		_, err := executil.RunWithVerboseError(ffmpeg)
		if err != nil {
			return err
		}

		processMp3(newWorkingCopy)

		//comit
		callMove(newWorkingCopy, newFile)
		os.Remove(path)

		log.Printf("Derived '%v' from '%v'", newFile, path)
		return nil
	case ".mp4":
		newWorkingCopy := fileWorkingCopy[:len(fileWorkingCopy)-len(".mp4")] + ".mp3"

		// convert mp4 -> mp3
		ffmpeg := exec.Command("ffmpeg", "-i", fileWorkingCopy, newWorkingCopy)
		_, err := executil.RunWithVerboseError(ffmpeg)
		if err != nil {
			return err
		}

		processMp3(newWorkingCopy)

		//commit
		callMove(newWorkingCopy, newFile)
		os.Remove(path)

		log.Printf("Derived '%v' from '%v'", newFile, path)
		return nil
	case ".webm":
		newWorkingCopy := fileWorkingCopy[:len(fileWorkingCopy)-len(".webm")] + ".mp3"

		// convert webm -> mp3
		ffmpeg := exec.Command("ffmpeg", "-i", fileWorkingCopy, newWorkingCopy)
		_, err := executil.RunWithVerboseError(ffmpeg)
		if err != nil {
			return err
		}

		processMp3(newWorkingCopy)

		//commit
		callMove(newWorkingCopy, newFile)
		os.Remove(path)

		log.Printf("Derived '%v' from '%v'", newFile, path)
		return nil
	case ".m4a":
		newWorkingCopy := fileWorkingCopy[:len(fileWorkingCopy)-len(".m4a")] + ".mp3"

		// convert m4a -> mp3
		ffmpeg := exec.Command("ffmpeg", "-i", fileWorkingCopy, "-map_metadata", "0:s:0", newWorkingCopy)
		_, err := executil.RunWithVerboseError(ffmpeg)
		if err != nil {
			return err
		}

		processMp3(newWorkingCopy)

		//commit
		callMove(newWorkingCopy, newFile)
		os.Remove(path)

		log.Printf("Derived '%v' from '%v'", newFile, path)
		return nil
	case ".mkv":
		newWorkingCopy := fileWorkingCopy[:len(fileWorkingCopy)-len(".mkv")] + ".mp3"

		// convert m4a -> mp3
		ffmpeg := exec.Command("ffmpeg", "-i", fileWorkingCopy, newWorkingCopy)
		_, err := executil.RunWithVerboseError(ffmpeg)
		if err != nil {
			return err
		}

		processMp3(newWorkingCopy)

		//commit
		callMove(newWorkingCopy, newFile)
		os.Remove(path)

		log.Printf("Derived '%v' from '%v'", newFile, path)
		return nil
	case ".ogg":
		newWorkingCopy := fileWorkingCopy[:len(fileWorkingCopy)-len(".ogg")] + ".mp3"

		// convert ogg -> mp3
		ffmpeg := exec.Command("ffmpeg", "-i", fileWorkingCopy, "-map_metadata", "0:s:0", newWorkingCopy)
		_, err := executil.RunWithVerboseError(ffmpeg)
		if err != nil {
			return err
		}

		processMp3(newWorkingCopy)

		//commit
		callMove(newWorkingCopy, newFile)
		os.Remove(path)

		log.Printf("Derived '%v' from '%v'", newFile, path)
		return nil
	case backupCopyExtension:
		originalFile := path[0 : len(path)-len(backupCopyExtension)]
		callMove(path, originalFile)
		log.Printf("Recovered '%v' from '%v'", originalFile, path)
		FixMusic(originalFile)
		return nil
	default:
		return errors.New(fmt.Sprintf("'%v' file extension is not supported", ext))
	}
}

func WalkFiles(done <-chan struct{}, root string, isRecursive bool) (<-chan string, <-chan error) {
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

			if ext := filepath.Ext(path); musicFiletypesSupported[ext] {
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
