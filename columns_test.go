package main

import (
	"bytes"
	"testing"
)

// Several cases below reuse zfs list's NAME/USED/AVAIL/REFER/MOUNTPOINT
// column shape purely as a familiar, readable fixture. Except where a case
// says otherwise, the values are fabricated to exercise a specific edge
// case (a blank cell, an accidental double space, ...) and aren't
// something zfs itself would actually produce.
func TestCompactColumns(t *testing.T) {
	tests := []struct {
		name     string
		lines    []string
		expected []string
	}{
		{
			name: "basic multi-column compaction",
			lines: []string{
				"NAME                 USED  AVAIL  REFER  MOUNTPOINT",
				"zroot                70.6G   154G    96K  legacy",
				"zroot/local/var-log  2.29G   154G  2.29G  legacy",
			},
			expected: []string{
				"NAME                 USED   AVAIL  REFER  MOUNTPOINT",
				"zroot                70.6G  154G   96K    legacy",
				"zroot/local/var-log  2.29G  154G   2.29G  legacy",
			},
		},
		{
			name: "wide padding is reclaimed",
			lines: []string{
				"NAME                                                                                               USED  AVAIL  REFER  MOUNTPOINT",
				"zroot                                                                                             70.6G   154G    96K  legacy",
			},
			expected: []string{
				"NAME    USED  AVAIL  REFER  MOUNTPOINT",
				"zroot  70.6G   154G    96K  legacy",
			},
		},
		{
			// Unlike the fabricated cases elsewhere in this file, this one
			// is real, unmodified zfs list output: NAME's padding is
			// fixed-width across all rows (so USED/AVAIL/REFER start at the
			// same offset regardless of name length) while USED/AVAIL/REFER's
			// own padding is on the left, right-aligning their values.
			// Compaction should shrink the columns without flipping any of
			// that alignment.
			name: "right-aligned numeric columns keep their alignment",
			lines: []string{
				"NAME                                                                                               USED  AVAIL  REFER  MOUNTPOINT",
				"zroot                                                                                             70.6G   154G    96K  legacy",
				"zroot/local                                                                                       46.8G   154G    96K  legacy",
				"zroot/local/data                                                                                  1.44G   154G  1.44G  legacy",
				"zroot/local/nix                                                                                   8.52G   154G  8.52G  legacy",
				"zroot/local/var-log                                                                               2.29G   154G  2.29G  legacy",
				"zroot/safe                                                                                        23.6G   154G    96K  legacy",
				"zroot/safe/home                                                                                    160K   154G    96K  legacy",
				"zroot/safe/root                                                                                   14.9G   154G  3.12G  legacy",
			},
			expected: []string{
				"NAME                  USED  AVAIL  REFER  MOUNTPOINT",
				"zroot                70.6G   154G    96K  legacy",
				"zroot/local          46.8G   154G    96K  legacy",
				"zroot/local/data     1.44G   154G  1.44G  legacy",
				"zroot/local/nix      8.52G   154G  8.52G  legacy",
				"zroot/local/var-log  2.29G   154G  2.29G  legacy",
				"zroot/safe           23.6G   154G    96K  legacy",
				"zroot/safe/home       160K   154G    96K  legacy",
				"zroot/safe/root      14.9G   154G  3.12G  legacy",
			},
		},
		{
			name: "ragged rows with differing field counts",
			lines: []string{
				"a  bb  ccc",
				"aa  b",
				"a",
			},
			expected: []string{
				"a   bb  ccc",
				"aa  b",
				"a",
			},
		},
		{
			// A blank value in the middle of an otherwise fixed-width table
			// (e.g. `docker ps` leaving a column empty) collapses into its
			// neighboring separators, so that row splits into one fewer
			// field than the others. Naively aligning fields by index would
			// then shift MOUNTPOINT left into REFER's column.
			name: "blank value in the middle of a fixed-width column recovers its slot",
			lines: []string{
				"NAME                                                                                               USED  AVAIL  REFER  MOUNTPOINT",
				"zroot                                                                                             70.6G   154G    96K  legacy",
				"zroot/local                                                                                       46.8G   154G    96K  legacy",
				"zroot/safe/home                                                                                    160K   154G         legacy",
			},
			expected: []string{
				"NAME              USED  AVAIL  REFER  MOUNTPOINT",
				"zroot            70.6G   154G    96K  legacy",
				"zroot/local      46.8G   154G    96K  legacy",
				"zroot/safe/home   160K   154G         legacy",
			},
		},
		{
			// realign matches one column at a time, so it should have no
			// trouble with two blank values in the same row, adjacent or
			// not, as long as each one is still bounded by real values it
			// can anchor against.
			name: "adjacent blank values are each recovered independently",
			lines: []string{
				"NAME                                                                                               USED  AVAIL  REFER  MOUNTPOINT",
				"zroot                                                                                             70.6G   154G    96K  legacy",
				"zroot/local                                                                                       46.8G   154G    96K  legacy",
				"zroot/safe/home                                                                                    160K                legacy",
			},
			expected: []string{
				"NAME              USED  AVAIL  REFER  MOUNTPOINT",
				"zroot            70.6G   154G    96K  legacy",
				"zroot/local      46.8G   154G    96K  legacy",
				"zroot/safe/home   160K                legacy",
			},
		},
		{
			// A blank leading value doesn't merge into a neighboring
			// separator the way a middle or trailing one does (there's no
			// separator before it to merge with), so it already survives as
			// a natural empty first field without needing realign at all.
			// It's still worth pinning down: colWidths correctly shrinks to
			// the widest *present* name once the widest one goes blank.
			name: "blank leading value survives as a natural empty field",
			lines: []string{
				"NAME                                                                                               USED  AVAIL  REFER  MOUNTPOINT",
				"zroot                                                                                             70.6G   154G    96K  legacy",
				"zroot/local                                                                                       46.8G   154G    96K  legacy",
				"                                                                                                   160K   154G    96K  legacy",
			},
			expected: []string{
				"NAME          USED  AVAIL  REFER  MOUNTPOINT",
				"zroot        70.6G   154G    96K  legacy",
				"zroot/local  46.8G   154G    96K  legacy",
				"              160K   154G    96K  legacy",
			},
		},
		{
			// A blank trailing value is stripped entirely by the initial
			// TrimRight before splitting even starts, so realign has to
			// reconstruct it from nothing via its "no more real fields"
			// branch rather than an offset mismatch. Also guards against
			// the recovered blank column leaving a dangling "  " separator
			// with nothing after it.
			name: "blank trailing value recovers without leaving trailing whitespace",
			lines: []string{
				"NAME                                                                                               USED  AVAIL  REFER  MOUNTPOINT",
				"zroot                                                                                             70.6G   154G    96K  legacy",
				"zroot/local                                                                                       46.8G   154G    96K  legacy",
				"zroot/safe/home                                                                                    160K   154G    96K",
			},
			expected: []string{
				"NAME              USED  AVAIL  REFER  MOUNTPOINT",
				"zroot            70.6G   154G    96K  legacy",
				"zroot/local      46.8G   154G    96K  legacy",
				"zroot/safe/home   160K   154G    96K",
			},
		},
		{
			// A value that happens to contain an accidental internal
			// double space (e.g. a quoted multi-word field) inflates that
			// one row's field count above every other row's. Only one row
			// reaching that count isn't enough corroboration to trust as a
			// reference shape, so realign must not attempt it — otherwise
			// every genuinely complete row would get padded with a
			// fabricated trailing column to match.
			name: "a lone row with an accidental double space doesn't corrupt the others",
			lines: []string{
				"NAME                                                                                               USED  AVAIL  REFER  MOUNTPOINT",
				"zroot                                                                                             70.6G   154G    96K  legacy",
				"zroot/local                                                                                       46.8G   154G    96K  legacy",
				"zroot/safe/home                                                                                    160K   154G    96K  legacy",
				"zroot/tmp                                                                                          160K   154G    96K  legacy  extra",
			},
			expected: []string{
				"NAME              USED  AVAIL  REFER  MOUNTPOINT",
				"zroot            70.6G   154G    96K  legacy",
				"zroot/local      46.8G   154G    96K  legacy",
				"zroot/safe/home   160K   154G    96K  legacy",
				"zroot/tmp         160K   154G    96K  legacy      extra",
			},
		},
		{
			// Known limitation: if every data row is missing the same
			// column, no row establishes its true shape except the header,
			// and a single sample isn't enough to trust as a reference (see
			// the double-space case above), so realign never attempts
			// recovery here. The column silently disappears instead of
			// coming back blank, and MOUNTPOINT's value ends up sharing
			// REFER's column slot in the header. This documents the current
			// behavior rather than asserting it's correct.
			name: "a column blank on every row can't be recovered from the header alone",
			lines: []string{
				"NAME                                                                                               USED  AVAIL  REFER  MOUNTPOINT",
				"zroot                                                                                             70.6G   154G         legacy",
				"zroot/local                                                                                       46.8G   154G         legacy",
				"zroot/safe/home                                                                                    160K   154G         legacy",
			},
			expected: []string{
				"NAME              USED  AVAIL  REFER   MOUNTPOINT",
				"zroot            70.6G   154G  legacy",
				"zroot/local      46.8G   154G  legacy",
				"zroot/safe/home   160K   154G  legacy",
			},
		},
		{
			name: "passthrough for lines with no multi-space run",
			lines: []string{
				"just a single-spaced sentence",
				"onetoken",
			},
			expected: []string{
				"just a single-spaced sentence",
				"onetoken",
			},
		},
		{
			name: "blank lines preserved",
			lines: []string{
				"a  b",
				"",
				"c  d",
			},
			expected: []string{
				"a  b",
				"",
				"c  d",
			},
		},
		{
			name: "unicode width column",
			lines: []string{
				"NAME  VALUE",
				"世界    x",
				"a       y",
			},
			expected: []string{
				"NAME  VALUE",
				"世界  x",
				"a     y",
			},
		},
		{
			name: "no trailing whitespace on last column",
			lines: []string{
				"a  bb   ",
				"aa  b",
			},
			expected: []string{
				"a   bb",
				"aa  b",
			},
		},
		{
			// Tools like helm ls / kubectl use text/tabwriter: each field is
			// padded with spaces to the column's max width and then followed
			// by a real tab, not more spaces. That padding disappears for
			// whichever row's value exactly equals the column's max width,
			// leaving a bare tab with no separating spaces at all. compactColumns
			// only recognizes runs of 2+ spaces as a column separator, so that
			// row fails to split and passes through with its tab untouched.
			name: "tab-separated fields where a value exactly fills the column (no padding before the tab)",
			lines: []string{
				"NAME      \tCOL2",
				"short     \tval1",
				"exactmatch\tval2",
			},
			expected: []string{
				"NAME        COL2",
				"short       val1",
				"exactmatch  val2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := compactColumns(tt.lines)

			if len(got) != len(tt.expected) {
				t.Fatalf("compactColumns(%v) = %v, want %v", tt.lines, got, tt.expected)
			}

			for i := range got {
				if got[i] != tt.expected[i] {
					t.Errorf("line %d: got %q, want %q", i, got[i], tt.expected[i])
				}
			}
		})
	}
}

