package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseHex16(t *testing.T) {
	tests := []struct {
		input    string
		expected uint16
		hasError bool
	}{
		{"#C000", 0xC000, false},
		{"0xC000", 0xC000, false},
		{"49152", 0xC000, false},
		{"#D0", 0xD0, false},
		{"208", 0xD0, false},
		{"0x170", 0x170, false},
		{"#170", 0x170, false},
		{"368", 0x170, false},
		{"invalid", 0, true},
		{"#ZZ", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := ParseHex16(tt.input)
			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}