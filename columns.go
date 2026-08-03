package main

import (
	"regexp"
	"strings"

	"github.com/mattn/go-runewidth"
)

var columnSeparator = regexp.MustCompile(" {2,}")

// compactColumns detects space-aligned columns across the given lines and
// re-pads each column down to the width of its widest actual value,
// instead of whatever padding the original input used.
func compactColumns(lines []string) []string {
	fieldsPerLine := make([][]string, len(lines))
	var colWidths []int

	for i, line := range lines {
		trimmed := strings.TrimRight(line, " ")
		fields := columnSeparator.Split(trimmed, -1)

		if len(fields) <= 1 {
			fieldsPerLine[i] = []string{trimmed}
			continue
		}

		fieldsPerLine[i] = fields
		for j, field := range fields {
			w := runewidth.StringWidth(field)
			if j >= len(colWidths) {
				colWidths = append(colWidths, w)
			} else if w > colWidths[j] {
				colWidths[j] = w
			}
		}
	}

	result := make([]string, len(lines))
	for i, fields := range fieldsPerLine {
		if len(fields) <= 1 {
			result[i] = fields[0]
			continue
		}

		var b strings.Builder
		for j, field := range fields {
			if j == len(fields)-1 {
				b.WriteString(field)
				continue
			}
			b.WriteString(runewidth.FillRight(field, colWidths[j]))
			b.WriteString("  ")
		}
		result[i] = b.String()
	}

	return result
}
