package dsk

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetNomDir(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantNom  string
		wantExt  string
	}{
		{"simple file", "test.bas", "TEST    ", "BAS"},
		{"long name", "longfilename.txt", "LONGFILE", "TXT"},
		{"short name", "a.b", "A       ", "B  "},
		{"no extension", "noext", "NOEXT   ", "EXT"},
		{"path with dir", "/some/path/file.scr", "FILE    ", "SCR"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := GetNomDir(tt.input)
			assert.Equal(t, tt.wantNom, string(entry.Nom[:]))
			assert.Equal(t, tt.wantExt, string(entry.Ext[:]))
		})
	}
}

func TestDisplayHex(t *testing.T) {
	t.Run("basic output", func(t *testing.T) {
		data := []byte{0x41, 0x42, 0x00, 0x7F}
		result := DisplayHex(data, 4)
		assert.Contains(t, result, "#0000")
		assert.Contains(t, result, "41 42 00 7F")
		assert.Contains(t, result, "AB")
	})

	t.Run("empty", func(t *testing.T) {
		result := DisplayHex([]byte{}, 8)
		assert.Empty(t, result)
	})

	t.Run("multi line", func(t *testing.T) {
		data := make([]byte, 32)
		for i := range data {
			data[i] = byte(i + 0x41)
		}
		result := DisplayHex(data, 16)
		assert.Contains(t, result, "#0000")
		assert.Contains(t, result, "#0010")
	})
}

func TestToAscii(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected string
	}{
		{"simple", []byte("HELLO"), "HELLO"},
		{"with high bit", []byte{0xC8, 0xC5, 0xCC, 0xCC, 0xCF}, "HELLO"},
		{"with spaces", []byte{0x20, 0x20, 0x41}, "A"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToAscii(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
