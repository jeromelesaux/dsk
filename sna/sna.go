package sna

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/jeromelesaux/dsk/amsdos"
	"github.com/jeromelesaux/dsk/dsk"
	m "github.com/jeromelesaux/m4client/cpc"
)

const SNATag = "DSK SID IMPACT"

type SNA struct {
	Header        SNAHeader
	Data          []byte
	CPCPlusChunck *CPCPlusChunck
	MemoryChuncks []*MemChunck
}

func NewSna(header SNAHeader) *SNA {
	s := &SNA{Header: header}
	s.Data = make([]byte, (int(s.Header.MemoryDumpSize)*1000)+(int(s.Header.ExternalMemoryDumpSize)*1000))
	return s
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
	PrinterRegister                 uint8     // 0xa3
	CRTCType                        uint8     // 0xa4
	Unused3                         [4]uint8
	CRTCHorizontalCharacterRegister uint8 // 0xa9 version 3
	Unused5                         uint8
	CRTCCharacterLineRegister       uint8     // 0xab
	CRTCRasterRegister              uint8     // 0xac version 3
	CRTCVerticalRegister            uint8     // 0xad version 3
	CRTCHorizontalCounter           uint8     // 0xae version 3
	CRTCVerticalCounter             uint8     // 0xaf version 3
	CRTCVsyncFlag                   uint8     // version 3
	CRTCHsyncFlag                   uint8     // version 3
	GADelayCounter                  uint8     // 0xb2 version 3
	GAInterruptScanlineCounter      uint8     // 0xb3 version 3
	InterruptFlag                   uint8     // 0xb4 version 3
	Unused4                         [75]uint8 // version 3 ended at 0xff
}

type CPCPlusChunck struct {
	ChunckLength                      [4]uint8
	SpritesBimaps                     [0x800]uint8
	SpritesAttributes                 [8 * 16]uint8
	Palette                           [32 * 2]uint8
	ProgrammableRasterInterrupt       uint8
	ScreenSplitScanLine               uint8
	ScreenSplitSecondaryScreenAddress [2]uint8
	SoftScrollControlRegister         uint8
	InterruptVector                   uint8
	Unused0                           [2]uint8
	AnalogueInputChannel              [8]uint8
	SoundDmaChannelAttributes         [3 * 4]uint8
	Unused1                           [3]uint8
	DMAControlStatus                  uint8
	DMAChannel02InternalRegisters     [3 * 7]uint8
	GateArrayA0RegisterValue          uint8
	GateArrayA0Lock                   uint8
	AsicUnlockSequenceState           uint8
}

type MemChunck struct {
	Data [0xffff]byte
}

func (m *MemChunck) Export() []byte {
	buf := make([]byte, 0)
	var index int
	for index < len(m.Data) {
		if m.Data[index] == m.Data[index+1] {
			var i int
			for m.Data[index] == m.Data[index+i] {
				i++
				if i == 0xff || i >= len(m.Data) {
					break
				}
			}
			buf = append(buf, 0xe5, byte(i), m.Data[index])
			index += i
		} else {
			buf = append(buf, m.Data[index])
			index++
		}
	}
	return buf
}

func (m *MemChunck) Feed(buf []byte) {
	var chunckIndex int
	var bufferIndex int
	for {
		if bufferIndex >= len(buf) {
			break
		}
		v := buf[bufferIndex]
		switch v {
		case 0xe5:
			occ := buf[bufferIndex+1]
			value := buf[bufferIndex+2]
			if occ == 0 {
				m.Data[chunckIndex] = v
				m.Data[chunckIndex+1] = occ
				m.Data[chunckIndex+2] = value
				bufferIndex += 3
				chunckIndex += 3
			} else {
				if value == 0 {
					chunckIndex += int(occ) - 1
					bufferIndex += 3
					continue
				} else {
					for i := 0; i < int(occ); i++ {
						m.Data[chunckIndex+i] = value
					}
					chunckIndex += int(occ) - 1
					bufferIndex += 3
				}
			}
		default:
			m.Data[chunckIndex] = v
			chunckIndex++
			bufferIndex++
		}
	}
}

