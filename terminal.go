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

	// Try stty size with /dev/tty (works in watch and other pseudo-TTY environments)
	if tty, err := os.Open("/dev/tty"); err == nil {
		defer tty.Close()
		cmd := exec.Command("stty", "size")
		cmd.Stdin = tty
		if out, err := cmd.Output(); err == nil {
			dimensions := strings.Split(strings.TrimSpace(string(out)), " ")
			if len(dimensions) == 2 {
				if w, err := strconv.Atoi(dimensions[1]); err == nil && w > 0 {
					return w
				}
			}
		}
	}

	// Try tput cols command (another reliable method)
	if out, err := exec.Command("tput", "cols").Output(); err == nil {
		if w, err := strconv.Atoi(strings.TrimSpace(string(out))); err == nil && w > 0 {
			return w
		}
	}

	// Try COLUMNS environment variable
	if cols := os.Getenv("COLUMNS"); cols != "" {
		if w, err := strconv.Atoi(cols); err == nil && w > 0 {
			return w
		}
	}

	// If all methods fail, fall back to the default width
	log.Printf("Warning: Unable to detect terminal size, using default width of %d", defaultTerminalWidth)
	return defaultTerminalWidth
}
