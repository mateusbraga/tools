package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/mateusbraga/tools/executil"
)

const (
	// Expects at least a 10% reduction of the filesize
	MINIMUM_FILESIZE_REDUCTION_EXPECTED = 0.1
)

func main() {
	maxFlag := flag.Bool("max", false, "Try to reduce size to the maximum")
	flag.Parse()

	if len(flag.Args()) != 1 {
		log.Fatalf("Usage: reducepdf file.pdf %v \n", flag.Args())
	}
	inputFile := flag.Arg(0)

	inputFileinfo, err := os.Stat(inputFile)
	if err != nil {
		log.Fatalf("Could not stat input file %v: %v", inputFile, err)
	}

	outputFile := inputFile[:len(inputFile)-len(".pdf")] + " - compressed.pdf"
	if *maxFlag {
		outputFile = inputFile[:len(inputFile)-len(".pdf")] + " - highly compressed.pdf"
	}

	reducePdfSizeUsingGhostScript(inputFile, outputFile, *maxFlag)

	outputFileinfo, err := os.Stat(outputFile)
	if err != nil {
		log.Fatalf("BUG: could not stat output file %v: %v", outputFile, err)
	}

	if outputFileinfo.Size() > int64(float64(inputFileinfo.Size())*(1-MINIMUM_FILESIZE_REDUCTION_EXPECTED)) {
		log.Println("Could not reduce the filesize significantly.")
		os.Remove(outputFile)
		return
	} else {
		log.Printf("\tReduced size of '%v' from %v to %v. '%v' created.\n", inputFile, inputFileinfo.Size(), outputFileinfo.Size(), outputFile)
	}
}

func reducePdfSizeUsingGhostScript(inputFile string, outputFile string, maxFlag bool) {
	// gs -sDEVICE=pdfwrite -dCompatibilityLevel=1.4 -dPDFSETTINGS=/ebook -dNOPAUSE -dQUIET -dBATCH -sOutputFile=output.pdf input.pdf
	outputFileArg := fmt.Sprintf("-sOutputFile=%v", outputFile)
	pdfSettings := "-dPDFSETTINGS=/ebook"
	if maxFlag {
		pdfSettings = "-dPDFSETTINGS=/screen"
	}
	args := []string{"-sDEVICE=pdfwrite", "-dCompatibilityLevel=1.4", pdfSettings, "-dNOPAUSE", "-dQUIET", "-dBATCH", outputFileArg, inputFile}

	gs := exec.Command("gs", args...)
	executil.MustRun(gs)
}
