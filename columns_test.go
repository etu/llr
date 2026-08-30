package main

import (
	"bytes"
	"testing"
)

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
			// Real zfs list output: NAME's padding is fixed-width across all
			// rows (so USED/AVAIL/REFER start at the same offset regardless
			// of name length) while USED/AVAIL/REFER's own padding is on the
			// left, right-aligning their values. Compaction should shrink
			// the columns without flipping any of that alignment.
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
			// collapses into its neighboring separators, so that row splits
			// into one fewer field than the others. Naively aligning fields
			// by index would then shift MOUNTPOINT left into REFER's column.
			//
			// The zfs list shape is just reused here as a familiar example;
			// zfs itself never leaves REFER blank like this. Other tools'
			// output can, though, e.g. `docker ps`.
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
