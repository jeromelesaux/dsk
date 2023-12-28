package dsk

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/jeromelesaux/dsk/amsdos"
	"github.com/jeromelesaux/m4client/cpc"
)

const (
	USER_DELETED uint8  = 0xE5
	SECTSIZE     uint16 = 512
	NOT_FOUND    int    = -1
)

// 0xE5 in binary is 11100101, which on MFM encoding is 0101010101010101 (0x5555)
var (
	ErrorUnsupportedDskFormat = errors.New("unsupported DSK Format")
	ErrorBadSectorNumber      = errors.New("dsk has wrong sector number")
	ErrorCatalogueExceed      = errors.New("catalogue indices exceed")
	ErrDiskFull               = errors.New("error no more free blocks available")
	ErrorNoDirEntry           = errors.New("error no more dir entry available")
	ErrorFileSizeExceed       = errors.New("filesize exceed")
)

type SaveMode uint8

const (
	SaveModeAscii SaveMode = iota
	SaveModeProtected
	SaveModeBinary
)

type DskFileType int

const (
	DskTypeBasic DskFileType = iota
	DskTypeExtended
	DskTypeSNA
)

var (
	EXTENDED_DSK_TYPE           = 1
	DSK_TYPE                    = 0
	DataFormat        DskFormat = 0
	VendorFormat      DskFormat = 1
)

const HeaderSize = 0x80

type DskFormat = int

type CPCEMUEnt struct {
	Header   [0x22]byte // "MV - CPCEMU Disk-File\r\nDisk-Info\r\n"
	Creator  [0xE]byte
	Tracks   uint8  // number of tracks
	Heads    uint8  // number of heads (1 or 2)
	DataSize uint16 // 0x1300 = 256 + ( 512 * sectors )
}

func (e *CPCEMUEnt) ToString() string {
	return fmt.Sprintf("Header:%s\nCreator:%s\nnbTracks:%d, nbHeads:%d, DataSize:%d",
		e.Header, e.Creator, e.Tracks, e.Heads, e.DataSize)
}

type CPCEMUSect struct { // length 8
	C    uint8 // track,
	H    uint8 // head
	S    uint8 // sect
	N    uint8 // size
	Un1  uint16
	Size uint16 // Sector size in bytes
}

func (c *CPCEMUSect) ToString() string {
	return fmt.Sprintf("C:%d,H:%d,S:%d,N:%d,Un1:%d:Size:%d", //,DataSize:%d",
		c.C, c.H, c.S, c.N, c.Un1, c.Size) //, len(c.Data))
}

func (c *CPCEMUSect) Write(w io.Writer) error {
	if err := binary.Write(w, binary.LittleEndian, &c.C); err != nil {
		fmt.Fprintf(os.Stderr, "Error while writing CPCEmuSect.C error :%v\n", err)
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, &c.H); err != nil {
		fmt.Fprintf(os.Stderr, "Error while writing CPCEmuSect.H error :%v\n", err)
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, &c.S); err != nil {
		fmt.Fprintf(os.Stderr, "Error while writing CPCEmuSect.S error :%v\n", err)
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, &c.N); err != nil {
		fmt.Fprintf(os.Stderr, "Error while writing CPCEmuSect.N error :%v\n", err)
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, &c.Un1); err != nil {
		fmt.Fprintf(os.Stderr, "Error while writing CPCEmuSect.Un1 error :%v\n", err)
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, &c.Size); err != nil {
		fmt.Fprintf(os.Stderr, "Error while writing CPCEmuSect.Size error :%v\n", err)
		return err
	}
	//	fmt.Fprintf(os.Stdout,"Sector %s\n",c.ToString())
	return nil
}

func (c *CPCEMUSect) Read(r io.Reader) error {
	if err := binary.Read(r, binary.LittleEndian, &c.C); err != nil {
		fmt.Fprintf(os.Stderr, "Error while reading CPCEmuSect.C error :%v\n", err)
		return err
	}
	if err := binary.Read(r, binary.LittleEndian, &c.H); err != nil {
		fmt.Fprintf(os.Stderr, "Error while reading CPCEmuSect.H error :%v\n", err)
		return err
	}
	if err := binary.Read(r, binary.LittleEndian, &c.S); err != nil {
		fmt.Fprintf(os.Stderr, "Error while reading CPCEmuSect.S error :%v\n", err)
		return err
	}
	if err := binary.Read(r, binary.LittleEndian, &c.N); err != nil {
		fmt.Fprintf(os.Stderr, "Error while reading CPCEmuSect.N error :%v\n", err)
		return err
	}
	if err := binary.Read(r, binary.LittleEndian, &c.Un1); err != nil {
		fmt.Fprintf(os.Stderr, "Error while reading CPCEmuSect.Un1 error :%v\n", err)
		return err
	}
	if err := binary.Read(r, binary.LittleEndian, &c.Size); err != nil {
		fmt.Fprintf(os.Stderr, "Error while reading CPCEmuSect.Size error :%v\n", err)
		return err
	}
	//	fmt.Fprintf(os.Stdout,"Sector %s\n",c.ToString())
	return nil
}

type CPCEMUTrack struct { // length 18 bytes
	ID          [0x10]byte // "Track-Info\r\n"
	Track       uint8
	Head        uint8
	Unused      [2]byte        // kept here to keep 1:1 with the disk format
	SectSize    uint8          // 2
	SectorCount uint8          // 9
	Gap3        uint8          // 0x4E
	OctRemp     uint8          // 0xE5
	Sect        [29]CPCEMUSect // 29 sectors max
	Data        []byte
}

