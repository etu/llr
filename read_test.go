package main

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestReadFile(t *testing.T) {
	// Create a temporary file with some data
	tempFile, err := ioutil.TempFile("", "test-")

	if err != nil {
		t.Fatal(err)
	}

	defer os.Remove(tempFile.Name())

	_, err = tempFile.Write([]byte("Hello, world!"))

	if err != nil {
		t.Fatal(err)
	}

	tempFile.Close()

	// Test reading from the file
	contents, err := readFileOrStdin(tempFile.Name())

	if err != nil {
		t.Fatal(err)
	}

	if string(contents) != "Hello, world!" {
		t.Errorf("Unexpected contents: %s", contents)
	}
}

func TestReadStdin(t *testing.T) {
	// Set up a pipe to simulate stdin
	oldStdin := os.Stdin
	r, w, err := os.Pipe()

	if err != nil {
		t.Fatal(err)
	}

	os.Stdin = r
	defer func() {
		os.Stdin = oldStdin
	}()

	// Write some data to the pipe
	go func() {
		w.Write([]byte("Hello, world!"))
		w.Close()
	}()

	// Test reading from stdin
	contents, err := readFileOrStdin("-")

	if err != nil {
		t.Fatal(err)
	}

	if string(contents) != "Hello, world!" {
		t.Errorf("Unexpected contents: %s", contents)
	}
}

func TestReadInvalidFile(t *testing.T) {
	_, err := readFileOrStdin("nonexistent-file")

	if err == nil {
		t.Error("Expected an error, but none was returned")
	}
}
