package main

import (
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/mattn/go-runewidth"
)

// ansiEscape matches ANSI/VT100 CSI escape sequences, such as the SGR codes
// used for terminal colors, so they can be passed through without counting
// toward the visible width of the line.
var ansiEscape = regexp.MustCompile("\x1b\\[[0-9;]*[a-zA-Z]")

// PrintLines prints the given lines to the given writer, truncating them to the given width.
// It returns an error if there was an error writing to the writer.
func printLines(w io.Writer, width int, lines []string) error {
	for i, line := range lines {
		// Skip the last line if it's empty
		if i == len(lines)-1 && len(line) == 0 {
			continue
		}

		// Replace tab characters with eight spaces
		line = strings.Replace(line, "\t", "        ", -1)

		var b strings.Builder
		var lineWidth int
		var truncated bool
		var sawEscape bool

		pos := 0
		for pos < len(line) {
			loc := ansiEscape.FindStringIndex(line[pos:])
			nextEscapeStart := len(line)
			if loc != nil {
				nextEscapeStart = pos + loc[0]
			}

			// Count and copy the visible runes up to the next escape
			// sequence (or the end of the line), stopping once the width
			// budget is used up.
			for _, r := range line[pos:nextEscapeStart] {
				rw := runewidth.RuneWidth(r)
				if lineWidth+rw > width {
					truncated = true
					break
				}
				lineWidth += rw
				b.WriteRune(r)
			}
			if truncated {
				break
			}

			if loc == nil {
				break
			}

			// Escape sequences carry terminal state (e.g. color) rather
			// than visible content, so pass them through without spending
			// any of the width budget.
			start, end := pos+loc[0], pos+loc[1]
			b.WriteString(line[start:end])
			sawEscape = true
			pos = end
		}

		// If the line was cut off partway through and it contained escape
		// sequences, reset the terminal so a dangling color/style doesn't
		// bleed into whatever gets printed after it.
		if truncated && sawEscape {
			b.WriteString("\x1b[0m")
		}

		// Write the line to the writer
		_, err := fmt.Fprintln(w, b.String())
		if err != nil {
			return err
		}
	}

	return nil
}
