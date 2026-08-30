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

// splitFields splits a single (already tab-expanded, right-trimmed) line
// into fields at runs of 2+ spaces, reporting whether any such run was
// found at all (a false result means the line is passed through untouched
// rather than treated as part of a column block).
func splitFields(line string) ([]field, bool) {
	seps := columnSeparator.FindAllStringIndex(line, -1)
	if len(seps) == 0 {
		return []field{{text: line, start: 0, end: len(line)}}, false
	}

	fields := make([]field, 0, len(seps)+1)
	start := 0
	for _, sep := range seps {
		fields = append(fields, field{text: line[start:sep[0]], start: start, end: sep[0]})
		start = sep[1]
	}
	fields = append(fields, field{text: line[start:], start: start, end: len(line)})
	return fields, true
}

// columnAnchor records the byte offset that stays constant for one column
// across every full row, and which edge of the value (start for
// left-aligned, end for right-aligned) that offset is measured from. In a
// fixed-width table each column occupies the same byte range on every row,
// so whichever edge isn't padded away is a reliable fingerprint for "this
// value belongs to this column".
type columnAnchor struct {
	rightAligned bool
	offset       int
}

// columnAnchors derives one anchor per column from the rows that already
// split into the full column count.
func columnAnchors(fieldsPerLine [][]field, fullRows []int, count int) []columnAnchor {
	anchors := make([]columnAnchor, count)
	for j := range count {
		starts := make(map[int]struct{})
		ends := make(map[int]struct{})
		for _, i := range fullRows {
			starts[fieldsPerLine[i][j].start] = struct{}{}
			ends[fieldsPerLine[i][j].end] = struct{}{}
		}
		if len(ends) == 1 && len(starts) > 1 {
			anchors[j] = columnAnchor{rightAligned: true, offset: fieldsPerLine[fullRows[0]][j].end}
		} else {
			anchors[j] = columnAnchor{rightAligned: false, offset: fieldsPerLine[fullRows[0]][j].start}
		}
	}
	return anchors
}

// realign expands a short row (one missing a blank value somewhere in the
// middle, e.g. `zfs list` leaving REFER empty) out to len(anchors) fields
// by matching each of its fields against the column anchors in order: a
// field whose stable edge doesn't land on the next column's anchor means
// that column was blank on this row, so an empty field is inserted for it
// instead. It reports ok=false if the fields can't be matched up cleanly,
// in which case the caller should leave the row as-is (genuinely ragged
// data rather than a fixed-width table with a blank cell).
func realign(fields []field, anchors []columnAnchor) ([]field, bool) {
	result := make([]field, 0, len(anchors))
	i := 0
	end := 0
	for _, a := range anchors {
		if i < len(fields) {
			f := fields[i]
			offset := f.start
			if a.rightAligned {
				offset = f.end
			}
			if offset == a.offset {
				result = append(result, f)
				end = f.end
				i++
				continue
			}
		}
		result = append(result, field{text: "", start: end, end: end})
	}

	if i != len(fields) {
		return nil, false
	}
	return result, true
}

// compactColumns detects space-aligned columns across the given lines and
// re-pads each column down to the width of its widest actual value,
// instead of whatever padding the original input used. Each column's
// original left/right alignment (e.g. right-aligned size columns like
// `zfs list`'s USED/AVAIL/REFER) is preserved.
func compactColumns(lines []string) []string {
	fieldsPerLine := make([][]field, len(lines))
	hadSeps := make([]bool, len(lines))
	maxFieldCount := 0

	for i, line := range lines {
		// Expand tabs before detecting column boundaries. Tools like helm ls
		// or kubectl (via text/tabwriter) pad each field with spaces only up
		// to the column's max width and then emit a real tab, so a value
		// that exactly fills its column has no padding spaces at all before
		// the tab. Left as-is, that bare tab wouldn't match the 2+ space
		// separator below and the row would fail to split.
		line = strings.ReplaceAll(line, "\t", "        ")
		trimmed := strings.TrimRight(line, " ")

		fields, ok := splitFields(trimmed)
		fieldsPerLine[i] = fields
		hadSeps[i] = ok
		if ok && len(fields) > maxFieldCount {
			maxFieldCount = len(fields)
		}
	}

	// Rows that split into fewer than the full column count are missing a
	// blank value somewhere. Recover it by matching each such row against
	// the column anchors established by rows that did split fully. A
	// single full row isn't enough corroboration to trust as a reference:
	// e.g. one row whose value happens to contain an accidental double
	// space inflates maxFieldCount for that row alone, and matching every
	// other (genuinely complete) row against it would just pad them all
	// with a fabricated trailing column.
	var fullRows []int
	for i, fields := range fieldsPerLine {
		if hadSeps[i] && len(fields) == maxFieldCount {
			fullRows = append(fullRows, i)
		}
	}

	if maxFieldCount > 0 && len(fullRows) >= 2 {
		anchors := columnAnchors(fieldsPerLine, fullRows, maxFieldCount)
		for i, fields := range fieldsPerLine {
			if !hadSeps[i] || len(fields) == maxFieldCount {
				continue
			}
			if canonical, ok := realign(fields, anchors); ok {
				fieldsPerLine[i] = canonical
			}
		}
	}

	var colWidths []int
	for i, fields := range fieldsPerLine {
		if !hadSeps[i] {
			continue
		}
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
	// varying offsets (i.e. the original padding was on the left). Blank
	// values recovered by realign carry no alignment information, so they're
	// excluded from this check.
	rightAligned := make([]bool, len(colWidths))
	for j := range colWidths {
		starts := make(map[int]struct{})
		ends := make(map[int]struct{})
		for i, fields := range fieldsPerLine {
			if !hadSeps[i] || j >= len(fields) || fields[j].text == "" {
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
		// A trailing column recovered as blank (e.g. a missing MOUNTPOINT)
		// leaves its preceding separator with nothing after it.
		result[i] = strings.TrimRight(b.String(), " ")
	}

	return result
}
