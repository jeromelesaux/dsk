package dsk

import (
	"fmt"
	"path"
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
	inputFile := strings.ToUpper(path.Base(s))
	var index int
	for i := 0; i < len(inputFile) && index < 8; i++ {
		if inputFile[i] == '.' {
			break
		}
		entry.Nom[index] = inputFile[i]
		index++
	}

	index = 0
	start := strings.Index(inputFile, ".")
	if start == -1 {
		start = len(inputFile) - 3
	} else {
		start++
	}
	for i := start; i < len(inputFile) && index < 3; i++ {
		entry.Ext[index] = inputFile[i]
		index++
	}
	return entry
}

func DisplayHex(b []byte, width int) string {
	var out string
	var ascii string
	var hexa string
	var offset int
	for i := 0; i < len(b); i += width {
		for j := 0; j < width; j++ {
			if j+i < len(b) {
				if b[j+i] > 32 && b[j+i] < 125 {
					ascii += fmt.Sprintf("%c", b[j+i])
				} else {
					ascii += "."
				}
				hexa += fmt.Sprintf("%.2X ", b[j+i])
			} else {
				ascii += " "
				hexa += "   "
			}
		}
		out += fmt.Sprintf("#%.4X ", offset) + hexa + " | " + ascii + " |\n"
		ascii = ""
		hexa = ""
		offset += width

	}
	return out
}
