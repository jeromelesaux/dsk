package utils

import (
	"strconv"
	"strings"
)

// ParseHex16 parses a hexadecimal string in various formats (#C000, 0xC000, 49152) and returns uint16
func ParseHex16(address string) (uint16, error) {
	var value16 uint16
	switch address[0] {
	case '#':
		value := strings.ReplaceAll(address, "#", "")
		v, err := strconv.ParseUint(value, 16, 16)
		if err != nil {
			return 0, err
		}
		value16 = uint16(v)
	case '0':
		if len(address) > 1 && address[1] == 'x' {
			value := strings.ReplaceAll(address, "0x", "")
			v, err := strconv.ParseUint(value, 16, 16)
			if err != nil {
				return 0, err
			}
			value16 = uint16(v)
		} else {
			v, err := strconv.ParseUint(address, 10, 16)
			if err != nil {
				return 0, err
			}
			value16 = uint16(v)
		}
	default:
		v, err := strconv.ParseUint(address, 10, 16)
		if err != nil {
			return 0, err
		}
		value16 = uint16(v)
	}
	return value16, nil
}
