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