func (t *CPCEMUTrack) Read(r io.Reader) error {
	if err := binary.Read(r, binary.LittleEndian, &t.ID); err != nil {
		fmt.Fprintf(os.Stderr, "Error while reading CPCEMUTrack.ID error :%v\n", err)
		return err
	}
	if err := binary.Read(r, binary.LittleEndian, &t.Track); err != nil {
		fmt.Fprintf(os.Stderr, "Error while reading CPCEMUTrack.Track error :%v\n", err)
		return err
	}
	if err := binary.Read(r, binary.LittleEndian, &t.Head); err != nil {
		fmt.Fprintf(os.Stderr, "Error while reading CPCEMUTrack.Head error :%v\n", err)
		return err
	}
	if err := binary.Read(r, binary.LittleEndian, &t.Unused); err != nil {
		fmt.Fprintf(os.Stderr, "Error while reading CPCEMUTrack.Unused error :%v\n", err)
		return err
	}
	if err := binary.Read(r, binary.LittleEndian, &t.SectSize); err != nil {
		fmt.Fprintf(os.Stderr, "Error while reading CPCEMUTrack.SectSize error :%v\n", err)
		return err
	}
	if err := binary.Read(r, binary.LittleEndian, &t.SectorCount); err != nil {
		fmt.Fprintf(os.Stderr, "Error while reading CPCEMUTrack.SectorCount error :%v\n", err)
		return err
	}
	if err := binary.Read(r, binary.LittleEndian, &t.Gap3); err != nil {
		fmt.Fprintf(os.Stderr, "Error while reading CPCEMUTrack.Gap3 error :%v\n", err)
		return err
	}
	if err := binary.Read(r, binary.LittleEndian, &t.OctRemp); err != nil {
		fmt.Fprintf(os.Stderr, "Error while reading CPCEMUTrack.OctRemp error :%v\n", err)
		return err
	}

	//	fmt.Fprintf(os.Stdout,"Track:%s\n",c.ToString())
	var sectorSize uint16
	for i := uint8(0); i < t.SectorCount && i < 29; i++ {
		sect := &CPCEMUSect{}
		if err := sect.Read(r); err != nil {
			fmt.Fprintf(os.Stderr, "Cannot read sector (%d) error :%v\n", i, err)
			return err
		}
		t.Sect[i] = *sect
		sectorSize += t.Sect[i].Size
	}
	for i := t.SectorCount; i < 29; i++ {
		sect := &CPCEMUSect{}
		err := sect.Read(r)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error while reading sector (%d), error :%v\n", i, err)
		}
	}
	if int(sectorSize) > int(t.SectSize)*0x100*int(t.SectorCount) {
		fmt.Fprintf(os.Stderr, "Warning : Sector size [%d] differs from the amount of data found [%d], enlarge data part\n",
			int(t.SectSize)*0x100*int(t.SectorCount),
			sectorSize)
		t.Data = make([]byte, sectorSize)
	} else {
		t.Data = make([]byte, int(t.SectSize)*0x100*int(t.SectorCount))
	}
	if err := binary.Read(r, binary.LittleEndian, &t.Data); err != nil {
		fmt.Fprintf(os.Stderr, "Error while reading CPCEmuSect.Data error :%v\n", err)
		return err
	}
	return nil
}

func (t *CPCEMUTrack) Write(w io.Writer) error {
	if err := binary.Write(w, binary.LittleEndian, &t.ID); err != nil {
		fmt.Fprintf(os.Stderr, "Error while writing CPCEMUTrack.ID error :%v\n", err)
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, &t.Track); err != nil {
		fmt.Fprintf(os.Stderr, "Error while writing CPCEMUTrack.Track error :%v\n", err)
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, &t.Head); err != nil {
		fmt.Fprintf(os.Stderr, "Error while writing CPCEMUTrack.Head error :%v\n", err)
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, &t.Unused); err != nil {
		fmt.Fprintf(os.Stderr, "Error while writing CPCEMUTrack.Unused error :%v\n", err)
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, &t.SectSize); err != nil {
		fmt.Fprintf(os.Stderr, "Error while writing CPCEMUTrack.SectSize error :%v\n", err)
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, &t.SectorCount); err != nil {
		fmt.Fprintf(os.Stderr, "Error while writing CPCEMUTrack.SectorCount error :%v\n", err)
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, &t.Gap3); err != nil {
		fmt.Fprintf(os.Stderr, "Error while writing CPCEMUTrack.Gap3 error :%v\n", err)
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, &t.OctRemp); err != nil {
		fmt.Fprintf(os.Stderr, "Error while writing CPCEMUTrack.OctRemp error :%v\n", err)
		return err
	}

	//	fmt.Fprintf(os.Stdout,"Track:%s\n",c.ToString())
	var i uint8
	for i = 0; i < t.SectorCount; i++ {
		if err := t.Sect[i].Write(w); err != nil {
			fmt.Fprintf(os.Stderr, "Cannot read sector (%d) error :%v\n", i, err)
			return err
		}
	}
	for i = t.SectorCount; i < 29; i++ {
		sect := &CPCEMUSect{}
		err := sect.Write(w)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error while writing sector (%d), error :%v\n", i, err)
		}
	}
	if err := binary.Write(w, binary.LittleEndian, &t.Data); err != nil {
		fmt.Fprintf(os.Stderr, "Error while reading CPCEmuSect.Data error :%v\n", err)
		return err
	}
	return nil
}

func (t *CPCEMUTrack) ToString() string {
	return fmt.Sprintf("ID:%s, Track:%d, Head:%d, SectSize:%d, nbSect:%d,Gap3:%d",
		t.ID, t.Track, t.Head, t.SectSize, t.SectorCount, t.Gap3)
}

type StDirEntry struct {
	User      uint8 // User number (0-15), 0xE5=deleted file
	Name      [8]byte
	Ext       [3]byte
	NumPage   uint8 // page number (if a file is big, it will have more than one entry)
	Unused    [2]uint8
	PageCount uint8     // number of pages
	Blocks    [16]uint8 // blocks used by the file
}

type DSK struct {
	Entry          CPCEMUEnt
	TrackSizeTable []byte // dsk format  [0xCC]byte
	Tracks         []CPCEMUTrack
	BitMap         [256]byte
	Catalogue      [64]StDirEntry // 64 entries in directory (64*32=2048 bytes)
	Extended       bool           // extended dsk format flag
}

func (d *DSK) CleanBitmap() {
	for i := 0; i < 256; i++ {
		d.BitMap[i] = 0
	}
}

func (d *DSK) Read(r io.Reader) error {
	if err := binary.Read(r, binary.LittleEndian, &d.Entry); err != nil {
		fmt.Fprintf(os.Stderr, "Cannot read CPCEmuEnt error :%v\n", err)
		return err
	}
	mv := make([]byte, 4)
	extended := make([]byte, 16)
	copy(mv, d.Entry.Header[0:4])
	copy(extended, d.Entry.Header[0:16])
	if string(mv) != "MV -" && string(extended) != "EXTENDED CPC DSK" {
		return ErrorUnsupportedDskFormat
	}
	if strings.Contains(string(extended), "EXTENDED CPC DSK") {
		d.Extended = true
		d.TrackSizeTable = make([]byte, d.Entry.Heads*(d.Entry.Tracks))
	} else {
		d.TrackSizeTable = make([]byte, 0xCC)
	}
	if err := binary.Read(r, binary.LittleEndian, &d.TrackSizeTable); err != nil {
		fmt.Fprintf(os.Stderr, "Cannot read CPCEmuEnt TrackSizeTable error :%v\n", err)
		return err
	}

	if d.Extended {
		offset := make([]byte, (0x100 - (52 + uint(d.Entry.Heads*d.Entry.Tracks))))
		if err := binary.Read(r, binary.LittleEndian, &offset); err != nil {
			fmt.Fprintf(os.Stderr, "Cannot read CPCEmuEnt padding 0x100 error :%v\n", err)
			return err
		}
	}
	d.Tracks = make([]CPCEMUTrack, d.Entry.Tracks)
	for i := uint8(0); i < d.Entry.Tracks; i++ {
		//	fmt.Fprintf(os.Stdout,"Loading track %d, total: %d\n", i, cpcEntry.Tracks)
		track := &CPCEMUTrack{}
		if err := track.Read(r); err != nil {
			fmt.Fprintf(os.Stderr, "Error track (%d) error :%v\n", i, err)
			track = &CPCEMUTrack{}
		}
		d.Tracks[i] = *track
		// fmt.Fprintf(os.Stdout, "Track %d %s\n", i, d.Tracks[i].ToString())
	}
	return nil
}

