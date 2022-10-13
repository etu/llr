package main // import "github.com/etu/llr"

import (
	"flag"
	"golang.org/x/crypto/ssh/terminal"
	"log"
)

func main() {
	var debug bool
	var width uint

	// Detect terminal size.
	detectedWidth, _, err := terminal.GetSize(0)

	// Error out if terminal size detection fails.
	if err != nil {
		log.Fatal(err)
	}

	// Register parameter flags.
	flag.BoolVar(&debug, "debug", false, "Enable or disable debug output")
	flag.UintVar(&width, "width", uint(detectedWidth), "Specify a fixed with instead of terminal width")
	flag.Parse()
}
