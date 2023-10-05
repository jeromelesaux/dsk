package hfe

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