func FormatDsk(sectors, tracks, heads uint8, diskFormat DskFormat, extendedDskType DskFileType) *DSK {
	dsk := &DSK{}
	entry := CPCEMUEnt{}
	if extendedDskType == DskTypeExtended || diskFormat == VendorFormat {
		dsk.Extended = true
		copy(entry.Header[:], "EXTENDED CPC DSK File\r\nDisk-Info\r\n")
	} else {
		copy(entry.Header[:], "MV - CPCEMU Disk-File\r\nDisk-Info\r\n")
	}
	copy(entry.Creator[:], "Sid DSK"[:])
	entry.DataSize = 0x100 + (SECTSIZE * uint16(sectors))
	entry.Tracks = tracks
	entry.Heads = heads
	if extendedDskType == DskTypeExtended {
		dsk.TrackSizeTable = make([]byte, entry.Heads*(entry.Tracks))
		for i := 0; i < len(dsk.TrackSizeTable); i++ {
			dsk.TrackSizeTable[i] = byte(0x100 + (SECTSIZE*uint16(sectors))/256 + 1)
		}
	} else {
		switch diskFormat {
		case DataFormat:
			dsk.TrackSizeTable = make([]byte, 0xCC)
		case VendorFormat:
			dsk.TrackSizeTable = make([]byte, entry.Heads*(entry.Tracks))
			for i := 0; i < len(dsk.TrackSizeTable); i++ {
				dsk.TrackSizeTable[i] = byte(0x100 + (SECTSIZE*uint16(sectors))/256 + 1)
			}
		default:
			fmt.Fprintf(os.Stderr, "Unknown format track.")
		}
	}
	dsk.Entry = entry
	dsk.Tracks = make([]CPCEMUTrack, tracks*heads)
	var i uint8
	if heads == 1 {
		for i = 0; i < tracks; i++ {
			switch diskFormat {
			case DataFormat:
				dsk.FormatTrack(i, i, 0, 0xC1, sectors)
			case VendorFormat:
				dsk.FormatTrack(i, i, 0, 0x41, sectors)
			default:
				fmt.Fprintf(os.Stderr, "Unknown format track.")
			}
		}
	} else {
		index := 0
		for i = 0; i < tracks; i++ {
			switch diskFormat {
			case DataFormat:
				dsk.FormatTrack(uint8(index), i, 0, 0xC1, sectors)
			case VendorFormat:
				dsk.FormatTrack(uint8(index), i, 0, 0x41, sectors)
			default:
				fmt.Fprintf(os.Stderr, "Unknown format track.")
			}
			index += 2
		}
		index = 1
		for i = 1; i < tracks; i++ {
			switch diskFormat {
			case DataFormat:
				dsk.FormatTrack(uint8(index), i, 1, 0xC1, sectors)
			case VendorFormat:
				dsk.FormatTrack(uint8(index), i, 1, 0x41, sectors)
			default:
				fmt.Fprintf(os.Stderr, "Unknown format track.")
			}
			index += 2
		}
	}
	return dsk
}

func (d *DSK) FormatTrack(indexTrack, track, head, minSect, nbSect uint8) {
	t := CPCEMUTrack{}
	copy(t.ID[:], "Track-Info\r\n")
	t.Track = track
	t.Head = head
	t.SectSize = 2
	t.SectorCount = nbSect
	t.Gap3 = 0x4E
	t.OctRemp = 0xE5
	//
	// Interleaving sectors
	//
	var s uint8
	var ss uint8
	var sectorSize uint16
	for s = 0; s < nbSect; {
		t.Sect[s].C = track
		t.Sect[s].H = head
		t.Sect[s].S = (ss + minSect)
		t.Sect[s].N = 2
		t.Sect[s].Size = 0x200
		sectorSize += t.Sect[s].Size
		ss++
		s++
		if s < nbSect {
			t.Sect[s].C = track
			t.Sect[s].H = head
			t.Sect[s].S = (ss + minSect + 4)
			t.Sect[s].N = 2
			t.Sect[s].Size = 0x200
			sectorSize += t.Sect[s].Size
			s++
		}
	}
	t.Data = make([]byte, sectorSize)
	for i := 0; i < len(t.Data); i++ {
		t.Data[i] = 0xE5
	}
	if len(d.Tracks) < int(track+1) {
		d.Tracks = append(d.Tracks, t)
		d.Entry.Tracks++
	} else {
		d.Tracks[indexTrack] = t
	}
}

func (d *DSK) Write(w io.Writer) error {
	if err := binary.Write(w, binary.LittleEndian, &d.Entry); err != nil {
		fmt.Fprintf(os.Stderr, "Cannot write CPCEmuEnt error :%v\n", err)
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, &d.TrackSizeTable); err != nil {
		fmt.Fprintf(os.Stderr, "Cannot write CPCEmuEnt error :%v\n", err)
		return err
	}
	if d.Extended {
		offset := make([]byte, (0x100 - (52 + uint(d.Entry.Heads*d.Entry.Tracks))))
		if err := binary.Write(w, binary.LittleEndian, &offset); err != nil {
			fmt.Fprintf(os.Stderr, "Cannot write CPCEmuEnt padding 0x100 error :%v\n", err)
			return err
		}
	}
	var i uint8
	for i = 0; i < d.Entry.Tracks; i++ {
		if err := d.Tracks[i].Write(w); err != nil {
			fmt.Fprintf(os.Stderr, "Error track (%d) error :%v\n", i, err)
		}
	}
	return nil
}

func WriteDsk(filePath string, d *DSK) error {
	f, err := os.Create(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot open file (%s) error %v\n", filePath, err)
		return err
	}
	if err := d.Write(f); err != nil {
		fmt.Fprintf(os.Stderr, "Error while reading (%s) error %v", filePath, err)
	}
	f.Close()
	return nil
}

