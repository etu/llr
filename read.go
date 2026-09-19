package main

import (
	"io"
	"os"
)

// openInput opens the named file for reading, or stdin if the name is "-".
// The caller must Close the result; closing stdin's wrapper is a no-op.
func openInput(filename string) (io.ReadCloser, error) {
	if filename == "-" {
		return io.NopCloser(os.Stdin), nil
	}

	return os.Open(filename)
}

func readFileOrStdin(filename string) ([]byte, error) {
	reader, err := openInput(filename)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	// Read the file or stdin
	contents, err := io.ReadAll(reader)

	if err != nil {
		return nil, err
	}

	return contents, nil
}