// TestCompactColumnsWithTruncation shows the actual point of the feature:
// compacting wasted padding lets trailing columns survive width truncation
// that would otherwise cut them off entirely.
func TestCompactColumnsWithTruncation(t *testing.T) {
	lines := []string{
		"NAME                                                                                               USED  AVAIL  REFER  MOUNTPOINT",
		"zroot                                                                                             70.6G   154G    96K  legacy",
	}

	width := 40

	// Without compaction, MOUNTPOINT is nowhere near reachable at width 40.
	var plain bytes.Buffer
	if err := printLines(&plain, width, lines); err != nil {
		t.Fatalf("printLines returned an error: %v", err)
	}
	if bytes.Contains(plain.Bytes(), []byte("MOUNTPOINT")) {
		t.Fatalf("expected MOUNTPOINT to be truncated away without compaction, got %q", plain.String())
	}

	// With compaction first, it fits.
	compacted := compactColumns(lines)
	var withCompaction bytes.Buffer
	if err := printLines(&withCompaction, width, compacted); err != nil {
		t.Fatalf("printLines returned an error: %v", err)
	}
	if !bytes.Contains(withCompaction.Bytes(), []byte("MOUNTPOINT")) {
		t.Fatalf("expected MOUNTPOINT to survive with compaction, got %q", withCompaction.String())
	}
	if !bytes.Contains(withCompaction.Bytes(), []byte("legacy")) {
		t.Fatalf("expected legacy to survive with compaction, got %q", withCompaction.String())
	}
}