func ReadDsk(filePath string) (*DSK, error) {
	f, err := os.Open(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot open file (%s) error %v\n", filePath, err)
		return &DSK{}, err
	}

	dsk := &DSK{}
	if err := dsk.Read(f); err != nil {
		fmt.Fprintf(os.Stderr, "Error while reading (%s) error %v", filePath, err)
	}
	f.Close()

	return dsk, nil
}

func (d *DSK) CheckDsk() error {
	// if d.Entry.Heads == 1 {
	minSectFirst := d.FirstSectorId()
	if minSectFirst != 0x41 && minSectFirst != 0xc1 && minSectFirst != 0x01 {
		fmt.Fprintf(os.Stderr, "Bad sector %.2x\n", minSectFirst)
		return ErrorBadSectorNumber
	}
	if d.Entry.Tracks > 80 {
		d.Entry.Tracks = 80
	}
	var track uint8
	for track = 0; track < d.Entry.Tracks; track++ {
		tr := d.Tracks[track]
		if !d.Extended {
			if tr.SectorCount != 9 {
				fmt.Fprintf(os.Stderr, "Warning : track :%d has %d sectors ! wanted 9\n", track, tr.SectorCount)
			}
		}
		var minSect, maxSect, s uint8
		minSect = 0xFF
		maxSect = 0
		nbSecteur := int(tr.SectorCount)
		if nbSecteur > len(tr.Sect) {
			nbSecteur = len(tr.Sect)
		}
		for s = 0; s < uint8(nbSecteur); s++ {
			if minSect > tr.Sect[s].S {
				minSect = tr.Sect[s].S
			}
			if maxSect < tr.Sect[s].S {
				maxSect = tr.Sect[s].S
			}
		}
		if !d.Extended {
			if maxSect-minSect != 8 {
				fmt.Fprintf(os.Stderr, "Warning : strange sector numbering in track %d! (maxSect:%X,minSect:%X)\n", track, maxSect, minSect)
			}
		}
		if !d.Extended {
			if minSect != minSectFirst {
				fmt.Fprintf(os.Stderr, "Warning : track %d start at sector %d while track 0 starts at %d\n", track, minSect, minSectFirst)
			}
		}
	}
	return nil
	//}
	//	return ErrorUnsupportedMultiHeadDsk
}

// FirstSectorId finds the lowest sector number of a track
func (d *DSK) FirstSectorId() uint8 {
	var Sect uint8 = 0xFF
	var s uint8
	tr := d.Tracks[0]
	// fmt.Fprintf(os.Stdout, "Track 0 nbSect :%d \n", tr.SectorCount)
	for s = 0; s < tr.SectorCount; s++ {
		//	fmt.Fprintf(os.Stdout, "Sector %d, S %d\n", s, tr.Sect[s].S)
		if Sect > tr.Sect[s].S {
			Sect = tr.Sect[s].S
		}
	}
	return Sect
}

// GetPosData returns the position of a sector in the DSK file, position in the Data structure
func (d *DSK) GetPosData(track, sect uint8, SectPhysique bool) uint16 {
	tr := d.Tracks[track]
	var SizeByte uint16
	var Pos uint16
	var s uint8
	// Pos += 256

	// fmt.Fprintf(os.Stdout,"Track:%d,Secteur:%d\n",track,sect)

	// Pos += 256
	for s = 0; s < tr.SectorCount; s++ {
		if (tr.Sect[s].S == sect && SectPhysique) || (s == sect && !SectPhysique) {
			break
		}
		SizeByte = tr.Sect[s].Size
		if SizeByte != 0 {
			Pos += SizeByte
		} else {
			Pos += (128 << tr.Sect[s].N)
		}
		// fmt.Fprintf(os.Stderr, "sizebyte:%d, t:%d,s:%d,tr.Sect[s].Size:%d, tr->Sect[ s ].N:%d, Pos:%d\n", Size,t,s,tr.Sect[s].Size,tr.Sect[ s ].N,Pos)
	}

	return Pos
}

func (d *DSK) GetFile(path string, indice int) error {
	i := indice
	current := make([]byte, 16)
	nomIndice := make([]byte, 16)
	lMax := 0x1000000
	cumul := 0
	err := d.RefreshCatalogue()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error while getting the catalogue, error :%v\n", err)
	}
	copy(nomIndice, d.Catalogue[i].Name[:])
	copy(nomIndice, d.Catalogue[i].Ext[:])
	fw, err := os.Create(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot open file (%s), error :%v\n", path, err)
		return err
	}
	defer fw.Close()
	for {
		// File length (?)
		var j uint8
		l := (d.Catalogue[i].PageCount + 7) >> 3
		for j = 0; j < l; j++ {
			tailleBloc := 1024
			p := d.ReadBloc(int(d.Catalogue[i].Blocks[j]))
			var nbOctets int
			if lMax > tailleBloc {
				nbOctets = tailleBloc
			} else {
				nbOctets = lMax
			}
			if nbOctets > 0 {
				if err := binary.Write(fw, binary.LittleEndian, p); err != nil {
					fmt.Fprintf(os.Stderr, "Cannot write data into file (%s) error %v\n", path, err)
				}
				cumul += nbOctets
			}
			lMax -= 1024
		}
		i++
		copy(current, d.Catalogue[i].Name[:])
		copy(current, d.Catalogue[i].Ext[:])
		if i > 64 {
			return ErrorCatalogueExceed
		}
		if string(nomIndice) == string(current) {
			break
		}
	} // while (! strncmp( NomIndice, current , max( strlen( NomIndice ), strlen( current ) )));

	return nil
}

func (d *DSK) DskSize() uint16 {
	err := d.RefreshCatalogue()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error while getting the catalogue, error :%v\n", err)
	}
	// var size uint16
	// size = len(d.Tracks) *
	// for _, t := range d.Tracks {
	// 	for _, s := range t.Sect {
	// 		size += uint16(s.N)
	// 	}
	// 	//	size /= uint16(t.SectorCount)
	// }
	return d.Entry.DataSize * uint16(d.Entry.Tracks)
}

func GetAmsDosName(mask string) string {
	amsdosFile := make([]byte, 12)
	for i := 0; i < 12; i++ {
		amsdosFile[i] = ' '
	}
	file := strings.ToUpper(strings.TrimSuffix(filepath.Base(mask), filepath.Ext(filepath.Base(mask))))
	filenameSize := len(file)
	if filenameSize > 8 {
		filenameSize = 8
	}
	copy(amsdosFile[0:filenameSize], file[0:filenameSize])
	amsdosFile[8] = '.'
	ext := strings.ToUpper(filepath.Ext(mask))
	copy(amsdosFile[9:12], ext[1:])
	return string(amsdosFile)
}

