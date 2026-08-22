package main

import (
	"regexp"
	"strings"

	"github.com/mattn/go-runewidth"
)

var columnSeparator = regexp.MustCompile(" {2,}")

// field is a column value together with the byte offsets it occupied in
// its (tab-expanded, right-trimmed) source line, used to detect whether the
// column was originally left- or right-aligned.
type field struct {
	text  string
	start int
	end   int
}

// compactColumns detects space-aligned columns across the given lines and
// re-pads each column down to the width of its widest actual value,
// instead of whatever padding the original input used. Each column's
// original left/right alignment (e.g. right-aligned size columns like
// `zfs list`'s USED/AVAIL/REFER) is preserved.
func compactColumns(lines []string) []string {
	fieldsPerLine := make([][]field, len(lines))
	var colWidths []int

	for i, line := range lines {
		// Expand tabs before detecting column boundaries. Tools like helm ls
		// or kubectl (via text/tabwriter) pad each field with spaces only up
		// to the column's max width and then emit a real tab, so a value
		// that exactly fills its column has no padding spaces at all before
		// the tab. Left as-is, that bare tab wouldn't match the 2+ space
		// separator below and the row would fail to split.
		line = strings.ReplaceAll(line, "\t", "        ")
		trimmed := strings.TrimRight(line, " ")
		seps := columnSeparator.FindAllStringIndex(trimmed, -1)

		if len(seps) == 0 {
			fieldsPerLine[i] = []field{{text: trimmed, start: 0, end: len(trimmed)}}
			continue
		}

		fields := make([]field, 0, len(seps)+1)
		start := 0
		for _, sep := range seps {
			fields = append(fields, field{text: trimmed[start:sep[0]], start: start, end: sep[0]})
			start = sep[1]
		}
		fields = append(fields, field{text: trimmed[start:], start: start, end: len(trimmed)})

		fieldsPerLine[i] = fields
		for j, f := range fields {
			w := runewidth.StringWidth(f.text)
			if j >= len(colWidths) {
				colWidths = append(colWidths, w)
			} else if w > colWidths[j] {
				colWidths[j] = w
			}
		}
	}

	// A column is right-aligned if, across every multi-column row, its
	// values consistently end at the same byte offset while starting at
	// varying offsets (i.e. the original padding was on the left).
	rightAligned := make([]bool, len(colWidths))
	for j := range colWidths {
		starts := make(map[int]struct{})
		ends := make(map[int]struct{})
		for _, fields := range fieldsPerLine {
			if len(fields) <= 1 || j >= len(fields) {
				continue
			}
			starts[fields[j].start] = struct{}{}
			ends[fields[j].end] = struct{}{}
		}
		rightAligned[j] = len(ends) == 1 && len(starts) > 1
	}

	result := make([]string, len(lines))
	for i, fields := range fieldsPerLine {
		if len(fields) <= 1 {
			result[i] = fields[0].text
			continue
		}

		var b strings.Builder
		for j, f := range fields {
			if j == len(fields)-1 {
				b.WriteString(f.text)
				continue
			}
			if rightAligned[j] {
				b.WriteString(runewidth.FillLeft(f.text, colWidths[j]))
			} else {
				b.WriteString(runewidth.FillRight(f.text, colWidths[j]))
			}
			b.WriteString("  ")
		}
		result[i] = b.String()
	}

	return result
}
