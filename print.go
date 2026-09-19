package main

import (
	"bufio"
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

// truncateLine expands tabs and cuts the line down to at most width visible
// columns, passing ANSI escape sequences through without counting them.
func truncateLine(line string, width int) string {
	// Replace tab characters with eight spaces
	line = strings.ReplaceAll(line, "\t", "        ")

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

	return b.String()
}

// PrintLines prints the given lines to the given writer, truncating them to the given width.
// It returns an error if there was an error writing to the writer.
func printLines(w io.Writer, width int, lines []string) error {
	for i, line := range lines {
		// Skip the last line if it's empty
		if i == len(lines)-1 && len(line) == 0 {
			continue
		}

		// Write the line to the writer
		if _, err := fmt.Fprintln(w, truncateLine(line, width)); err != nil {
			return err
		}
	}

	return nil
}

// streamLines reads r line by line and writes each line, truncated to the
// given width, to w as soon as it has been read, so it works on sources that
// never reach EOF (e.g. `journalctl -f`). Output is only held back while more
// input is already waiting; it is flushed whenever the reader runs dry, so a
// line is never delayed until the next one arrives.
func streamLines(r io.Reader, w io.Writer, width int) error {
	in := bufio.NewReader(r)
	out := bufio.NewWriter(w)

	for {
		line, readErr := in.ReadString('\n')

		// A final line without a trailing newline is still printed, but
		// an empty remainder after the last newline is not a line.
		if len(line) > 0 {
			line = strings.TrimSuffix(line, "\n")
			if _, err := fmt.Fprintln(out, truncateLine(line, width)); err != nil {
				return err
			}
		}

		if readErr != nil || in.Buffered() == 0 {
			if err := out.Flush(); err != nil {
				return err
			}
		}

		if readErr == io.EOF {
			return nil
		}
		if readErr != nil {
			return readErr
		}
	}
}