func (d *DSK) PutFile(path string, typeModeImport SaveMode, loadAddress, exeAddress, userNumber uint16, isSystemFile, readOnly bool) error {
	buff := make([]byte, 0x20000)
	cFileName := GetAmsDosName(path)
	header := &amsdos.AmsDosHeader{}
	var addHeader bool
	var err error

	err = d.RefreshCatalogue()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error while getting the catalogue, error :%v\n", err)
	}
	fr, err := os.Open(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot read file (%s) error :%v\n", path, err)
		return err
	}
	fileLength, err := fr.Read(buff)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot read the content of the file (%s) with error %v\n", path, err)
		return err
	}
	fmt.Fprintf(os.Stderr, "file (%s) read (%d bytes).\n", path, fileLength)
	_, err = fr.Seek(0, io.SeekStart)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error while seeking in file error :%v\n", err)
	}

	if err = binary.Read(fr, binary.LittleEndian, header); err != nil {
		fmt.Fprintf(os.Stderr, "No header found for file :%s, error :%v\n", path, err)
	}

	if typeModeImport == SaveModeAscii && fileLength%128 != 0 {
		buff[fileLength] = 0x1A
	}

	if typeModeImport == SaveModeProtected && fileLength%128 != 0 {
		buff[fileLength] = 0x1A
	}

	var isAmsdos bool
	//
	// Check if the file contains a header or not
	//
	if err == nil && header.Checksum == header.ComputedChecksum16() {
		isAmsdos = true
	}
	if !isAmsdos {
		// Create a default amsdos header
		fmt.Fprintf(os.Stderr, "Create header... (%s)\n", path)
		header = &amsdos.AmsDosHeader{}
		header.User = byte(userNumber)
		header.Size = uint16(fileLength)
		header.Size2 = uint16(fileLength)
		header.LogicalSize = uint16(fileLength)
		copy(header.Filename[:], []byte(cFileName[0:12]))
		header.Address = loadAddress
		if loadAddress != 0 {
			typeModeImport = SaveModeBinary
		}
		header.Exec = exeAddress
		if exeAddress != 0 || loadAddress != 0 {
			typeModeImport = SaveModeBinary
		}
		header.Type = byte(typeModeImport)

		// We must recalculate the checksum by counting addresses!
		header.Checksum = header.ComputedChecksum16()

	} else {
		fmt.Fprintf(os.Stderr, "File has already header...(%s)\n", path)
	}
	//
	// Depending on the import mode...
	//
	switch typeModeImport {
	case SaveModeAscii:
		//
		// Importation en mode ASCII
		//
		if isAmsdos {
			// Remove header if it exists
			fmt.Fprintf(os.Stderr, "Removing header...(%s)\n", path)
			copy(buff[0:], buff[binary.Size(amsdos.AmsDosHeader{}):])
		}
	case SaveModeBinary:
		//
		// Binary mode import
		//

		if !isAmsdos {
			//
			// Indicates that a header must be added
			//
			addHeader = true
		}
	}
	//
	// If file is ok to be imported
	//
	if addHeader {
		// Add the amsdos header if necessary

		var rbuff bytes.Buffer
		err = binary.Write(&rbuff, binary.LittleEndian, header)
		if err != nil {
			fmt.Fprintf(os.Stdout, "error while writing in header %v\n", err)
		}
		err = binary.Write(&rbuff, binary.LittleEndian, buff)
		if err != nil {
			fmt.Fprintf(os.Stdout, "error while writing in content %v\n", err)
		}
		buff = rbuff.Bytes()
		//	memmove( &Buff[ sizeof( AmsDosHeader ) ], Buff, Lg );
		//         	memcpy( Buff, e, sizeof( AmsDosHeader ) );
		//       	Lg += sizeof( AmsDosHeader );
	}
	if fileLength > 65536 {
		return ErrorFileSizeExceed
	}
	// if (MODE_BINAIRE) ClearAmsdos(Buff); // Replace unused bytes by 0 in the header
	return d.CopyFile(buff, cFileName, uint16(fileLength), 256, userNumber, isSystemFile, readOnly)
}

// CopyFile copies a file on the DSK
//
// the size is determined by the number of PageCount
// check why different from another DSK
func (d *DSK) CopyFile(bufFile []byte, fileName string, fileLength, maxBloc, userNumber uint16, isSystemFile, readOnly bool) error {
	var nbPages int
	d.MapDisk()
	dirLoc := d.NewDirEntry(fileName)
	var posFile uint16                       // Build the entry to put in the catalog
	for posFile = 0; posFile < fileLength; { // For each block of the file
		posDir, err := d.FindFreeDirEntry() // Find first free entry in the catalog
		if err == nil {
			dirLoc.User = uint8(userNumber) // User number
			if isSystemFile {
				dirLoc.Ext[0] |= 0x80
			}
			if readOnly {
				dirLoc.Ext[0] |= 0x80
			}
			dirLoc.NumPage = uint8(nbPages) // entry number in the file
			nbPages++
			blocksCount := (int(fileLength) - int(posFile) + 127) >> 7 // Page size (we add 127 to round up)

			if blocksCount > 128 { // Limit to 128 blocks (128K) per file
				blocksCount = 128
			}

			dirLoc.PageCount = uint8(blocksCount)
			l := (dirLoc.PageCount + 7) >> 3 // Number of blocks=PageCount/8 rounded up
			for i := 0; i < 16; i++ {
				dirLoc.Blocks[i] = 0
			}
			var j uint8
			for j = 0; j < l; j++ { // For each block of the file
				block := d.FindFreeBlock(int(maxBloc)) // Find first free block
				//	fmt.Fprintf(os.Stdout,"Bloc:%d, MaxBloc:%d\n",block,maxBloc)
				if block != 0 {
					// and put it on the disk
					dirLoc.Blocks[j] = block
					err = d.WriteBloc(int(block), bufFile, posFile)
					if err != nil {
						fmt.Fprintf(os.Stdout, "error while writing block %v\n", err)
					}
					posFile += 1024 // Go to the next block

				} else {
					return ErrDiskFull
				}
			}
			// fmt.Fprintf(os.Stdout, "posDir:%d dirloc:%v\n", posDir, dirLoc)
			err = d.SetInfoDirEntry(posDir, dirLoc)
			if err != nil {
				fmt.Fprintf(os.Stdout, "error while set info in directory %v\n", err)
			}
		} else {
			return ErrorNoDirEntry
		}
	}
	return nil
}

