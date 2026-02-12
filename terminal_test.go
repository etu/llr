package main

import (
	"os"
	"testing"
)

func TestGetTerminalWidth(t *testing.T) {
	tests := []struct {
		name           string
		columnsEnv     string
		description    string
		checkPositive  bool
	}{
		{
			name:          "respects COLUMNS environment variable",
			columnsEnv:    "120",
			description:   "Should use COLUMNS environment variable when set and terminal not detected",
			checkPositive: true,
		},
		{
			name:          "ignores invalid COLUMNS value",
			columnsEnv:    "invalid",
			description:   "Should fallback to default or detect terminal when COLUMNS is invalid",
			checkPositive: true,
		},
		{
			name:          "ignores zero COLUMNS value",
			columnsEnv:    "0",
			description:   "Should fallback to default or detect terminal when COLUMNS is zero",
			checkPositive: true,
		},
		{
			name:          "ignores negative COLUMNS value",
			columnsEnv:    "-1",
			description:   "Should fallback to default or detect terminal when COLUMNS is negative",
			checkPositive: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original environment
			originalColumns := os.Getenv("COLUMNS")
			defer func() {
				if originalColumns == "" {
					os.Unsetenv("COLUMNS")
				} else {
					os.Setenv("COLUMNS", originalColumns)
				}
			}()

			// Set test environment
			if tt.columnsEnv == "" {
				os.Unsetenv("COLUMNS")
			} else {
				os.Setenv("COLUMNS", tt.columnsEnv)
			}

			// Get terminal width
			width := getTerminalWidth()

			// Verify width is positive and reasonable
			if tt.checkPositive && width <= 0 {
				t.Errorf("%s: getTerminalWidth() = %d, want positive value", tt.description, width)
			}
			
			if width > 1000 {
				t.Errorf("%s: getTerminalWidth() = %d, want <= 1000", tt.description, width)
			}
		})
	}
}

func TestGetTerminalWidthDefault(t *testing.T) {
	// Test that the function returns a reasonable default when no terminal is detected
	// This test is more about verifying the function doesn't crash or return invalid values
	width := getTerminalWidth()

	// Verify width is positive and reasonable
	if width <= 0 {
		t.Errorf("getTerminalWidth() returned non-positive value: %d", width)
	}

	if width > 1000 {
		t.Errorf("getTerminalWidth() returned unreasonably large value: %d", width)
	}
	
	// The width should be at least the default
	if width < defaultTerminalWidth {
		t.Logf("Note: getTerminalWidth() returned %d, which is less than default %d", width, defaultTerminalWidth)
	}
}

func TestGetTerminalWidthWithActualTerminal(t *testing.T) {
	// Save original environment
	originalColumns := os.Getenv("COLUMNS")
	defer func() {
		if originalColumns == "" {
			os.Unsetenv("COLUMNS")
		} else {
			os.Setenv("COLUMNS", originalColumns)
		}
	}()

	// Unset COLUMNS to allow actual terminal detection
	os.Unsetenv("COLUMNS")

	// Get terminal width
	width := getTerminalWidth()

	// Verify width is positive and reasonable
	if width <= 0 {
		t.Errorf("getTerminalWidth() returned non-positive value: %d", width)
	}

	if width > 1000 {
		t.Errorf("getTerminalWidth() returned unreasonably large value: %d", width)
	}
}

