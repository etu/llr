package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
)

// TestMainFlow tests the complete flow from reading to printing with various width overrides
func TestMainFlow(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		width    int
		expected string
	}{
		{
			name:     "simple text with default width",
			input:    "Hello\nWorld\n",
			width:    80,
			expected: "Hello\nWorld\n",
		},
		{
			name:     "long lines with width override",
			input:    "This is a very long line that should be truncated\nAnother long line\n",
			width:    20,
			expected: "This is a very long \nAnother long line\n",
		},
		{
			name:     "text with tabs",
			input:    "a\tb\tc\n",
			width:    20,
			expected: "a        b        c\n",
		},
		{
			name:     "empty input",
			input:    "",
			width:    80,
			expected: "",
		},
		{
			name:     "single line without newline",
			input:    "single line",
			width:    80,
			expected: "single line\n",
		},
		{
			name:     "width override with very narrow terminal",
			input:    "testing narrow width",
			width:    5,
			expected: "testi\n",
		},
		{
			name:     "width override with very wide terminal",
			input:    "This should fit completely in a very wide terminal window",
			width:    200,
			expected: "This should fit completely in a very wide terminal window\n",
		},
		{
			name:     "multiple lines with various lengths",
			input:    "short\nmedium length line\nvery very very long line that will be truncated\n",
			width:    30,
			expected: "short\nmedium length line\nvery very very long line that \n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the main flow
			lines := strings.Split(tt.input, "\n")
			
			var buf bytes.Buffer
			err := printLines(&buf, tt.width, lines)

			if err != nil {
				t.Errorf("printLines returned an error: %v", err)
			}

			if got := buf.String(); got != tt.expected {
				t.Errorf("got:\n%q\nwant:\n%q", got, tt.expected)
			}
		})
	}
}

// TestWidthOverride tests that width override works correctly
func TestWidthOverride(t *testing.T) {
	// Test different width values
	widths := []int{1, 5, 10, 20, 50, 80, 100, 200}
	input := "This is a test line that is reasonably long to test different width values"

	for _, width := range widths {
		t.Run(fmt.Sprintf("width_%d", width), func(t *testing.T) {
			lines := []string{input, ""}
			var buf bytes.Buffer
			
			err := printLines(&buf, width, lines)
			if err != nil {
				t.Errorf("printLines with width %d returned an error: %v", width, err)
			}

			output := buf.String()
			// Remove the trailing newline for length check
			outputLine := strings.TrimSuffix(output, "\n")
			
			// The output should not exceed the specified width
			// Note: we need to count runes, not bytes, and account for display width
			if len([]rune(outputLine)) > width {
				t.Errorf("Output length %d exceeds width %d: %q", len([]rune(outputLine)), width, outputLine)
			}
		})
	}
}

// TestReadAndPrint tests reading from a file and printing with width override
func TestReadAndPrint(t *testing.T) {
	// Create a temporary file
	content := []byte("Line 1: This is a test\nLine 2: Another line\nLine 3: Yet another line\n")
	tmpfile, err := os.CreateTemp("", "llr-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write(content); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	// Read the file
	readContent, err := readFileOrStdin(tmpfile.Name())
	if err != nil {
		t.Fatalf("readFileOrStdin failed: %v", err)
	}

	// Process and print with different widths
	widths := []int{10, 20, 80}
	for _, width := range widths {
		t.Run(fmt.Sprintf("width_%d", width), func(t *testing.T) {
			lines := strings.Split(string(readContent), "\n")
			var buf bytes.Buffer
			
			err := printLines(&buf, width, lines)
			if err != nil {
				t.Errorf("printLines with width %d failed: %v", width, err)
			}

			// Verify output is not empty
			if buf.Len() == 0 {
				t.Errorf("Expected non-empty output for width %d", width)
			}
		})
	}
}

// TestStdinRead tests reading from stdin
func TestStdinRead(t *testing.T) {
	// Set up a pipe to simulate stdin
	oldStdin := os.Stdin
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdin = r
	defer func() {
		os.Stdin = oldStdin
	}()

	// Write test data to the pipe
	testData := "Line 1\nLine 2\nLine 3\n"
	go func() {
		io.WriteString(w, testData)
		w.Close()
	}()

	// Read from stdin
	contents, err := readFileOrStdin("-")
	if err != nil {
		t.Fatalf("readFileOrStdin(\"-\") failed: %v", err)
	}

	if string(contents) != testData {
		t.Errorf("got %q, want %q", string(contents), testData)
	}

	// Process the content
	lines := strings.Split(string(contents), "\n")
	var buf bytes.Buffer
	err = printLines(&buf, 80, lines)
	if err != nil {
		t.Errorf("printLines failed: %v", err)
	}

	expected := "Line 1\nLine 2\nLine 3\n"
	if buf.String() != expected {
		t.Errorf("got %q, want %q", buf.String(), expected)
	}
}