// MapDisk updates the disk bitmap, based on the catalog - returns the number of blocks used
func (d *DSK) MapDisk() int {
	for i := 0; i < len(d.BitMap); i++ {
		d.BitMap[i] = 0
	}
	d.BitMap[0] = 1 // The first two blocks are reserved for the catalog
	d.BitMap[1] = 1
	var blocksUsed int
	for i := 0; i < 64; i++ {
		dir, _ := d.GetInfoDirEntry(uint8(i))
		if dir.User != USER_DELETED {
			for j := 0; j < 16; j++ {
				b := dir.Blocks[j]
				if b > 1 && d.BitMap[b] != 1 {
					d.BitMap[b] = 1
					blocksUsed++
				}
			}
		}
	}
	return blocksUsed
}

func (d *DSK) NewDirEntry(fileName string) StDirEntry {
	e := StDirEntry{}
	for i := 0; i < 8; i++ {
		e.Name[i] = ' '
	}

	for i := 0; i < 3; i++ {
		e.Ext[i] = ' '
	}
	copy(e.Ext[:], fileName[9:12])
	copy(e.Name[:], fileName[0:8])
	return e
}

func (d *DSK) CopyRawFile(bufFile []byte, fileLength uint16, track, sector int) (int, int, error) {
	d.MapDisk()

	var posFile uint16 // Build the entry to put in the catalog
	var err error
	var written int
	for posFile = 0; posFile < fileLength; { // For each block of the file
		track, sector, written, err = d.WriteAtTrackSector(track, sector, bufFile, posFile)
		if err != nil {
			return track, sector, err
		}
		posFile += uint16(written) // Move to the next block
	}
	return track, sector, nil
}

func (d *DSK) WriteAtTrackSector(track int, sect int, bufBloc []byte, offset uint16) (int, int, int, error) {
	var dataWritten int
	minSect := d.FirstSectorId()
	//
	// Adjusts the number of tracks if capacity is exceeded
	//
	if track > int(d.Entry.Tracks-1) {
		if d.Entry.Heads == 1 {
			d.FormatTrack(0, uint8(track), 0, minSect, 9)
		} else {
			currentHead := d.Tracks[track-1].Head
			if currentHead == 0 {
				d.FormatTrack(0, uint8(track), 0, minSect, 9)
			} else {
				d.FormatTrack(0, uint8(track), 1, minSect, 9)
			}
		}
	}
	if sect > 8 {
		track++
		sect = 0
	}
	sectorSize := d.Tracks[track].Sect[sect].Size
	pos := d.GetPosData(uint8(track), uint8(sect)+minSect, true)
	maxSize := sectorSize
	if len(bufBloc) < int(offset+maxSize) {
		maxSize = uint16(len(bufBloc))
	}
	dataWritten += copy(d.Tracks[track].Data[pos:], bufBloc[offset:offset+maxSize])
	if len(bufBloc) < int(offset+(maxSize*2)) {
		return track, sect, dataWritten, nil
	}
	sect++
	if sect > 8 {
		track++
		sect = 0
	}
	if track > int(d.Entry.Tracks-1) {
		if d.Entry.Heads == 1 {
			d.FormatTrack(0, uint8(track), 0, minSect, 9)
		} else {
			currentHead := d.Tracks[track-1].Head
			if currentHead == 0 {
				d.FormatTrack(0, uint8(track), 0, minSect, 9)
			} else {
				d.FormatTrack(0, uint8(track), 1, minSect, 9)
			}
		}
	}
	sectorSize = d.Tracks[track].Sect[sect].Size
	pos = d.GetPosData(uint8(track), uint8(sect)+minSect, true)
	maxSize = sectorSize*2 + offset
	if (len(bufBloc) - dataWritten - int(offset)) < int(maxSize) {
		maxSize = uint16(len(bufBloc))
	}
	dataWritten += copy(d.Tracks[track].Data[pos:], bufBloc[offset+sectorSize:maxSize])
	sect++
	return track, sect, dataWritten, nil
}

func (d *DSK) WriteBloc(bloc int, bufBloc []byte, offset uint16) error {
	track := (bloc << 1) / 9
	sect := (bloc << 1) % 9
	minSect := d.FirstSectorId()
	if minSect == 0x41 {
		track += 2
	} else {
		if minSect == 0x01 {
			track++
		}
	}
	//
	// Adjust the number of tracks if overflowing
	//
	if track > int(d.Entry.Tracks-1) {
		if d.Entry.Heads == 1 {
			d.FormatTrack(0, uint8(track), 0, minSect, 9)
		} else {
			currentHead := d.Tracks[track-1].Head
			if currentHead == 0 {
				d.FormatTrack(0, uint8(track), 0, minSect, 9)
			} else {
				d.FormatTrack(0, uint8(track), 1, minSect, 9)
			}
		}
	}
	if sect > 8 {
		track++
		sect = 0
	}
	pos := d.GetPosData(uint8(track), uint8(sect)+minSect, true)
	copy(d.Tracks[track].Data[pos:], bufBloc[offset:offset+SECTSIZE])
	sect++
	if sect > 8 {
		track++
		sect = 0
	}
	if track > int(d.Entry.Tracks-1) {
		if d.Entry.Heads == 1 {
			d.FormatTrack(0, uint8(track), 0, minSect, 9)
		} else {
			currentHead := d.Tracks[track-1].Head
			if currentHead == 0 {
				d.FormatTrack(0, uint8(track), 0, minSect, 9)
			} else {
				d.FormatTrack(0, uint8(track), 1, minSect, 9)
			}
		}
	}
	pos = d.GetPosData(uint8(track), uint8(sect)+minSect, true)
	copy(d.Tracks[track].Data[pos:], bufBloc[offset+SECTSIZE:offset+(SECTSIZE*2)])
	return nil
}

func (d *DSK) ExtractRawFile(fileLength uint16, track, sector int) (int, int, []byte) {
	d.MapDisk()
	content := make([]byte, 0)
	var posFile uint16 // Build the entry to put in the catalog
	var buf []byte

	for posFile = 0; posFile < fileLength; { // For each block of the file
		track, sector, buf = d.ReadAtTrackSector(track, sector)
		posFile += uint16(len(buf)) // Move to the next block
		content = append(content, buf...)
	}
	return track, sector, content
}

