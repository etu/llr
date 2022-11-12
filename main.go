package main // import "github.com/etu/llr"

import (
	"bufio"
	"flag"
	"fmt"
	"golang.org/x/crypto/ssh/terminal"
	"log"
	"os"
)

func main() {
	var debug bool
	var width uint
	var scanner *bufio.Scanner

	// Detect terminal size.
	detectedWidth, _, err := terminal.GetSize(1)

	// Error out if terminal size detection fails.
	if err != nil {
		log.Fatal(err)
	}

	// Register parameter flags.
	flag.BoolVar(&debug, "debug", false, "Enable or disable debug output")
	flag.UintVar(&width, "width", uint(detectedWidth), "Specify a fixed with instead of terminal width")
	flag.Parse()

	// If we get args, try to find a filename
	if len(flag.Args()) >= 1 {
		filename := flag.Args()[0]

		// If filename is "-", we can just consider it stdin. So just
		// ignore this filename if we actually have a filename.
		if filename != "-" {
			file, err := os.Open(filename)

			// Print error if it doesn't exist
			if err != nil {
				log.Fatal(err)
			}

			defer file.Close()

			scanner = bufio.NewScanner(file)
		}
	}

	// If we don't already have a scanner from reading a file... then
	// we fall back to reading stdin.
	if scanner == nil {
		scanner = bufio.NewScanner(os.Stdin)
	}

	for scanner.Scan() {
		line := scanner.Text()

		// If empty line, just print empty line
		if len(line) == 0 {
			fmt.Println()
		} else if len(line) > int(width) {
			fmt.Println(string(line[0:width]))
		} else {
			fmt.Println(line)
		}
	}
}
