package main

import (
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"golang.org/x/crypto/ssh/terminal"
)

const defaultTerminalWidth = 80

// GetTerminalWidth gets the width of the terminal.
func getTerminalWidth() int {
	width, _, err := terminal.GetSize(int(os.Stdout.Fd()))

	if err == nil {
		return width
	}

	// Fall back to using 'stty size' if terminal dimensions cannot be retrieved
	cmd := exec.Command("stty", "size")
	cmd.Stdin = os.Stdin
	out, err := cmd.Output()

	if err != nil {
		// If stty fails (e.g., when running in watch or without a TTY),
		// fall back to the default width
		log.Printf("Warning: Unable to detect terminal size, using default width of %d", defaultTerminalWidth)
		return defaultTerminalWidth
	}

	dimensions := strings.Split(string(out), " ")

	if len(dimensions) != 2 {
		log.Printf("Warning: Invalid terminal size format, using default width of %d", defaultTerminalWidth)
		return defaultTerminalWidth
	}

	// Convert string to int
	width, err = strconv.Atoi(strings.TrimRight(dimensions[1], "\n"))

	if err != nil {
		log.Printf("Warning: Invalid terminal width value, using default width of %d", defaultTerminalWidth)
		return defaultTerminalWidth
	}

	return width
}
