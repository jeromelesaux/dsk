package cpc

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
)

// CpcHead structure describes the Amsdos header
type CpcHead struct {
	User        byte
	Filename    [15]byte
	BlockNum    byte
	LastBlock   byte
	Type        byte
	Size        uint16
	Address     uint16
	FirstBlock  byte
	LogicalSize uint16
	Exec        uint16
	NotUsed     [0x24]byte
	Size2       uint16
	BigLength   byte
	Checksum    uint16
	NotUsed2    [0x3B]byte
}

func NewCpcHeader(f *os.File) (*CpcHead, error) {
	header := &CpcHead{}
	data := make([]byte, 128)
	_, err := f.Read(data)
	if err != nil {
		return &CpcHead{}, err
	}
	buf := bytes.NewBuffer(data)
	err = binary.Read(buf, binary.LittleEndian, header)
	if err != nil {
		return &CpcHead{}, err
	}

	return header, nil
}

func (c *CpcHead) ComputedChecksum16() uint16 {
	var checksum uint16
	bf := make([]byte, 2)
	checksum += uint16(c.User)
	checksum += Checksum16(c.Filename[:])
	checksum += uint16(c.BlockNum)
	checksum += uint16(c.LastBlock)
	checksum += uint16(c.Type)
	binary.LittleEndian.PutUint16(bf, uint16(c.Size))
	checksum += Checksum16(bf)
	binary.LittleEndian.PutUint16(bf, uint16(c.Address))
	checksum += Checksum16(bf)
	checksum += uint16(c.FirstBlock)
	binary.LittleEndian.PutUint16(bf, uint16(c.LogicalSize))
	checksum += Checksum16(bf)
	binary.LittleEndian.PutUint16(bf, uint16(c.Exec))
	checksum += Checksum16(bf)
	checksum += Checksum16(c.NotUsed[:])
	binary.LittleEndian.PutUint16(bf, uint16(c.Size2))
	checksum += Checksum16(bf)
	checksum += uint16(c.BigLength)

	return checksum
}

// ToString Will dislay the CpcHead structure content
func (c *CpcHead) ToString() string {
	return fmt.Sprintf("User:%x\nFilename:%s\nType:%d\nSize:&%.2x\nAddress of loading:&%.2x\nAddress of execution:&%.2x\nChecksum:&%.2x\nComputed Checksum:&%.2x\n",
		int(c.User),
		string(c.Filename[:]),
		c.Type,
		c.Size2,
		c.Address,
		c.Exec,
		c.Checksum,
		c.ComputedChecksum16())
}

func (c *CpcHead) PrettyPrint() {
	fmt.Printf("%x", *c)
	return
}

var (
	xorstream1 = []byte{0xE2, 0x9D, 0xDB, 0x1A, 0x42, 0x29, 0x39, 0xC6, 0xB3, 0xC6, 0x90, 0x45, 0x8A}
	xorstream2 = []byte{0x49, 0xB1, 0x36, 0xF0, 0x2E, 0x1E, 0x06, 0x2A, 0x28, 0x19, 0xEA}
)

// DecryptHash returns value from decryptage of the data
func DecryptHash(data []byte) int {
	size := len(data)
	idx1 := 0
	idx2 := 0
	i := 0
	j := 0
	for j < size {
		if i == 0x80 {
			idx1 = 0
			idx2 = 0
			i = 0
		}
		data[j] ^= xorstream1[idx1]
		idx1++
		data[j] ^= xorstream2[idx2]
		idx2++
		if idx1 == 13 {
			idx1 = 0
		}
		if idx2 == 11 {
			idx2 = 0
		}
		i++
		j++
	}
	return 0
}

// Checksum16 will generate the checksum of the data amsdos header
func Checksum16(data []byte) uint16 {
	var checksum uint16
	for i := 0; i < len(data); i++ {
		checksum += uint16(data[i])
	}
	return checksum
}
