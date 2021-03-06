package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mateusbraga/tools/executil"
)

var filetypesSupported = map[string]bool{
	".jpg": true,
	".png": true,
	".pdf": true,
}

func CallConvert(inputFile string, outputFile string) {
	convert := exec.Command("convert", inputFile, outputFile)
	executil.MustRun(convert)
}

func CallGhostScript(outputFile string, orderedFiles []string) {
	args := []string{"-o", outputFile, "-sDEVICE=pdfwrite", "-dPDFSETTINGS=/prepress"}
	args = append(args, orderedFiles...)

	gs := exec.Command("gs", args...)
	executil.MustRun(gs)
}

func CallMove(inputFile string, outputFile string) {
	move := exec.Command("mv", inputFile, outputFile)
	executil.MustRun(move)
}

func MakePdf(files []string, outputFile string) {
	err := executil.HasExecutables("convert", "gs", "mv")
	if err != nil {
		log.Fatalln(err)
	}

	tempDir, err := ioutil.TempDir("", "gomakepdf")
	if err != nil {
		log.Fatalln(err)
	}
	defer os.RemoveAll(tempDir)

	// Check
	for _, file := range files {
		if isFiletypeSupported(file) {
			log.Fatalf("Extension of file '%v' is not supported.\n", file)
		}
	}
	if ext := filepath.Ext(outputFile); ext != ".pdf" {
		log.Fatalf("outputFile must be a pdf file: '%v'\n", outputFile)
	}

	var inputFiles []string
	// Convert/Prepare
	for _, path := range files {
		switch ext := filepath.Ext(path); ext {
		case ".png":
			pdfFile := path[:len(path)-len(".png")] + ".pdf"

			tempFile := filepath.Join(tempDir, pdfFile)

			CallConvert(path, tempFile)

			log.Printf("Derived '%v' from '%v'", tempFile, path)
			inputFiles = append(inputFiles, tempFile)
		case ".jpg":
			pdfFile := path[:len(path)-len(".jpg")] + ".pdf"

			tempFile := filepath.Join(tempDir, pdfFile)

			CallConvert(path, tempFile)

			log.Printf("Derived '%v' from '%v'", tempFile, path)
			inputFiles = append(inputFiles, tempFile)
		case ".pdf":
			inputFiles = append(inputFiles, path)
		}
	}

	// Create pdf
	CallGhostScript(outputFile, inputFiles)

	if len(files) < 10 {
		log.Printf("Done generating '%v' from '%v'", outputFile, files)
	} else {
		log.Printf("Done generating '%v' from %v files", outputFile, len(files))
	}
}

func isFiletypeSupported(filename string) bool {
	ext := filepath.Ext(filename)
	return !filetypesSupported[strings.ToLower(ext)]
}

func main() {
	finalOutputFile := flag.String("output", "output.pdf", "Set the pdf output file to be created")
	flag.Parse()

	var inputFiles []string
	if len(flag.Args()) > 0 {
		inputFiles = flag.Args()
	} else {
		fmt.Println("You need to give the list of files to be merged.\n\t Example: newpdf page1.jpg page2.jpg others.pdf")
		return
	}

	log.Printf("Merge %v files into %v", len(inputFiles), *finalOutputFile)
	MakePdf(inputFiles, *finalOutputFile)
}
