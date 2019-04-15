package dsk

import (
	"strings"
)

func GetNomDir(s string) StDirEntry {
	entry := StDirEntry{}
	for i := 0; i < 8; i++ {
		entry.Nom[i] = 32
	}
	for i := 0; i < 3; i++ {
		entry.Ext[i] = 32
	}
	inputFile := strings.ToUpper(s)
	var index int
	for i := 0; i < len(inputFile) && index < 8; i++ {
		if inputFile[i] == '.' {
			break
		}
		entry.Nom[index] = inputFile[i]
		index++
	}

	index = 0
	for i := len(inputFile) - 3; i < len(inputFile) && index < 3; i++ {
		entry.Ext[index] = inputFile[i]
		index++
	}
	return entry
}

func DisplayHex(b []byte) string {
	var out string

	return out
}
