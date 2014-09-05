package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("Usage: reducepdf file.pdf\n")
	}
	inputFile := os.Args[1]

	inputFileinfo, err := os.Stat(inputFile)
	if err != nil {
		log.Fatalf("Could not stat input file %v: %v", inputFile, err)
	}

	outputFile := inputFile[:len(inputFile)-len(".pdf")] + " - compressed.pdf"

	reducePdfSizeUsingGhostScript(inputFile, outputFile)

	outputFileinfo, err := os.Stat(outputFile)
	if err != nil {
		log.Fatalf("Could not stat input file %v: %v", inputFile, err)
	}

	log.Printf("\tReduced size of %v from %v to %v. %v created.\n", inputFile, inputFileinfo, outputFileinfo, outputFile)
}

func reducePdfSizeUsingGhostScript(inputFile string, outputFile string) {
	// gs -sDEVICE=pdfwrite -dCompatibilityLevel=1.4 -dPDFSETTINGS=/printer -dNOPAUSE -dQUIET -dBATCH -sOutputFile=output.pdf input.pdf
	outputFileArg := fmt.Sprintf("-sOutputFile=%v", outputFile)
	args := []string{"-sDEVICE=pdfwrite", "-dCompatibilityLevel=1.4", "-dPDFSETTINGS=/printer", "-dNOPAUSE", "-dQUIET", "-dBATCH", outputFileArg, inputFile}

	gs := exec.Command("gs", args...)
	var gsOut bytes.Buffer
	var gsErr bytes.Buffer
	gs.Stdout = &gsOut
	gs.Stderr = &gsErr
	if err := gs.Run(); err != nil {
		log.Println("gs args:", args)
		log.Println("gs command:", gs)
		log.Printf("gs stdout:\n---Start---\n%v\n---End---", gsOut.String())
		log.Printf("gs stderr:\n---Start---\n%v\n---End---", gsErr.String())
		log.Fatalln("gs failed: ", err)
	}
}
