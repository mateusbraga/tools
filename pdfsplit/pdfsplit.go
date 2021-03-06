package main

import (
	"fmt"
	"log"
	"math"
	"os"
	"os/exec"
	"strconv"

	"github.com/mateusbraga/tools/executil"
)

const (
	MAX_NUMBER_OF_PAGES = 500
)

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("Usage: splitpdf file.pdf\n")
	}
	inputFile := os.Args[1]

	numberOfPages := getNumberOfPagesUsingGhostScript(inputFile)

	numberOfOutputFiles := int(math.Ceil(float64(numberOfPages) / float64(MAX_NUMBER_OF_PAGES)))

	log.Printf("Splitting %v in %v files of %v pages (total %v pages)\n", inputFile, numberOfOutputFiles, MAX_NUMBER_OF_PAGES, numberOfPages)
	for i := 0; i < numberOfOutputFiles; i++ {
		log.Printf("\tFrom %v to %v\n", i*MAX_NUMBER_OF_PAGES+1, (i+1)*MAX_NUMBER_OF_PAGES)
		outputFile := fmt.Sprintf("%d_%v", i+1, inputFile)
		splitUsingGhostScript(inputFile, i*MAX_NUMBER_OF_PAGES+1, (i+1)*MAX_NUMBER_OF_PAGES, outputFile)
	}
	log.Printf("Done\n")
}

func splitUsingGhostScript(inputFile string, initialPage, lastPage int, outputFile string) {
	//gs -sDEVICE=pdfwrite -dNOPAUSE -dBATCH -dSAFER -dFirstPage=1 -dLastPage=4 -sOutputFile=outputT4.pdf T4.pdf
	initialPageArg := fmt.Sprintf("-dFirstPage=%d", initialPage)
	lastPageArg := fmt.Sprintf("-dLastPage=%d", lastPage)
	outputFileArg := fmt.Sprintf("-sOutputFile=%v", outputFile)
	args := []string{"-sDEVICE=pdfwrite", "-dNOPAUSE", "-dBATCH", "-dSAFER", initialPageArg, lastPageArg, outputFileArg, inputFile}

	gs := exec.Command("gs", args...)
	executil.MustRun(gs)
}

func getNumberOfPagesUsingGhostScript(pdfFile string) int {
	//gs -q -dNODISPLAY -c "(Code Complete - Steve McConnel.pdf) (r) file runpdfbegin pdfpagecount = quit"
	cmdArg := fmt.Sprintf("(%v) (r) file runpdfbegin pdfpagecount = quit", pdfFile)
	args := []string{"-q", "-dNODISPLAY", "-c", cmdArg}

	gs := exec.Command("gs", args...)
	output := executil.MustRun(gs)

	// remove new line
	numberOfPages, err := strconv.ParseInt(output[:len(output)-1], 10, 0)
	if err != nil {
		log.Fatalln("Failed to get number of pages:", err)
	}
	return int(numberOfPages)
}
