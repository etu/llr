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
				"NAME   USED   AVAIL  REFER  MOUNTPOINT",
				"zroot  70.6G  154G   96K    legacy",
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
