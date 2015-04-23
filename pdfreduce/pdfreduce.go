package main

import (
	"flag"
	"fmt"
	"math"
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
		fmt.Printf("Usage: reducepdf file.pdf %v \n", flag.Args())
		os.Exit(1)
	}
	inputFile := flag.Arg(0)

	inputFileinfo, err := os.Stat(inputFile)
	if err != nil {
		fmt.Printf("Could not stat input file %v: %v", inputFile, err)
		os.Exit(1)
	}

	outputFile := inputFile[:len(inputFile)-len(".pdf")] + " - compressed.pdf"
	if *maxFlag {
		outputFile = inputFile[:len(inputFile)-len(".pdf")] + " - highly compressed.pdf"
	}

	reducePdfSizeUsingGhostScript(inputFile, outputFile, *maxFlag)

	outputFileinfo, err := os.Stat(outputFile)
	if err != nil {
		fmt.Printf("BUG: could not stat output file %v: %v", outputFile, err)
		os.Exit(1)
	}

	if outputFileinfo.Size() > int64(float64(inputFileinfo.Size())*(1-MINIMUM_FILESIZE_REDUCTION_EXPECTED)) {
		fmt.Println("Could not reduce the filesize significantly.")
		os.Remove(outputFile)
		return
	} else {
		humanReadableInputSize := HumanReadableSizeBytes(inputFileinfo.Size())
		humanReadableOutputSize := HumanReadableSizeBytes(outputFileinfo.Size())
		reductionPercentage := float64(outputFileinfo.Size()) / float64(inputFileinfo.Size())

		fmt.Printf("\tReduced size of '%v' from %v to %v (%.2f%%). '%v' created.\n", inputFile, humanReadableInputSize, humanReadableOutputSize, reductionPercentage, outputFile)
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

func HumanReadableSizeBytes(size int64) string {
	sizeFloat := float64(size)

	magnitude := math.Floor(math.Log10(sizeFloat))

	switch {
	case magnitude >= 12:
		return fmt.Sprintf("%v%v", sizeFloat/math.Pow10(12), "TB")
	case magnitude >= 9:
		return fmt.Sprintf("%v%v", sizeFloat/math.Pow10(9), "GB")
	case magnitude >= 6:
		return fmt.Sprintf("%v%v", sizeFloat/math.Pow10(6), "MB")
	case magnitude >= 3:
		return fmt.Sprintf("%v%v", sizeFloat/math.Pow10(3), "KB")
	default:
		return fmt.Sprintf("%v%v", size, "B")
	}
}
