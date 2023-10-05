package utils

import (
	"fmt"
	"math"
	"unicode"
)

var (
	Nbre      [11]string   = [11]string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10"}
	MotsClefs [0x80]string = [0x80]string{
		"AFTER", "AUTO", "BORDER", "CALL", "CAT", "CHAIN", "CLEAR", "CLG",
		"CLOSEIN", "CLOSEOUT", "CLS", "CONT", "DATA", "DEF", "DEFINT",
		"DEFREAL", "DEFSTR", "DEG", "DELETE", "DIM", "DRAW", "DRAWR", "EDIT",
		"ELSE", "END", "ENT", "ENV", "ERASE", "ERROR", "EVERY", "FOR",
		"GOSUB", "GOTO", "IF", "INK", "INPUT", "KEY", "LET", "LINE", "LIST",
		"LOAD", "LOCATE", "MEMORY", "MERGE", "MID$", "MODE", "MOVE", "MOVER",
		"NEXT", "NEW", "ON", "ON BREAK", "ON ERROR GOTO", "SQ", "OPENIN",
		"OPENOUT", "ORIGIN", "OUT", "PAPER", "PEN", "PLOT", "PLOTR", "POKE",
		"PRINT", "'", "RAD", "RANDOMIZE", "READ", "RELEASE", "REM", "RENUM",
		"RESTORE", "RESUME", "RETURN", "RUN", "SAVE", "SOUND", "SPEED", "STOP",
		"SYMBOL", "TAG", "TAGOFF", "TROFF", "TRON", "WAIT", "WEND", "WHILE",
		"WIDTH", "WINDOW", "WRITE", "ZONE", "DI", "EI", "FILL", "GRAPHICS",
		"MASK", "FRAME", "CURSOR", "#E2", "ERL", "FN", "SPC", "STEP", "SWAP",
		"#E8", "#E9", "TAB", "THEN", "TO", "USING", ">", "=", ">=", "<", "<>",
		"<=", "+", "-", "*", "/", "^", "\\ ", "AND", "MOD", "OR", "XOR", "NOT",
		"#FF",
	}
)
var Fcts [0x80]string = [0x80]string{
	"ABS", "ASC", "ATN", "CHR$", "CINT", "COS", "CREAL", "EXP", "FIX",
	"FRE", "INKEY", "INP", "INT", "JOY", "LEN", "LOG", "LOG10", "LOWER$",
	"PEEK", "REMAIN", "SGN", "SIN", "SPACE$", "SQ", "SQR", "STR$", "TAN",
	"UNT", "UPPER$", "VAL", "", "", "", "", "", "", "", "", "", "", "",
	"", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "",
	"", "", "", "", "", "", "EOF", "ERR", "HIMEM", "INKEY$", "PI", "RND",
	"TIME", "XPOS", "YPOS", "DERR", "", "", "", "", "", "", "", "", "", "",
	"", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "",
	"", "", "", "", "", "", "", "", "", "", "", "BIN$", "DEC$", "HEX$",
	"INSTR", "LEFT$", "MAX", "MIN", "POS", "RIGHT$", "ROUND", "STRING$",
	"TEST", "TESTR", "COPYCHR$", "VPOS",
}

var DproBasic [128]byte = [128]byte{
	0xAB, 0x2C, 0xED, 0xEA, 0x6C, 0x37, 0x3F, 0xEC,
	0x9B, 0xDF, 0x7A, 0x0C, 0x3B, 0xD4, 0x6D, 0xF5,
	0x04, 0x44, 0x03, 0x11, 0xDF, 0x59, 0x8F, 0x21,
	0x73, 0x7A, 0xCC, 0x83, 0xDD, 0x30, 0x6A, 0x30,
	0xD3, 0x8F, 0x02, 0xF0, 0x60, 0x6B, 0x94, 0xE4,
	0xB7, 0xF3, 0x03, 0xA8, 0x60, 0x88, 0xF0, 0x43,
	0xE8, 0x8E, 0x43, 0xA0, 0xCA, 0x84, 0x31, 0x53,
	0xF3, 0x1F, 0xC9, 0xE8, 0xAD, 0xC0, 0xBA, 0x6D,
	0x93, 0x08, 0xD4, 0x6A, 0x2C, 0xB2, 0x07, 0x27,
	0xC0, 0x99, 0xEE, 0x89, 0xAF, 0xC3, 0x53, 0xAB,
	0x2B, 0x34, 0x5C, 0x2F, 0x13, 0xEE, 0xAA, 0x2C,
	0xD9, 0xF4, 0xBC, 0x12, 0xB3, 0xC5, 0x1C, 0x68,
	0x01, 0x20, 0x2C, 0xFA, 0x77, 0xA6, 0xB5, 0xA4,
	0xFC, 0x9B, 0xF1, 0x32, 0x5B, 0xC3, 0x70, 0x77,
	0x85, 0x36, 0xBE, 0x5B, 0x8C, 0xC8, 0xB5, 0xC2,
	0xF0, 0x0B, 0x98, 0x0F, 0x36, 0x9D, 0xD8, 0x96,
}

func getByte(buf []byte, pos uint16, deprotect uint8) byte {
	return buf[pos] ^ (DproBasic[pos&0x7F] * deprotect)
}

