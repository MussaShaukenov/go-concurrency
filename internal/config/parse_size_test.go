package config

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseSize(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedSize  int
		expectedError bool
		errorContains string
	}{
		// Valid cases - bytes
		{
			name:         "Plain number",
			input:        "1024",
			expectedSize: 1024,
		},
		{
			name:         "Number with B suffix",
			input:        "1024B",
			expectedSize: 1024,
		},
		{
			name:         "Number with b suffix",
			input:        "1024b",
			expectedSize: 1024,
		},
		{
			name:         "Zero bytes",
			input:        "0",
			expectedSize: 0,
		},

		// Valid cases - KB
		{
			name:         "KB uppercase",
			input:        "4KB",
			expectedSize: 4 * 1024, // 4096
		},
		{
			name:         "Kb mixed case",
			input:        "4Kb",
			expectedSize: 4 * 1024,
		},
		{
			name:         "kb lowercase",
			input:        "4kb",
			expectedSize: 4 * 1024,
		},
		{
			name:         "1KB",
			input:        "1KB",
			expectedSize: 1024,
		},

		// Valid cases - MB
		{
			name:         "MB uppercase",
			input:        "16MB",
			expectedSize: 16 * 1024 * 1024, // 16777216
		},
		{
			name:         "Mb mixed case",
			input:        "16Mb",
			expectedSize: 16 * 1024 * 1024,
		},
		{
			name:         "mb lowercase",
			input:        "16mb",
			expectedSize: 16 * 1024 * 1024,
		},

		// Valid cases - GB
		{
			name:         "GB uppercase",
			input:        "2GB",
			expectedSize: 2 * 1024 * 1024 * 1024, // 2147483648
		},
		{
			name:         "Gb mixed case",
			input:        "2Gb",
			expectedSize: 2 * 1024 * 1024 * 1024,
		},
		{
			name:         "gb lowercase",
			input:        "2gb",
			expectedSize: 2 * 1024 * 1024 * 1024,
		},

		// Large numbers
		{
			name:         "Large number",
			input:        "999999",
			expectedSize: 999999,
		},
		{
			name:         "Large number with KB",
			input:        "999KB",
			expectedSize: 999 * 1024,
		},

		// Error cases - empty/invalid input
		{
			name:          "Empty string",
			input:         "",
			expectedError: true,
			errorContains: "incorrect size",
		},
		{
			name:          "Non-numeric start",
			input:         "abc",
			expectedError: true,
			errorContains: "incorrect size",
		},
		{
			name:          "Starts with letter",
			input:         "a123",
			expectedError: true,
			errorContains: "incorrect size",
		},
		{
			name:          "Starts with special char",
			input:         "-123",
			expectedError: true,
			errorContains: "incorrect size",
		},
		{
			name:          "Starts with space",
			input:         " 123",
			expectedError: true,
			errorContains: "incorrect size",
		},

		// Error cases - invalid suffixes
		{
			name:          "Invalid suffix",
			input:         "123TB",
			expectedError: true,
			errorContains: "incorrect size",
		},
		{
			name:          "Invalid suffix PB",
			input:         "123PB",
			expectedError: true,
			errorContains: "incorrect size",
		},
		{
			name:          "Random suffix",
			input:         "123XYZ",
			expectedError: true,
			errorContains: "incorrect size",
		},
		{
			name:          "Number with space",
			input:         "123 KB",
			expectedError: true,
			errorContains: "incorrect size",
		},
		{
			name:          "Mixed invalid suffix",
			input:         "123kB", // Note: kB is not supported, only KB/Kb/kb
			expectedError: true,
			errorContains: "incorrect size",
		},

		// Edge cases
		{
			name:          "Only suffix",
			input:         "KB",
			expectedError: true,
			errorContains: "incorrect size",
		},
		{
			name:          "Only suffix lowercase",
			input:         "mb",
			expectedError: true,
			errorContains: "incorrect size",
		},
		{
			name:         "Single digit",
			input:        "5",
			expectedSize: 5,
		},
		{
			name:         "Single digit with suffix",
			input:        "5MB",
			expectedSize: 5 * 1024 * 1024,
		},

		// Numbers with mixed content
		{
			name:          "Number with letters in middle",
			input:         "12a34",
			expectedError: true,
			errorContains: "incorrect size",
		},
		{
			name:          "Multiple suffixes",
			input:         "123KBMB",
			expectedError: true,
			errorContains: "incorrect size",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseSize(tt.input)

			if tt.expectedError {
				assert.Error(t, err, "Expected error for input: %s", tt.input)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				assert.Equal(t, 0, result, "Result should be 0 on error")
			} else {
				assert.NoError(t, err, "Unexpected error for input: %s", tt.input)
				assert.Equal(t, tt.expectedSize, result, "Incorrect size for input: %s", tt.input)
			}
		})
	}
}

func TestParseSize_Calculations(t *testing.T) {
	// Test specific calculations to ensure bit shifting is correct
	tests := []struct {
		input    string
		expected int
	}{
		{"1KB", 1 << 10}, // 1024
		{"2KB", 2 << 10}, // 2048
		{"1MB", 1 << 20}, // 1048576
		{"4MB", 4 << 20}, // 4194304
		{"1GB", 1 << 30}, // 1073741824
		{"3GB", 3 << 30}, // 3221225472
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("Calculation_%s", tt.input), func(t *testing.T) {
			result, err := ParseSize(tt.input)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseSize_EdgeCasesNumbers(t *testing.T) {
	// Test various number parsing edge cases
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"Leading zeros", "0001KB", 1024},
		{"Multiple zeros", "000", 0},
		{"Large number", "123456789", 123456789},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseSize(tt.input)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}
