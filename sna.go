package dsk

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
)

type SNA struct {
	Header SNAHeader
	Data   []byte
}

type SNAHeader struct {
	Identifier                      [8]uint8  // 0x0
	Unused                          [8]uint8  // 0x08
	Version                         uint8     // 0X10
	RegisterF                       uint8     // 0X11
	RegisterA                       uint8     // 0x12
	RegisterC                       uint8     // 0x13
	RegisterB                       uint8     // 0x14
	RegisterE                       uint8     // 0x15
	RegisterD                       uint8     // 0x16
	RegisterL                       uint8     // 0x17
	RegisterH                       uint8     // 0x18
	RegisterR                       uint8     // 0x19
	RegisterI                       uint8     // 0x1a
	InterruptIFF0                   uint8     // 0x1b
	InterruptIFF1                   uint8     // 0x1c
	RegisterIXLow                   uint8     // 0x1d
	RegisterIXHigh                  uint8     // 0x1e
	RegisterIYLow                   uint8     // 0x1f
	RegisterIYHigh                  uint8     // 0x20
	RegisterSPLow                   uint8     // 0x21
	RegisterSPHigh                  uint8     // 0x22
	RegisterPCLow                   uint8     // 0x23
	RegisterPCHigh                  uint8     // 0x24
	InterruptMode                   uint8     // 0x25
	RegisterF2                      uint8     // 0x26
	RegisterA2                      uint8     // 0x27
	RegisterC2                      uint8     // 0x28
	RegisterB2                      uint8     // 0x29
	RegisterE2                      uint8     // 0x2a
	RegisterD2                      uint8     // 0x2b
	RegisterL2                      uint8     // 0x2c
	RegisterH2                      uint8     // 0x2d
	GAIndex                         uint8     // 0x2e
	GAPalette                       [17]uint8 // 0x2f
	GAMultiConfiguration            uint8     // 0x40
	RAMConfiguration                uint8     // 0x41
	CRTCIndex                       uint8     // 0x42
	CRTCConfiguration               [18]uint8 // 0x43
	ROMSelection                    uint8     // 0x55
	PPIPortA                        uint8     // 0x56
	PPIPortB                        uint8     // 0x57
	PPIPortC                        uint8     // 0x58
	PPIControlPort                  uint8     // 0x59
	PSGIndexRegister                uint8     // 0x5a
	PSGRegisters                    [16]uint8 // 0x5b
	MemoryDumpSize                  uint8     // 0x6b 108 memory dump size in Kilobytes (e.g. 64 for 64K, 128 for 128k) (note 18)
	ExternalMemoryDumpSize          uint8
	CPCType                         uint8     // 0x6d  version 2 / 3
	InterruptNumber                 uint8     // 0x6e version 2 / 3
	MultimodeBytes                  [6]uint8  // 0x6f version 2 / 3
	Unused2                         [41]uint8 // 0x73
	FDDState                        [4]uint8  // 0x9d version 3
	FDDTrack                        uint8     // 0x9d
	PrinterRegister                 uint8
	CRTCType                        uint8 // 0xa4
	Unused3                         [4]uint8
	CRTCHorizontalCharacterRegister uint8 // version 3
	Unused5                         uint8
	CRTCCharacterLineRegister       uint8
	CRTCRasterRegister              uint8     // version 3
	CRTCVerticalRegister            uint8     // version 3
	CRTCHorizontalCounter           uint8     // version 3
	CRTCVerticalCounter             uint8     // version 3
	CRTCVsyncFlag                   uint8     // version 3
	CRTCHsyncFlag                   uint8     // version 3
	GADelayCounter                  uint8     // version 3
	GAInterruptScanlineCounter      uint8     // version 3
	InterruptFlag                   uint8     // version 3
	Unused4                         [75]uint8 // version 3
}

var (
	ErrorNoHeaderOrStartAddress = errors.New("No Amsdos header found and no startAddress. Quit.")
)

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
	if s.Header.Version == 3 {

	} else {
		s.Data = make([]byte, int(s.Header.MemoryDumpSize)*1000+int(s.Header.ExternalMemoryDumpSize)*1000)
	}
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

func (s *SNA) Put(content []byte, startAddress uint16) error {
	isAmsdos, header := CheckAmsdos(content)
	if isAmsdos && startAddress == 0 {
		copy(s.Data[header.Exec:], content[128:])
		return nil
	}
	if startAddress != 0 {
		if isAmsdos {
			copy(s.Data[startAddress:], content[128:])
		} else {
			copy(s.Data[startAddress:], content[:])
		}
		return nil
	}
	if startAddress == 0 && !isAmsdos {
		return ErrorNoHeaderOrStartAddress
	}
	return nil
}

func (s *SNA) Get(startAddress, lenght uint16) ([]byte, error) {
	content := make([]byte, lenght)
	if int(startAddress)+int(lenght) > len(s.Data) {
		return content, ErrorFileSizeExceed
	}
	copy(content[:], s.Data[startAddress:])
	return content, nil
}
