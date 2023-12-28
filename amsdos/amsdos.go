package amsdos

import (
	"bytes"
	"encoding/binary"

	"github.com/jeromelesaux/m4client/cpc"
)

type AmsDosHeader = cpc.CpcHead

func CheckAmsdos(buf []byte) (bool, *AmsDosHeader) {
	header := &AmsDosHeader{}
	rbuff := bytes.NewReader(buf)
	if err := binary.Read(rbuff, binary.LittleEndian, header); err != nil {
		return false, &AmsDosHeader{}
	}
	if header.Checksum == header.ComputedChecksum16() {
		return true, header
	}
	return false, &AmsDosHeader{}
}
