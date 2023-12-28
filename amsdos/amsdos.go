package amsdos

import (
	"bytes"
	"encoding/binary"

	"github.com/jeromelesaux/m4client/cpc"
)

type StAmsdos = cpc.CpcHead

func CheckAmsdos(buf []byte) (bool, *StAmsdos) {
	header := &StAmsdos{}
	rbuff := bytes.NewReader(buf)
	if err := binary.Read(rbuff, binary.LittleEndian, header); err != nil {
		return false, &StAmsdos{}
	}
	if header.Checksum == header.ComputedChecksum16() {
		return true, header
	}
	return false, &StAmsdos{}
}
