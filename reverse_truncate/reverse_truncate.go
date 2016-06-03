package main

import (
	"fmt"
	"io/ioutil"

	"github.com/pivotal-golang/bytefmt"
)
import "flag"

func main() {
	targetSize := flag.String("size", "", "Target size for input file")
	flag.Parse()

	if *targetSize == "" {
		fmt.Println("flag size must be specified.\n\t Example: reverse_truncate --size 5M large.log")
		return
	}
	targetSizeInBytes, err := bytefmt.ToBytes(*targetSize)
	if err != nil {
		fmt.Printf("invalid required flag size: %s.\n\t example: reverse_truncate --size 5m large.log\n", err)
		return
	}
	if targetSizeInBytes <= 0 {
		fmt.Printf("invalid required flag size: %s.\n\t Example: reverse_truncate --size 5M large.log\n")
		return
	}

	var filepath string
	if len(flag.Args()) == 1 {
		filepath = flag.Args()[0]
	} else {
		fmt.Println("You need to give the file to be truncated.\n\t Example: reverse_truncate large.log")
		return
	}

	bytes, err := ioutil.ReadFile(filepath)
	originalSize := uint64(len(bytes))
	if originalSize < targetSizeInBytes {
		return
	}

	err = ioutil.WriteFile(filepath+".backup", bytes, 0644)
	if err != nil {
		fmt.Println("Failed to write backup file before truncating file: ", err)
		return
	}

	bytesToDelete := originalSize - targetSizeInBytes
	//bytesUntilNextNewline := 0
	//for ch := range bytes[bytesToDelete:] {
	//bytesUntilNextNewline++
	//if ch == "\n" {
	//break
	//}
	//}

	//bytesToDelete += bytesUntilNextNewline

	err = ioutil.WriteFile(filepath, bytes[bytesToDelete:], 0644)
	if err != nil {
		fmt.Println("Failed to write truncated file: ", err)
		return
	}
}