func NewSnaHeader() SNAHeader {
	h := SNAHeader{
		Version:              1,
		MemoryDumpSize:       64,
		GAMultiConfiguration: 0x8D,
		CPCType:              uint8(CPC464),
		InterruptMode:        1,
		CRTCConfiguration:    [18]uint8{0x3f, 40, 46, 0x8e, 38, 0, 25, 30, 0, 7, 0, 0, 0x30},
		PSGRegisters:         [16]uint8{0, 0, 0, 0, 0, 0, 0x3f, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		PPIControlPort:       0x82,
		RegisterSPHigh:       0xC0,
		RegisterPCLow:        0xec,
		RegisterPCHigh:       0xbf,
		RegisterA:            0x2f,
		RegisterA2:           0x8d,
		RegisterB:            0xf5,
		RegisterB2:           0x7f,
		RegisterC:            0,
		RegisterC2:           0x8d,
		RegisterD:            0x3f,
		RegisterD2:           0x30,
		RegisterE:            0xe7,
		RegisterE2:           0,
		RegisterF:            0x3f,
		RegisterF2:           0,
		RegisterH:            0x3f,
		RegisterH2:           0xae,
		RegisterL:            0xe8,
		RegisterL2:           0x57,
		RegisterR:            0x44,
		RegisterI:            0x8,
		GAIndex:              0xf,
		GAPalette:            [17]uint8{0x04, 0x0A, 0x15, 0x1C, 0x18, 0x1D, 0x0C, 0x05, 0x0D, 0x16, 0x06, 0x17, 0x1E, 0x00, 0x1F, 0x0E, 0x04},
	}
	copy(h.Identifier[:], "MV - SNA")
	copy(h.Unused4[:], []byte(SNATag))
	return h
}

func NewSnaV2Header() SNAHeader {
	h := SNAHeader{
		Version:              2,
		MemoryDumpSize:       128,
		GAMultiConfiguration: 0x8D,
		CPCType:              uint8(UM6845R),
		CRTCType:             uint8(CPC464),
		InterruptMode:        1,
		CRTCConfiguration:    [18]uint8{0x3f, 40, 46, 0x8e, 38, 0, 25, 30, 0, 7, 0, 0, 0x30},
		PSGRegisters:         [16]uint8{0, 0, 0, 0, 0, 0, 0x3f, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		PPIControlPort:       0x82,
		RegisterSPHigh:       0xC0,
		RegisterPCLow:        0xec,
		RegisterPCHigh:       0xbf,
		RegisterA:            0x2f,
		RegisterA2:           0x8d,
		RegisterB:            0xf5,
		RegisterB2:           0x7f,
		RegisterC:            0,
		RegisterC2:           0x8d,
		RegisterD:            0x3f,
		RegisterD2:           0x30,
		RegisterE:            0xe7,
		RegisterE2:           0,
		RegisterF:            0x3f,
		RegisterF2:           0,
		RegisterH:            0x3f,
		RegisterH2:           0xae,
		RegisterL:            0xe8,
		RegisterL2:           0x57,
		RegisterR:            0x44,
		RegisterI:            0x8,
		GAIndex:              0xf,
		GAPalette:            [17]uint8{0x04, 0x0A, 0x15, 0x1C, 0x18, 0x1D, 0x0C, 0x05, 0x0D, 0x16, 0x06, 0x17, 0x1E, 0x00, 0x1F, 0x0E, 0x04},
	}
	copy(h.Identifier[:], "MV - SNA")
	copy(h.Unused4[:], []byte(SNATag))
	return h
}

func NewSnaV3Header() SNAHeader {
	h := SNAHeader{
		Version:              3,
		MemoryDumpSize:       128,
		GAMultiConfiguration: 0x8D,
		CPCType:              uint8(UM6845R),
		CRTCType:             uint8(CPC464),
		InterruptMode:        1,
		CRTCConfiguration:    [18]uint8{0x3f, 40, 46, 0x8e, 38, 0, 25, 30, 0, 7, 0, 0, 0x30},
		PSGRegisters:         [16]uint8{0, 0, 0, 0, 0, 0, 0x3f, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		PPIControlPort:       0x82,
		RegisterSPHigh:       0xC0,
		RegisterPCLow:        0xec,
		RegisterPCHigh:       0xbf,
		RegisterA:            0x2f,
		RegisterA2:           0x8d,
		RegisterB:            0xf5,
		RegisterB2:           0x7f,
		RegisterC:            0,
		RegisterC2:           0x8d,
		RegisterD:            0x3f,
		RegisterD2:           0x30,
		RegisterE:            0xe7,
		RegisterE2:           0,
		RegisterF:            0x3f,
		RegisterF2:           0,
		RegisterH:            0x3f,
		RegisterH2:           0xae,
		RegisterL:            0xe8,
		RegisterL2:           0x57,
		RegisterR:            0x44,
		RegisterI:            0x8,
		GAIndex:              0xf,
		GAPalette:            [17]uint8{0x04, 0x0A, 0x15, 0x1C, 0x18, 0x1D, 0x0C, 0x05, 0x0D, 0x16, 0x06, 0x17, 0x1E, 0x00, 0x1F, 0x0E, 0x04},
	}
	copy(h.Identifier[:], "MV - SNA")
	copy(h.Unused4[:], []byte(SNATag))
	return h
}

func (s *SNAHeader) String() string {
	out := fmt.Sprintf("\tIdentifier:%s\n\tVersion:%d\n", s.Identifier, s.Version)
	out += fmt.Sprintf("\tRegisterA:#%x\n\tRegisterA2:#%x\n", s.RegisterA, s.RegisterA2)
	out += fmt.Sprintf("\tRegisterB:#%x\n\tRegisterB2:#%x\n", s.RegisterB, s.RegisterB2)
	out += fmt.Sprintf("\tRegisterC:#%x\n\tRegisterC2:#%x\n", s.RegisterC, s.RegisterC2)
	out += fmt.Sprintf("\tRegisterD:#%x\n\tRegisterD2:#%x\n", s.RegisterD, s.RegisterD2)
	out += fmt.Sprintf("\tRegisterE:#%x\n\tRegisterE2:#%x\n", s.RegisterE, s.RegisterE2)
	out += fmt.Sprintf("\tRegisterH:#%x\n\tRegisterH2:#%x\n", s.RegisterH, s.RegisterH2)
	out += fmt.Sprintf("\tRegisterL:#%x\n\tRegisterL2:#%x\n", s.RegisterL, s.RegisterL2)
	out += fmt.Sprintf("\tRegisterR:#%x\n\tRegisterI:#%x\n", s.RegisterR, s.RegisterI)
	out += fmt.Sprintf("\tInterruptIFF0:#%x\n\tInterruptIFF1:#%x\n", s.InterruptIFF0, s.InterruptIFF1)
	out += fmt.Sprintf("\tRegisterIXLow:#%x\n\tRegisterIXHigh:#%x\n", s.RegisterIXLow, s.RegisterIXHigh)
	out += fmt.Sprintf("\tRegisterIYLow:#%x\n\tRegisterIYHigh:#%x\n", s.RegisterIYLow, s.RegisterIYHigh)
	out += fmt.Sprintf("\tRegisterSPLow:#%x\n\tRegisterSPHigh:#%x\n", s.RegisterSPLow, s.RegisterSPHigh)
	out += fmt.Sprintf("\tRegisterPCLow:#%x\n\tRegisterPCHigh:#%x\n", s.RegisterPCLow, s.RegisterPCHigh)
	out += fmt.Sprintf("\tInterruptionMode:#%x\n", s.InterruptMode)
	out += fmt.Sprintf("\tGAIndex:#%x\n\tGADelayCounter:#%x\n", s.GAIndex, s.GADelayCounter)
	out += fmt.Sprintf("\tGAInterruptScanlineCounter:#%x\n\tGAMultiConfiguration:#%x\n", s.GAInterruptScanlineCounter, s.GAMultiConfiguration)
	out += fmt.Sprintf("\tGAPalette:#%x\n", s.GAPalette)
	return out
}

var ErrorNoHeaderOrStartAddress = errors.New("no Amsdos header found and no startAddress. Quit")

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

func CRTCValue(crtc CRTC) uint8 {
	switch crtc {
	case HD6845S_UM6845:
		return 0
	case UM6845R:
		return 1
	case MC6845:
		return 2
	case ASIC_6845:
		return 3
	case Pre_ASIC:
		return 4
	}
	return 0
}

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

func CPCValue(cpc CPC) uint8 {
	switch cpc {
	case CPC464:
		return 0
	case CPC664:
		return 1
	case CPC6128:
		return 2
	case Unknown:
		return 3
	case CPCPlus6128:
		return 4
	case CPCPlus464:
		return 5
	case GX4000:
		return 6
	}
	return 0
}

func CPCType(cpc int) CPC {
	switch cpc {
	case 0:
		return CPC464
	case 1:
		return CPC664
	case 2:
		return CPC6128
	case 3:
		return Unknown
	case 4:
		return CPCPlus6128
	case 5:
		return CPCPlus464
	case 6:
		return GX4000
	default:
		return Unknown
	}
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
	if s.Header.Version == 3 && s.Header.MemoryDumpSize == 0 {
		var buf [4]byte

		for {
			if err := binary.Read(r, binary.LittleEndian, &buf); err != nil {
				if errors.Is(err, io.EOF) {
					break
				} else {
					fmt.Fprintf(os.Stderr, "Cannot read SNA header error :%v\n", err)
					return err
				}
			}
			if string(buf[:]) == "CPC+" {
				s.CPCPlusChunck = &CPCPlusChunck{}
				if err := binary.Read(r, binary.LittleEndian, s.CPCPlusChunck); err != nil {
					fmt.Fprintf(os.Stderr, "Cannot read SNA CPC Chunck error :%v\n", err)
					return err
				}
				continue
			}
			var nbBank int
			for i := 0; i < nbBank; i++ {
				s.MemoryChuncks = append(s.MemoryChuncks, &MemChunck{})
			}
			var size uint32
			if err := binary.Read(r, binary.LittleEndian, &size); err != nil {
				fmt.Fprintf(os.Stderr, "Cannot read SNA Memory Chunck error :%v\n", err)
				return err
			}
			memBuf := make([]byte, size)
			if err := binary.Read(r, binary.LittleEndian, &memBuf); err != nil {
				fmt.Fprintf(os.Stderr, "Cannot read SNA Memory Chunck error :%v\n", err)
				return err
			}

			mem := &MemChunck{}
			mem.Feed(memBuf)

			s.MemoryChuncks = append(s.MemoryChuncks, mem)
			nbBank++
		}
	} else {
		s.Data = make([]byte, int(s.Header.MemoryDumpSize)*1000+int(s.Header.ExternalMemoryDumpSize)*1000)
		if err := binary.Read(r, binary.LittleEndian, &s.Data); err != nil {
			fmt.Fprintf(os.Stderr, "Cannot read SNA data error :%v\n", err)
			return err
		}
	}

	return nil
}

func (s *SNA) Write(w io.Writer) error {
	if err := binary.Write(w, binary.LittleEndian, &s.Header); err != nil {
		fmt.Fprintf(os.Stderr, "Cannot write SNA header error :%v\n", err)
		return err
	}
	if s.Header.Version == 3 {
		if s.CPCPlusChunck != nil {
			name := []byte("CPC+")
			if err := binary.Write(w, binary.LittleEndian, &name); err != nil {
				fmt.Fprintf(os.Stderr, "Cannot write CPCPlusChunck name SNA data error :%v\n", err)
				return err
			}
			if err := binary.Write(w, binary.LittleEndian, s.CPCPlusChunck); err != nil {
				fmt.Fprintf(os.Stderr, "Cannot write CPCPlusChunck SNA data error :%v\n", err)
				return err
			}
		}
		for i := 0; i < len(s.MemoryChuncks); i++ {
			name := []byte("MEM" + fmt.Sprintf("%d", i))
			if err := binary.Write(w, binary.LittleEndian, &name); err != nil {
				fmt.Fprintf(os.Stderr, "Cannot write CPCPlusChunck name SNA data error :%v\n", err)
				return err
			}
			data := s.MemoryChuncks[i].Export()
			size := uint32(len(data))
			if err := binary.Write(w, binary.LittleEndian, &size); err != nil {
				fmt.Fprintf(os.Stderr, "Cannot write CPCPlusChunck SNA data error :%v\n", err)
				return err
			}
			if err := binary.Write(w, binary.LittleEndian, &data); err != nil {
				fmt.Fprintf(os.Stderr, "Cannot write CPCPlusChunck SNA data error :%v\n", err)
				return err
			}
		}
	} else {
		if err := binary.Write(w, binary.LittleEndian, &s.Data); err != nil {
			fmt.Fprintf(os.Stderr, "Cannot write SNA data error :%v\n", err)
			return err
		}
	}

	return nil
}

func (s *SNA) Put(content []byte, startAddress, length uint16) error {
	isAmsdos, header := amsdos.CheckAmsdos(content)
	if isAmsdos && startAddress == 0 {
		copy(s.Data[header.Exec:], content[128:])
		return nil
	}
	fmt.Fprintf(os.Stderr, "Copying into SNA start address #%4x is amsdos %v\n", startAddress, isAmsdos)
	if startAddress != 0 {
		if isAmsdos {
			copy(s.Data[startAddress:(len(content)-128)], content[128:])
		} else {
			copy(s.Data[startAddress:len(content)], content[:])
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
		return content, dsk.ErrorFileSizeExceed
	}
	copy(content[:], s.Data[startAddress:])
	return content, nil
}

func ExportFromSna(snaPath string) ([]byte, error) {
	f, err := os.Open(snaPath)
	if err != nil {
		return []byte{}, err
	}
	defer f.Close()
	s := &SNA{}
	if err = s.Read(f); err != nil {
		return []byte{}, err
	}
	var startingAddress uint16
	startingAddress += uint16(s.Header.RegisterPCHigh) << 8
	startingAddress += uint16(s.Header.RegisterPCLow) | 0xff
	_, err = s.Get(startingAddress, uint16(len(s.Data))-startingAddress)
	return s.Data, err
}

func ImportInSna(filePath, snaPath string, execAddress uint16, screenMode uint8, cpcType CPC, crtcType CRTC, version int) error {
	var sna *SNA
	switch version {
	case 1:
		sna = NewSna(NewSnaHeader())
	case 2:
		sna = NewSna(NewSnaV2Header())
	default:
		return dsk.ErrorUnsupportedDskFormat
	}

	var filesize uint16
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	header := &m.CpcHead{}
	if err := binary.Read(f, binary.LittleEndian, header); err != nil {
		return err
	}
	filesize = header.Size
	if header.Size == 0 {
		filesize = header.LogicalSize
	}
	//	f.Seek(0, 0)
	fmt.Fprintf(os.Stderr, "Import file %s at address:#%4x size:%4x\n", filePath, header.Address, filesize)
	buff := make([]byte, 0xFFFF)
	nb, err := f.Read(buff)
	if err != nil {
		return err
	}
	if err := sna.Put(buff[0:nb], header.Address, filesize); err != nil {
		return err
	}
	if execAddress == 0 {
		if header.Exec != 0 {
			execAddress = header.Exec
		}
	}

	sna.Header.RegisterPCHigh = uint8(execAddress >> 8)
	sna.Header.RegisterPCLow = uint8(execAddress & 0xff)
	// sna.Header.GAMultiConfiguration = 0x88
	sna.Header.CPCType = CPCValue(cpcType)
	sna.Header.CRTCType = CRTCValue(crtcType)

	switch screenMode {
	case 0:
		sna.Header.GAMultiConfiguration = 0x8c
	case 1:
		sna.Header.GAMultiConfiguration = 0x8d
	case 2:
		sna.Header.GAMultiConfiguration = 0x8e
	}
	w, err := os.Create(snaPath)
	if err != nil {
		return err
	}
	defer w.Close()
	if err := binary.Write(w, binary.LittleEndian, sna.Header); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, sna.Data); err != nil {
		return err
	}
	return nil
}

func CreateSna(snaPath string, snaVersion int) (*SNA, error) {
	var s *SNA
	switch snaVersion {
	case 1:
		s = NewSna(NewSnaHeader())
	case 2:
		s = NewSna(NewSnaV2Header())
	default:
		return nil, dsk.ErrorUnsupportedDskFormat
	}

	w, err := os.Create(snaPath)
	if err != nil {
		return s, err
	}
	defer w.Close()
	if err := binary.Write(w, binary.LittleEndian, s.Header); err != nil {
		return s, err
	}
	if err := binary.Write(w, binary.LittleEndian, s.Data); err != nil {
		return s, err
	}
	return s, nil
}

func (s *SNA) Hexadecimal() string {
	return dsk.DisplayHex(s.Data, 16)
}
