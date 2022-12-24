package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/mattn/go-runewidth"
)

// PrintLines prints the given lines to the given writer, truncating them to the given width.
// It returns an error if there was an error writing to the writer.
func printLines(w io.Writer, width int, lines []string) error {
	for _, line := range lines {
		// Replace tab characters with eight spaces
		line = strings.Replace(line, "\t", "        ", -1)

		// Split the line into runes
		runes := []rune(line)

		// Iterate over the runes to count the rendered width of the line
		var lineWidth int
		var lineRunes []rune
		for _, r := range runes {
			lineWidth += runewidth.RuneWidth(r)
			if lineWidth > width {
				break
			}
			lineRunes = append(lineRunes, r)
		}

		// Write the line to the writer
		_, err := fmt.Fprintln(w, string(lineRunes))
		if err != nil {
			return err
		}
	}
	return nil
}