func (d *DSK) ReadAtTrackSector(track, sect int) (int, int, []byte) {
	minSect := d.FirstSectorId()
	if minSect == 0x41 {
		track += 2
	} else {
		if minSect == 0x01 {
			track++
		}
	}
	sectorSize := d.Tracks[track].Sect[sect].Size
	pos := d.GetPosData(uint8(track), uint8(sect)+minSect, true)
	bufBloc1 := make([]byte, sectorSize)
	copy(bufBloc1, d.Tracks[track].Data[pos:pos+sectorSize])
	// int Pos = GetPosData( track, sect + MinSect, true );
	// memcpy( BufBloc, &ImgDsk[ Pos ], SECTSIZE );
	sect++
	if sect > 8 {
		track++
		sect = 0
	}
	sectorSize = d.Tracks[track].Sect[sect].Size

	bufBloc2 := make([]byte, sectorSize)
	pos = d.GetPosData(uint8(track), uint8(sect)+minSect, true)
	copy(bufBloc2, d.Tracks[track].Data[pos:pos+sectorSize])
	sect++
	//   Pos = GetPosData( track, sect + MinSect, true );
	// memcpy( &BufBloc[ SECTSIZE ], &ImgDsk[ Pos ], SECTSIZE );
	bufBloc1 = append(bufBloc1, bufBloc2...)
	return track, sect, bufBloc1
}

func (d *DSK) ReadBloc(bloc int) []byte {
	bufBloc := make([]byte, SECTSIZE*2)
	track := (bloc << 1) / 9
	sect := (bloc << 1) % 9
	minSect := d.FirstSectorId()
	if minSect == 0x41 {
		track += 2
	} else {
		if minSect == 0x01 {
			track++
		}
	}
	pos := d.GetPosData(uint8(track), uint8(sect)+minSect, true)
	copy(bufBloc[0:], d.Tracks[track].Data[pos:pos+SECTSIZE])
	// int Pos = GetPosData( track, sect + MinSect, true );
	// memcpy( BufBloc, &ImgDsk[ Pos ], SECTSIZE );
	sect++
	if sect > 8 {
		track++
		sect = 0
	}
	pos = d.GetPosData(uint8(track), uint8(sect)+minSect, true)
	copy(bufBloc[SECTSIZE:], d.Tracks[track].Data[pos:pos+SECTSIZE])
	//   Pos = GetPosData( track, sect + MinSect, true );
	// memcpy( &BufBloc[ SECTSIZE ], &ImgDsk[ Pos ], SECTSIZE );
	return bufBloc
}

// FindFreeBlock finds first free block to be filled
func (d *DSK) FindFreeBlock(maxBloc int) uint8 {
	for i := 2; i < maxBloc; i++ {
		if d.BitMap[i] == 0 {
			d.BitMap[i] = 1
			return uint8(i)
		}
	}
	return 0
}

// FindFreeDirEntry finds first free directory entry
func (d *DSK) FindFreeDirEntry() (uint8, error) {
	for i := 0; i < 64; i++ {
		dir, _ := d.GetInfoDirEntry(uint8(i))
		if dir.User == USER_DELETED {
			return uint8(i), nil
		}
	}
	return 0, ErrorNoDirEntry
}

func (d *DSK) DisplayCatalogue() {
	err := d.RefreshCatalogue()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error while getting the catalogue error :%v\n", err)
	}
	for i := 0; i < 64; i++ {
		entry := d.Catalogue[i]
		if entry.User != USER_DELETED && entry.NumPage != 0 {
			fmt.Fprintf(os.Stderr, "%s.%s : %d\n", entry.Name, entry.Ext, entry.User)
		}
	}
}

func (d *DSK) GetCatEntryName(num int) string {
	err := d.RefreshCatalogue()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error while getting the catalogue error :%v\n", err)
	}
	var nom string
	for i := 0; i < 64; i++ {
		entry := d.Catalogue[i]
		if entry.User != USER_DELETED && entry.NumPage == 0 && i == num {
			nom = fmt.Sprintf("%.8s.%.3s", entry.Name, entry.Ext)
			//	fmt.Fprintf(os.Stdout,"%s.%s : %d\n",entry.Name,entry.Ext,entry.User )
			return nom
		}
	}
	return nom
}

func (d *DSK) GetCatEntrySizeStr(num int) string {
	err := d.RefreshCatalogue()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error while getting the catalogue error :%v\n", err)
	}
	for i := 0; i < 64; i++ {
		entry := d.Catalogue[i]
		if entry.User != USER_DELETED && entry.NumPage == 0 && i == num {
			var p, t int
			for {
				if d.Catalogue[p+i].User == entry.User {
					t += int(d.Catalogue[p+i].PageCount)
				}
				p++
				if d.Catalogue[p+i].NumPage != 0 || p+i >= 64 {
					break
				}
			}
			return fmt.Sprintf("%d KiB", (t+7)>>3) // (t+7)/8 to round up
		}
	}
	return ""
}

func (d *DSK) GetFilesize(s StDirEntry) int {
	t := 0
	for i := 0; i < 64; i++ {
		if d.Catalogue[i].User != USER_DELETED {
			if d.Catalogue[i].Name == s.Name &&
				d.Catalogue[i].Ext == s.Ext {
				t += int(d.Catalogue[i].PageCount)
			}
		}
	}
	return (t + 7) >> 3
}

func (d *DSK) GetFileIndices() []int {
	indices := make([]int, 0)
	cache := make(map[string]bool)
	for i := 0; i < 64; i++ {
		if d.Catalogue[i].User != USER_DELETED {
			filename := fmt.Sprintf("%s.%s", d.Catalogue[i].Name, d.Catalogue[i].Ext)
			if !cache[filename] {
				cache[filename] = true
				indices = append(indices, i)
			}
		}
	}

	return indices
}

func (d *DSK) RefreshCatalogue() error {
	for i := 0; i < 64; i++ {
		dirEntry, err := d.GetInfoDirEntry(uint8(i))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error while reading catalogue error :%v\n", err)
		}
		d.Catalogue[i] = dirEntry
	}
	return nil
}

func (d *DSK) SetInfoDirEntry(numDir uint8, e StDirEntry) error {
	minSect := d.FirstSectorId()
	s := (numDir >> 4) + minSect
	var t uint8
	if minSect == 0x41 {
		t = 2
	}

	if minSect == 1 {
		t = 1
	}
	var data bytes.Buffer

	if err := binary.Write(&data, binary.LittleEndian, &e); err != nil {
		fmt.Fprintf(os.Stderr, "Error while writing StDirEntry structure with error :%v\n", err)
		return err
	}
	entry := data.Bytes()
	for i := 0; i < 16; i++ {
		pos := d.GetPosData(t, s, true)
		copy(d.Tracks[t].Data[((uint16(numDir)&15)<<5)+pos:((uint16(numDir)&15)<<5)+pos+uint16(binary.Size(entry))], entry[:])
	}
	return nil
}

