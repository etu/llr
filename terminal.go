package main

import (
	"log"

	"golang.org/x/crypto/ssh/terminal"
)

// GetTerminalWidth gets the width of the terminal.
func getTerminalWidth() int {
	// Detect terminal size.
	detectedWidth, _, err := terminal.GetSize(1)

	// Error out if terminal size detection fails.
	if err != nil {
		log.Fatal(err)
	}

	return detectedWidth
}
