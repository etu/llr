package main

import (
	"bytes"
	"testing"
)

func TestPrintLines(t *testing.T) {
	tests := []struct {
		name     string
		width    int
		lines    []string
		expected string
	}{
		{
			name:     "empty lines",
			width:    10,
			lines:    []string{},
			expected: "",
		},
		{
			name:  "short lines",
			width: 10,
			lines: []string{
				"short",
				"line",
				"",
			},
			expected: "short\nline\n",
		},
		{
			name:  "short lines with missing line break on last line",
			width: 10,
			lines: []string{
				"short",
				"line",
			},
			expected: "short\nline\n",
		},
		{
			name:  "print first character",
			width: 1,
			lines: []string{
				"short",
				"line",
				"",
			},
			expected: "s\nl\n",
		},
		{
			name:  "print two first characters",
			width: 2,
			lines: []string{
				"short",
				"line",
				"",
			},
			expected: "sh\nli\n",
		},
		{
			name:  "long lines",
			width: 10,
			lines: []string{
				"this is a very long line",
				"this is another very long line",
				"",
			},
			expected: "this is a \nthis is an\n",
		},
		{
			name:  "tab characters",
			width: 10,
			lines: []string{
				"\tthis is a line with a tab character",
				"this is a line with\ta tab character",
				"",
			},
			expected: "        th\nthis is a \n",
		},
		{
			name:  "unicode characters",
			width: 10,
			lines: []string{
				"Hello 世界",
				"😀😃😄😁",
				"",
			},
			expected: "Hello 世界\n😀😃😄😁\n",
		},
		{
			name:  "zero width",
			width: 0,
			lines: []string{
				"test",
				"line",
				"",
			},
			expected: "\n\n",
		},
		{
			name:  "exact width match",
			width: 5,
			lines: []string{
				"exact",
				"words",
				"",
			},
			expected: "exact\nwords\n",
		},
		{
			name:  "width override with very wide terminal",
			width: 200,
			lines: []string{
				"This is a relatively short line that should fit completely",
				"Another line that fits",
				"",
			},
			expected: "This is a relatively short line that should fit completely\nAnother line that fits\n",
		},
		{
			name:  "width override with narrow terminal",
			width: 20,
			lines: []string{
				"This is a relatively long line that will be truncated",
				"Another very long line",
				"",
			},
			expected: "This is a relatively\nAnother very long li\n",
		},
		{
			name:  "multiple tabs in a line",
			width: 20,
			lines: []string{
				"a\tb\tc\td",
				"",
			},
			expected: "a        b        c \n",
		},
		{
			name:  "empty line in the middle",
			width: 10,
			lines: []string{
				"first",
				"",
				"third",
				"",
			},
			expected: "first\n\nthird\n",
		},
		{
			name:  "line with only spaces",
			width: 10,
			lines: []string{
				"     ",
				"text",
				"",
			},
			expected: "     \ntext\n",
		},
	}

	// Loop through tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			err := printLines(&buf, tt.width, tt.lines)

			if err != nil {
				t.Errorf("PrintLines returned an error: %v", err)
			}

			if got := buf.String(); got != tt.expected {
				t.Errorf("PrintLines(%d, %v) = %q, want %q", tt.width, tt.lines, got, tt.expected)
			}
		})
	}
}