func getWord(buf []byte, pos uint16, deprotect uint8) int {
	ret := int(buf[pos] ^ (DproBasic[pos&0x7F] * deprotect))
	pos++
	ret += int(int(buf[pos])^int((DproBasic[pos&0x7F]*deprotect))) << 8
	return ret
}

func addWord(buf []byte, pos uint16, listing []byte, deprotect uint8) ([]byte, uint16) {
	var lenVar int
	for {
		b := getByte(buf, pos, deprotect)
		pos++
		listing = append(listing, (b & 0x7f))
		if b&0x80 != 0 || lenVar >= 0xff {
			break
		}
		lenVar++
	}
	return listing, pos
}

func Basic(buf []byte, fileSize uint16, isBasic bool) []byte {
	var token, deprotect uint8
	var pos uint16
	listing := make([]byte, 0)
	token = getByte(buf, 0, deprotect)
	for {
		if isBasic {
			lg := getWord(buf, pos, deprotect)
			pos += 2
			if lg == 0 {
				break
			}
			numLigne := getWord(buf, pos, deprotect)
			pos += 2
			tmp := fmt.Sprintf("%d ", numLigne)
			listing = append(listing, tmp...)
		} else {
			if token != 0 || token == 0x1a {
				break
			}
		}
		dansChaine := 0
		for {
			token = getByte(buf, pos, deprotect)
			pos++
			if !isBasic && token == 0x1a {
				break
			}
			if dansChaine == 1 || !isBasic {
				listing = append(listing, token)
				if token == '"' {
					dansChaine ^= 1
				}
			} else {
				if token > 0x7F && token < 0xFF {
					if listing[len(listing)-1] == ':' && token == 0x97 {
						listing[len(listing)-1] = 0
					}
					listing = append(listing, MotsClefs[token&0x7f]...)

				} else {
					if token >= 0x0E && token <= 0x18 {
						listing = append(listing, Nbre[token-0x0e]...)
					} else {
						if token >= 0x20 && token < 0x7C {

							listing = append(listing, token)
							if token == '"' {
								dansChaine ^= 1
							}
						} else {
							tmp := make([]byte, 2)
							switch token {
							case 0x01:
								listing = append(listing, ':')
							case 0x02: // Variable entière (type %)
								listing, pos = addWord(buf, 2+pos, listing, deprotect)
								listing = append(listing, '%')
							case 0x03: // Variable chaine (type $)
								listing, pos = addWord(buf, 2+pos, listing, deprotect)
								listing = append(listing, '$')
							case 0x04: // Variable float (type !)
								listing, pos = addWord(buf, 2+pos, listing, deprotect)
								listing = append(listing, '!')
							case 0x0B:
							case 0x0C:
							case 0x0D: // Variable "standard"
								listing, pos = addWord(buf, 2+pos, listing, deprotect)
							case 0x19: // Constante entière 8 bits
								val := fmt.Sprintf("%d", getByte(buf, pos, deprotect))
								listing = append(listing, val...)
								pos++
							case 0x1A:
							case 0x1E: // Constante entière 16 bits
								w := getWord(buf, pos, deprotect)
								val := fmt.Sprintf("%d", w)
								listing = append(listing, val...)
								pos += 2
							case 0x1B:
								w := getWord(buf, pos, deprotect)
								val := fmt.Sprintf("&X%X", w)
								listing = append(listing, val...)
								pos += 2
							case 0x1C:
								w := getWord(buf, pos, deprotect)
								val := fmt.Sprintf("&%X", w)
								listing = append(listing, val...)
								pos += 2

							case 0x1F: // Constante flottante
								f := float64((int(getByte(buf, pos+2, deprotect)) << 16) +
									(int(getByte(buf, pos+1, deprotect)) << 8) +
									int(getByte(buf, pos, deprotect)) +
									((int(getByte(buf, pos+3, deprotect) & 0x7F)) << 24))
								f = 1 + (f / 0x80000000)

								if getByte(buf, pos+3, deprotect)&0x80 == 0 {
									f = -f
								}

								exp := getByte(buf, pos+4, deprotect) - 129
								pos += 5
								val := fmt.Sprintf("%f", f*math.Pow(float64(2), float64(exp)))
								// Suppression des '0' inutiles
								listing = append(listing, val...)

							case 0x7C:
								listing = append(listing, '|')
								listing, pos = addWord(buf, 1+pos, listing, deprotect)
							case 0xFF:
								if getByte(buf, pos, deprotect) < 0x80 {
									listing = append(listing, Fcts[getByte(buf, pos, deprotect)]...)
									pos++
								} else {
									tmp[1] = 0
									tmp[0] = getByte(buf, pos, deprotect) & 0x7F
									pos++
									listing = append(listing, tmp...)
								}
							default:

							}
						}
					}
				}
			}
			if token == 0 {
				break
			}
		}
		listing = append(listing, "\n"...)
		if pos >= fileSize {
			break
		}
	}
	// Conversion des caractères accentués si nécessaire
	for i := len(listing) - 1; i >= 0; i-- {
		if !unicode.IsPrint(rune(listing[i])) && listing[i] != '\n' && listing[i] != '\r' {
			listing[i] = '?'
		}
	}

	return listing
}
