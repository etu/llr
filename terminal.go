package main

import (
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"golang.org/x/crypto/ssh/terminal"
)

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
		log.Fatal("Failed to execute stty size, got error: ", err)
	}

	dimensions := strings.Split(string(out), " ")

	if len(dimensions) != 2 {
		log.Fatal("Error parsing output of 'stty size'")
	}

	// Convert string to int
	width, err = strconv.Atoi(strings.TrimRight(dimensions[1], "\n"))

	if err != nil {
		log.Fatal("Error converting string to int:", err)
	}

	return width
}