func (d *DSK) GetInfoDirEntry(numDir uint8) (StDirEntry, error) {
	dir := StDirEntry{}
	minSect := d.FirstSectorId()
	s := (numDir >> 4) + minSect
	var t uint8
	if minSect == 0x41 {
		t = 2
	}

	if minSect == 1 {
		t = 1
	}

	pos := d.GetPosData(t, s, true)
	if int(pos) >= len(d.Tracks[t].Data) {
		pos = 0
	}
	data := d.Tracks[t].Data[((uint16(numDir)&15)<<5)+pos : ((uint16(numDir)&15)<<5)+pos+32]
	buffer := bytes.NewReader(data[:])
	if err := binary.Read(buffer, binary.LittleEndian, &dir); err != nil {
		return dir, err
	}
	// memcpy( &Dir
	//		, &ImgDsk[ ( ( NumDir & 15 ) << 5 ) + GetPosData( t, s, true ) ]
	//		, sizeof( StDirEntry )
	//		);
	return dir, nil
}

func (d *DSK) FileTypeStr(ams *amsdos.AmsDosHeader) string {
	if ams.Checksum == ams.ComputedChecksum16() {
		switch ams.Type {
		case 0:
			return "BASIC"
		case 1:
			return "BASIC(P)"
		case 2:
			return "BINARY"
		case 3:
			return "BINARY(P)"
		default:
			return "UNKNOWN"

		}
	}
	return "ASCII"
}

// FileExists checks if a file exists on the DSK and returns its index in the directory
func (d *DSK) FileExists(entry StDirEntry) int {
	for i := 0; i < 64; i++ {
		dir, err := d.GetInfoDirEntry(uint8(i))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error while getting info dir entry (%d) error :%v\n", i, err)
		} else {
			for q := 0; q < 8; q++ {
				dir.Name[q] &= 127
			}
			for q := 0; q < 3; q++ {
				dir.Ext[q] &= 127
			}
			if dir.User != USER_DELETED && dir.Name == entry.Name && dir.Ext == entry.Ext {
				return i
			}
		}

	}
	return NOT_FOUND
}

func (d *DSK) GetFileIn(filename string, indice int) ([]byte, error) {
	i := indice
	bytesLeft := 0x1000000
	b := make([]byte, 0)
	firstBlock := true
	/*	tabDir := make([]StDirEntry, 64)
		for j := 0; j < 64; j++ {
			tabDir[j], _ = d.GetInfoDirEntry(uint8(j))
		}*/
	err := d.RefreshCatalogue()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error while getting the catalogue error :%v\n", err)
	}
	dirEntry := d.Catalogue[i]
	var bytesRead, totalSize int
	var isAmsdos bool
	for {
		l := (d.Catalogue[i].PageCount + 7) >> 3
		for j := 0; uint8(j) < l; j++ {
			blockSize := 1024
			bloc := d.ReadBloc(int(d.Catalogue[i].Blocks[j]))
			if firstBlock {
				var header *cpc.CpcHead
				isAmsdos, header = amsdos.CheckAmsdos(bloc)
				if isAmsdos {
					totalSize = int(header.Size) + 0x80
				}
				firstBlock = false
			}
			var byteCount = min(bytesLeft, blockSize)
			if byteCount > 0 {
				b = append(b, bloc...)
				bytesRead += byteCount
			}
			bytesLeft -= byteCount
		}
		i++
		if i >= 64 {
			return b, errors.New("cannot get the file, Exceed catalogue indice")
		}
		if dirEntry.Name != d.Catalogue[i].Name || dirEntry.Ext != d.Catalogue[i].Ext {
			break
		}
	}
	// totalSize can stay 0 if the file doesn't have the AMSDOS header (size is unknown)
	if totalSize <= 0 || totalSize <= bytesRead {
		totalSize = bytesRead
	}

	/*if !isAmsdos {
		keepOn := true
		for i = totalSize - 1; i >= 0; i-- {
			if b[i] == 0 {
				totalSize--
			} else {
				keepOn = false
			}
			if !keepOn {
				break
			}
		}
	}*/

	return b[0:totalSize], nil
}

func (d *DSK) ViewFile(indice int) ([]byte, int, error) {
	i := indice
	bytesLeft := 0x1000000
	b := make([]byte, 0)
	firstBlock := true
	err := d.RefreshCatalogue()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error while getting the catalogue error :%v\n", err)
	}
	dirEntry := d.Catalogue[i]
	var fileSize, bytesRead int
	for {
		l := (d.Catalogue[i].PageCount + 7) >> 3
		var j uint8
		for j = 0; j < l; j++ {
			blockSize := 1024
			block := d.ReadBloc(int(d.Catalogue[i].Blocks[j]))
			if firstBlock {
				isAmsdos, header := amsdos.CheckAmsdos(block)
				if isAmsdos {
					t := make([]byte, len(block))
					copy(t, block[HeaderSize:])
					block = t
					blockSize -= HeaderSize
					fileSize = int(header.Size)
				}
				firstBlock = false
			}
			var byteCount = min(bytesLeft, blockSize)
			if byteCount > 0 {
				b = append(b, block...)
				bytesRead += byteCount
			}
			bytesLeft -= 1024
		}
		i++
		if i >= 64 {
			return b, bytesRead, errors.New("cannot get the file, Exceed catalogue indice")
		}
		if dirEntry.Name != d.Catalogue[i].Name || dirEntry.Ext != d.Catalogue[i].Ext {
			break
		}
	}
	if fileSize == 0 {
		fileSize = bytesRead
	}
	// for files padded with 0xE5, remove the padding from total size
	for i := len(b) - 1; i >= 0; i-- {
		if b[i] == 0xE5 {
			fileSize = i
		} else {
			break
		}
	}
	return b, fileSize, nil
}

func (d *DSK) RemoveFile(index uint8) error {
	err := d.RefreshCatalogue()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error while getting the catalogue error :%v\n", err)
	}
	firstDirEntry := d.Catalogue[index]

	// this assumes files are contiguous in the directory, is it always the case?
	for {
		entry, err := d.GetInfoDirEntry(index)
		if err != nil {
			return ErrorNoDirEntry
		}
		if firstDirEntry.Name != d.Catalogue[index].Name || firstDirEntry.Ext != d.Catalogue[index].Ext {
			break
		}
		d.Catalogue[index].User = USER_DELETED
		entry.User = USER_DELETED
		if err := d.SetInfoDirEntry(index, entry); err != nil {
			return ErrorNoDirEntry
		}
		index++
	}
	return nil
}
