package main

import (
	"io"
	"io/ioutil"
	"os"
)

func readFileOrStdin(filename string) ([]byte, error) {
	// Open the file or stdin
	var reader io.Reader

	if filename == "-" {
		reader = os.Stdin
	} else {
		file, err := os.Open(filename)
		if err != nil {
			return nil, err
		}
		defer file.Close()
		reader = file
	}

	// Read the file or stdin
	contents, err := ioutil.ReadAll(reader)

	if err != nil {
		return nil, err
	}

	return contents, nil
}
