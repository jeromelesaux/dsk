package dsk

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

type SNA struct {
	Header SNAHeader
	Data   []byte
}

type SNAHeader struct {
	Identifier                 [8]uint8
	Unused                     [8]uint8
	Version                    uint8
	RegisterF                  uint8
	RegisterA                  uint8
	RegisterC                  uint8
	RegisterB                  uint8
	RegisterE                  uint8
	RegisterD                  uint8
	RegisterL                  uint8
	RegisterH                  uint8
	RegisterR                  uint8
	RegisterI                  uint8
	InterruptIFF0              uint8
	InterruptIFF1              uint8
	RegisterIXLow              uint8
	RegisterIXHigh             uint8
	RegisterIYLow              uint8
	RegisterIYHigh             uint8
	RegisterSPLow              uint8
	RegisterSPHigh             uint8
	RegisterPCLow              uint8
	RegisterPCHigh             uint8
	InterruptMode              uint8
	RegisterF2                 uint8
	RegisterA2                 uint8
	RegisterC2                 uint8
	RegisterB2                 uint8
	RegisterE2                 uint8
	RegisterD2                 uint8
	RegisterL2                 uint8
	RegisterH2                 uint8
	GAIndex                    uint8
	GAPalette                  [17]uint8
	GAMultiConfiguration       uint8
	RAMConfiguration           uint8
	CRTCIndex                  uint8
	CRTCConfiguration          [18]uint8
	ROMSelection               uint8
	PPIPortA                   uint8
	PPIPortB                   uint8
	PPIPortC                   uint8
	PPIControlPort             uint8
	PSGIndexRegister           uint8
	PSGRegisters               [16]uint8
	MemoryDumpSize             uint8    // memory dump size in Kilobytes (e.g. 64 for 64K, 128 for 128k) (note 18)
	CPCType                    uint8    // version 2 / 3
	InterruptNumber            uint8    // version 2 / 3
	MultimodeBytes             [6]uint8 // version 2 / 3
	Unused2                    [39]uint8
	FDDState                   uint8 // version 3
	FDDTrack                   uint8
	PrinterRegister            uint8
	CRTCType                   uint8
	Unused3                    [4]uint8
	CRTCCharacterRegister      uint8     // version 3
	CRTCRasterRegister         uint8     // version 3
	CRTCVerticalRegister       uint8     // version 3
	CRTCHorizontalCounter      uint8     // version 3
	CRTCVerticalCounter        uint8     // version 3
	CRTCVsyncFlag              uint8     // version 3
	CRTCHsyncFlag              uint8     // version 3
	GADelayCounter             uint8     // version 3
	GAInterruptScanlineCounter uint8     // version 3
	InterruptFlag              uint8     // version 3
	Unused4                    [75]uint8 // version 3
}

type CRTC uint8

const (
	HD6845S_UM6845 CRTC = iota
	UM6845R
	MC6845
	ASIC_6845
	Pre_ASIC
)

type CPC uint8

const (
	CPC464 CPC = iota
	CPC664
	CPC6128
	Unknown
	CPCPlus6128
	CPCPlus464
	GX4000
)

func (s *SNA) CRTCType() string {
	switch CRTC(s.Header.CRTCType) {
	case HD6845S_UM6845:
		return "HD6845S_UM6845"
	case UM6845R:
		return "UM6845R"
	case MC6845:
		return "MC6845"
	case ASIC_6845:
		return "ASIC_6845"
	case Pre_ASIC:
		return "Pre_ASIC"
	}
	return "Not defined"
}

func (s *SNA) CPCType() string {
	switch CPC(s.Header.CPCType) {
	case CPC464:
		return "CPC464"
	case CPC664:
		return "CPC664"
	case CPC6128:
		return "CPC6128"
	case Unknown:
		return "Unknown"
	case CPCPlus6128:
		return "CPCPlus6128"
	case CPCPlus464:
		return "CPCPlus464"
	case GX4000:
		return "GX4000"
	}
	return "No defined"
}

func ReadSna(filePath string) (*SNA, error) {
	f, err := os.Open(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot open file (%s) error %v\n", filePath, err)
		return &SNA{}, err
	}

	sna := &SNA{}
	if err := sna.Read(f); err != nil {
		fmt.Fprintf(os.Stderr, "Error while reading (%s) error %v", filePath, err)
	}
	f.Close()
	return sna, nil
}

func (s *SNA) Read(r io.Reader) error {
	if err := binary.Read(r, binary.LittleEndian, &s.Header); err != nil {
		fmt.Fprintf(os.Stderr, "Cannot read SNA header error :%v\n", err)
		return err
	}
	
	s.Data = make([]byte, int(s.Header.MemoryDumpSize)*1000)
	if err := binary.Read(r, binary.LittleEndian, &s.Data); err != nil {
		fmt.Fprintf(os.Stderr, "Cannot read SNA data error :%v\n", err)
		return err
	}

	return nil
}

func (s *SNA) Write(w io.Writer) error {
	if err := binary.Write(w, binary.LittleEndian, &s.Header); err != nil {
		fmt.Fprintf(os.Stderr, "Cannot write SNA header error :%v\n", err)
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, &s.Data); err != nil {
		fmt.Fprintf(os.Stderr, "Cannot write SNA data error :%v\n", err)
		return err
	}

	return nil
}
