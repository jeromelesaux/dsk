package hfe

import (
	"encoding/binary"
	"errors"
	"io"
)

var (
	NoHFEFormatFound       = errors.New("no hfe format found")
	HFERevisionUnsupported = errors.New("hfe revision format no supported")
)

type PicFileFormatHeader struct {
	HeaderSignature     [8]uint8
	FormatRevision      uint8
	NbTracks            uint8
	NbSide              uint8
	TrackEncoding       uint8
	BitRate             uint16
	FloppyRPM           uint16
	FloppyInterfaceMode FloppyInterfaceMode
	DNU                 uint8
	TrackListOffset     uint16
	WriteAllowed        uint8
	SingleStep          uint8
	Track0s0Altencoding TrackEncoding
	Track0s0Encoding    TrackEncoding
	Track0s1Altencoding TrackEncoding
	Track0s1Encoding    TrackEncoding
}

const Signature = "HXCPICFE"

type FloppyInterfaceMode = uint8

const (
	IBMPC_DD_FLOPPYMODE            FloppyInterfaceMode = 0x0
	IBMPC_HD_FLOPPYMODE            FloppyInterfaceMode = 0x01
	ATARIST_DD_FLOPPYMODE          FloppyInterfaceMode = 0x02
	ATARIST_HD_FLOPPYMODE          FloppyInterfaceMode = 0x03
	AMIGA_DD_FLOPPYMODE            FloppyInterfaceMode = 0x04
	AMIGA_HD_FLOPPYMODE            FloppyInterfaceMode = 0x05
	CPC_DD_FLOPPYMODE              FloppyInterfaceMode = 0x06
	GENERIC_SHUGGART_DD_FLOPPYMODE FloppyInterfaceMode = 0x07
	IBMPC_ED_FLOPPYMODE            FloppyInterfaceMode = 0x08
	MSX2_DD_FLOPPYMODE             FloppyInterfaceMode = 0x09
	C64_DD_FLOPPYMODE              FloppyInterfaceMode = 0x0A
	EMU_SHUGART_FLOPPYMODE         FloppyInterfaceMode = 0x0B
	S950_DD_FLOPPYMODE             FloppyInterfaceMode = 0x0C
	S950_HD_FLOPPYMODE             FloppyInterfaceMode = 0x0D
	DISABLE_FLOPPYMODE             FloppyInterfaceMode = 0xFE
)

var floppyInterfaceModes map[FloppyInterfaceMode]string = map[FloppyInterfaceMode]string{
	IBMPC_DD_FLOPPYMODE:            "IBMPC_DD_FLOPPYMODE",
	IBMPC_HD_FLOPPYMODE:            "IBMPC_HD_FLOPPYMODE",
	ATARIST_DD_FLOPPYMODE:          "ATARIST_DD_FLOPPYMODE",
	ATARIST_HD_FLOPPYMODE:          "ATARIST_HD_FLOPPYMODE",
	CPC_DD_FLOPPYMODE:              "CPC_DD_FLOPPYMODE",
	GENERIC_SHUGGART_DD_FLOPPYMODE: "GENERIC_SHUGGART_DD_FLOPPYMODE",
	IBMPC_ED_FLOPPYMODE:            "IBMPC_ED_FLOPPYMODE",
	MSX2_DD_FLOPPYMODE:             "MSX2_DD_FLOPPYMODE",
	C64_DD_FLOPPYMODE:              "C64_DD_FLOPPYMODE",
	EMU_SHUGART_FLOPPYMODE:         "EMU_SHUGART_FLOPPYMODE",
	S950_DD_FLOPPYMODE:             "S950_DD_FLOPPYMODE",
	S950_HD_FLOPPYMODE:             "S950_HD_FLOPPYMODE",
	DISABLE_FLOPPYMODE:             "DISABLE_FLOPPYMODE",
}

func String(f FloppyInterfaceMode) string {
	return floppyInterfaceModes[f]
}

type TrackEncoding = uint8

const (
	ISOIBM_MFM_ENCODING TrackEncoding = 0x00
	AMIGA_MFM_ENCODING  TrackEncoding = 0x01
	ISOIBM_FM_ENCODING  TrackEncoding = 0x02
	EMU_FM_ENCODING     TrackEncoding = 0x03
	UNKNOWN_ENCODING    TrackEncoding = 0xFF
)

type PicTrack struct {
	Offset   uint16
	TrackLen uint16
}

type HFE struct {
	Header PicFileFormatHeader
	Tracks []PicTrack
	Size   uint
}

type uFloppyBit struct {
	data    byte
	weak    byte
	strong  byte
	mark    byte
	index   byte
	garbage [3]byte
	content byte
}

type FloppyBit struct {
	cell     *uFloppyBit
	position uint16
	next     *FloppyBit
}

func ReadHeader(r io.Reader) (PicFileFormatHeader, error) {
	h := PicFileFormatHeader{}
	err := binary.Read(r, binary.LittleEndian, &h)
	if err != nil {
		return PicFileFormatHeader{}, err
	}

	if string(h.HeaderSignature[:]) != Signature {
		return PicFileFormatHeader{}, NoHFEFormatFound
	}

	if h.FormatRevision > 3 {
		return PicFileFormatHeader{}, HFERevisionUnsupported
	}

	if h.NbSide < 1 || h.NbSide > 2 || h.NbTracks == 0 || h.BitRate == 0 {
		return PicFileFormatHeader{}, HFERevisionUnsupported
	}
	return h, nil
}

func Read(r io.Reader) (HFE, error) {
	h := HFE{}
	var err error
	h.Header, err = ReadHeader(r)
	if err != nil {
		return HFE{}, err
	}
	h.Tracks = make([]PicTrack, h.Header.NbTracks)
	for i := 0; i < int(h.Header.NbTracks); i++ {
		err := binary.Read(r, binary.LittleEndian, &h.Tracks[i])
		if err != nil {
			return HFE{}, err
		}
	}
	return h, nil
}
