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

var (
	USER_DELETED uint8  = 0xE5
	SECTSIZE     uint16 = 512
	NOT_FOUND    int    = -1
)

var (
	ErrorUnsupportedDskFormat    = errors.New("unsupported DSK Format")
	ErrorUnsupportedMultiHeadDsk = errors.New("multi-side dsk ! Expected 1 head")
	ErrorBadSectorNumber         = errors.New("dsk has wrong sector number")
	ErrorCatalogueExceed         = errors.New("catalogue indice exceed")
	ErrorNoBloc                  = errors.New("error no more block available")
	ErrorNoDirEntry              = errors.New("error no more dir entry available")
	ErrorFileSizeExceed          = errors.New("filesize exceed")
)

var (
	MODE_ASCII        uint8     = 0
	MODE_PROTECTED    uint8     = 1
	MODE_BINAIRE      uint8     = 2
	EXTENDED_DSK_TYPE           = 1
	DSK_TYPE                    = 0
	DataFormat        DskFormat = 0
	VendorFormat      DskFormat = 1
)

const HeaderSize = 0x80

type DskFormat = int

type CPCEMUEnt struct {
	Debut    [0x22]byte // "MV - CPCEMU Disk-File\r\nDisk-Info\r\n"
	Creator  [0xE]byte
	NbTracks uint8
	NbHeads  uint8
	DataSize uint16 // 0x1300 = 256 + ( 512 * nbsecteurs )
}

func (e *CPCEMUEnt) ToString() string {
	return fmt.Sprintf("Debut:%s\nCreator:%s\nnbTracks:%d, nbHeads:%d, DataSize:%d",
		e.Debut, e.Creator, e.NbTracks, e.NbHeads, e.DataSize)
}

type CPCEMUSect struct { // length 8
	C        uint8 // track,
	H        uint8 // head
	R        uint8 // sect
	N        uint8 // size
	Un1      uint16
	SizeByte uint16 // Taille secteur en octets
	//
}

func (c *CPCEMUSect) ToString() string {
	return fmt.Sprintf("C:%d,H:%d,R:%d,N:%d,Un1:%d:SizeByte:%d", //,DataSize:%d",
		c.C, c.H, c.R, c.N, c.Un1, c.SizeByte) //, len(c.Data))
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
	if err := binary.Write(w, binary.LittleEndian, &c.R); err != nil {
		fmt.Fprintf(os.Stderr, "Error while writing CPCEmuSect.R error :%v\n", err)
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
	if err := binary.Write(w, binary.LittleEndian, &c.SizeByte); err != nil {
		fmt.Fprintf(os.Stderr, "Error while writing CPCEmuSect.SizeByte error :%v\n", err)
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
	if err := binary.Read(r, binary.LittleEndian, &c.R); err != nil {
		fmt.Fprintf(os.Stderr, "Error while reading CPCEmuSect.R error :%v\n", err)
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
	if err := binary.Read(r, binary.LittleEndian, &c.SizeByte); err != nil {
		fmt.Fprintf(os.Stderr, "Error while reading CPCEmuSect.SizeByte error :%v\n", err)
		return err
	}
	//	fmt.Fprintf(os.Stdout,"Sector %s\n",c.ToString())
	return nil
}

type CPCEMUTrack struct { // length 18 bytes
	ID       [0x10]byte // "Track-Info\r\n"
	Track    uint8
	Head     uint8
	Unused   [2]byte
	SectSize uint8 // 2
	NbSect   uint8 // 9
	Gap3     uint8 // 0x4E
	OctRemp  uint8 // 0xE5
	Sect     [29]CPCEMUSect
	Data     []byte
}

func (c *CPCEMUTrack) Read(r io.Reader) error {
	if err := binary.Read(r, binary.LittleEndian, &c.ID); err != nil {
		fmt.Fprintf(os.Stderr, "Error while reading CPCEMUTrack.ID error :%v\n", err)
		return err
	}
	if err := binary.Read(r, binary.LittleEndian, &c.Track); err != nil {
		fmt.Fprintf(os.Stderr, "Error while reading CPCEMUTrack.Track error :%v\n", err)
		return err
	}
	if err := binary.Read(r, binary.LittleEndian, &c.Head); err != nil {
		fmt.Fprintf(os.Stderr, "Error while reading CPCEMUTrack.Head error :%v\n", err)
		return err
	}
	if err := binary.Read(r, binary.LittleEndian, &c.Unused); err != nil {
		fmt.Fprintf(os.Stderr, "Error while reading CPCEMUTrack.Unused error :%v\n", err)
		return err
	}
	if err := binary.Read(r, binary.LittleEndian, &c.SectSize); err != nil {
		fmt.Fprintf(os.Stderr, "Error while reading CPCEMUTrack.SectSize error :%v\n", err)
		return err
	}
	if err := binary.Read(r, binary.LittleEndian, &c.NbSect); err != nil {
		fmt.Fprintf(os.Stderr, "Error while reading CPCEMUTrack.NbSect error :%v\n", err)
		return err
	}
	if err := binary.Read(r, binary.LittleEndian, &c.Gap3); err != nil {
		fmt.Fprintf(os.Stderr, "Error while reading CPCEMUTrack.Gap3 error :%v\n", err)
		return err
	}
	if err := binary.Read(r, binary.LittleEndian, &c.OctRemp); err != nil {
		fmt.Fprintf(os.Stderr, "Error while reading CPCEMUTrack.OctRemp error :%v\n", err)
		return err
	}

	//	fmt.Fprintf(os.Stdout,"Track:%s\n",c.ToString())
	var i uint8
	var sectorSize uint16
	for i = 0; i < c.NbSect && i < 29; i++ {
		sect := &CPCEMUSect{}
		if err := sect.Read(r); err != nil {
			fmt.Fprintf(os.Stderr, "Cannot read sector (%d) error :%v\n", i, err)
			return err
		}
		c.Sect[i] = *sect
		sectorSize += c.Sect[i].SizeByte
	}
	for i = c.NbSect; i < 29; i++ {
		sect := &CPCEMUSect{}
		err := sect.Read(r)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error while reading sector (%d), error :%v\n", i, err)
		}
	}
	if int(sectorSize) > int(c.SectSize)*0x100*int(c.NbSect) {
		fmt.Fprintf(os.Stderr, "Warning : Sector size [%d] differs from the amount of data found [%d], enlarge data part\n",
			int(c.SectSize)*0x100*int(c.NbSect),
			sectorSize)
		c.Data = make([]byte, sectorSize)
	} else {
		c.Data = make([]byte, int(c.SectSize)*0x100*int(c.NbSect))
	}
	if err := binary.Read(r, binary.LittleEndian, &c.Data); err != nil {
		fmt.Fprintf(os.Stderr, "Error while reading CPCEmuSect.Data error :%v\n", err)
		return err
	}
	return nil
}

func (c *CPCEMUTrack) Write(w io.Writer) error {
	if err := binary.Write(w, binary.LittleEndian, &c.ID); err != nil {
		fmt.Fprintf(os.Stderr, "Error while writing CPCEMUTrack.ID error :%v\n", err)
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, &c.Track); err != nil {
		fmt.Fprintf(os.Stderr, "Error while writing CPCEMUTrack.Track error :%v\n", err)
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, &c.Head); err != nil {
		fmt.Fprintf(os.Stderr, "Error while writing CPCEMUTrack.Head error :%v\n", err)
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, &c.Unused); err != nil {
		fmt.Fprintf(os.Stderr, "Error while writing CPCEMUTrack.Unused error :%v\n", err)
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, &c.SectSize); err != nil {
		fmt.Fprintf(os.Stderr, "Error while writing CPCEMUTrack.SectSize error :%v\n", err)
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, &c.NbSect); err != nil {
		fmt.Fprintf(os.Stderr, "Error while writing CPCEMUTrack.NbSect error :%v\n", err)
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, &c.Gap3); err != nil {
		fmt.Fprintf(os.Stderr, "Error while writing CPCEMUTrack.Gap3 error :%v\n", err)
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, &c.OctRemp); err != nil {
		fmt.Fprintf(os.Stderr, "Error while writing CPCEMUTrack.OctRemp error :%v\n", err)
		return err
	}

	//	fmt.Fprintf(os.Stdout,"Track:%s\n",c.ToString())
	var i uint8
	for i = 0; i < c.NbSect; i++ {
		if err := c.Sect[i].Write(w); err != nil {
			fmt.Fprintf(os.Stderr, "Cannot read sector (%d) error :%v\n", i, err)
			return err
		}
	}
	for i = c.NbSect; i < 29; i++ {
		sect := &CPCEMUSect{}
		err := sect.Write(w)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error while writing sector (%d), error :%v\n", i, err)
		}
	}
	if err := binary.Write(w, binary.LittleEndian, &c.Data); err != nil {
		fmt.Fprintf(os.Stderr, "Error while reading CPCEmuSect.Data error :%v\n", err)
		return err
	}
	return nil
}

func (t *CPCEMUTrack) ToString() string {
	return fmt.Sprintf("ID:%s, Track:%d, Head:%d, SectSize:%d, nbSect:%d,Gap3:%d",
		t.ID, t.Track, t.Head, t.SectSize, t.NbSect, t.Gap3)
}

type StDirEntry struct {
	User    uint8
	Nom     [8]byte
	Ext     [3]byte
	NumPage uint8
	Unused  [2]uint8
	NbPages uint8
	Blocks  [16]uint8
}

type DSK struct {
	Entry           CPCEMUEnt
	TrackSizeTable  []byte // dsk format  [0xCC]byte
	Tracks          []CPCEMUTrack
	BitMap          [256]byte
	Catalogue       [64]StDirEntry
	catalogueLoaded bool
	Extended        bool
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
	copy(mv, d.Entry.Debut[0:4])
	copy(extended, d.Entry.Debut[0:16])
	if string(mv) != "MV -" && string(extended) != "EXTENDED CPC DSK" {
		return ErrorUnsupportedDskFormat
	}
	if strings.Contains(string(extended), "EXTENDED CPC DSK") {
		d.Extended = true
		d.TrackSizeTable = make([]byte, d.Entry.NbHeads*(d.Entry.NbTracks))
	} else {
		d.TrackSizeTable = make([]byte, 0xCC)
	}
	if err := binary.Read(r, binary.LittleEndian, &d.TrackSizeTable); err != nil {
		fmt.Fprintf(os.Stderr, "Cannot read CPCEmuEnt TrackSizeTable error :%v\n", err)
		return err
	}

	if d.Extended {
		offset := make([]byte, (0x100 - (52 + uint(d.Entry.NbHeads*d.Entry.NbTracks))))
		if err := binary.Read(r, binary.LittleEndian, &offset); err != nil {
			fmt.Fprintf(os.Stderr, "Cannot read CPCEmuEnt padding 0x100 error :%v\n", err)
			return err
		}
	}
	d.Tracks = make([]CPCEMUTrack, d.Entry.NbTracks)
	var i uint8
	for i = 0; i < d.Entry.NbTracks; i++ {
		//	fmt.Fprintf(os.Stdout,"Loading track %d, total: %d\n", i, cpcEntry.NbTracks)
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

func FormatDsk(nbSect, nbTrack, nbHead uint8, diskFormat DskFormat, extendedDskType int) *DSK {
	dsk := &DSK{}
	entry := CPCEMUEnt{}
	if extendedDskType == EXTENDED_DSK_TYPE || diskFormat == VendorFormat {
		dsk.Extended = true
		copy(entry.Debut[:], "EXTENDED CPC DSK File\r\nDisk-Info\r\n")
	} else {
		copy(entry.Debut[:], "MV - CPCEMU Disk-File\r\nDisk-Info\r\n")
	}
	copy(entry.Creator[:], "Sid DSK"[:])
	entry.DataSize = 0x100 + (SECTSIZE * uint16(nbSect))
	entry.NbTracks = nbTrack
	entry.NbHeads = nbHead
	if extendedDskType == EXTENDED_DSK_TYPE {
		dsk.TrackSizeTable = make([]byte, entry.NbHeads*(entry.NbTracks))
		for i := 0; i < len(dsk.TrackSizeTable); i++ {
			dsk.TrackSizeTable[i] = byte(0x100 + (SECTSIZE*uint16(nbSect))/256 + 1)
		}
	} else {
		switch diskFormat {
		case DataFormat:
			dsk.TrackSizeTable = make([]byte, 0xCC)
		case VendorFormat:
			dsk.TrackSizeTable = make([]byte, entry.NbHeads*(entry.NbTracks))
			for i := 0; i < len(dsk.TrackSizeTable); i++ {
				dsk.TrackSizeTable[i] = byte(0x100 + (SECTSIZE*uint16(nbSect))/256 + 1)
			}
		default:
			fmt.Fprintf(os.Stderr, "Unknown format track.")
		}
	}
	dsk.Entry = entry
	dsk.Tracks = make([]CPCEMUTrack, nbTrack*nbHead)
	var i uint8
	if nbHead == 1 {
		for i = 0; i < nbTrack; i++ {
			switch diskFormat {
			case DataFormat:
				dsk.FormatTrack(i, i, 0, 0xC1, nbSect)
			case VendorFormat:
				dsk.FormatTrack(i, i, 0, 0x41, nbSect)
			default:
				fmt.Fprintf(os.Stderr, "Unknown format track.")
			}
		}
	} else {
		index := 0
		for i = 0; i < nbTrack; i++ {
			switch diskFormat {
			case DataFormat:
				dsk.FormatTrack(uint8(index), uint8(i), 0, 0xC1, nbSect)
			case VendorFormat:
				dsk.FormatTrack(uint8(index), uint8(i), 0, 0x41, nbSect)
			default:
				fmt.Fprintf(os.Stderr, "Unknown format track.")
			}
			index += 2
		}
		index = 1
		for i = 1; i < nbTrack; i++ {
			switch diskFormat {
			case DataFormat:
				dsk.FormatTrack(uint8(index), uint8(i), 1, 0xC1, nbSect)
			case VendorFormat:
				dsk.FormatTrack(uint8(index), uint8(i), 1, 0x41, nbSect)
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
	t.NbSect = nbSect
	t.Gap3 = 0x4E
	t.OctRemp = 0xE5
	//
	// Gestion "entrelacement" des secteurs
	//
	var s uint8
	var ss uint8
	var sectorSize uint16
	for s = 0; s < nbSect; {
		t.Sect[s].C = track
		t.Sect[s].H = head
		t.Sect[s].R = (ss + minSect)
		t.Sect[s].N = 2
		t.Sect[s].SizeByte = 0x200
		sectorSize += t.Sect[s].SizeByte
		ss++
		s++
		if s < nbSect {
			t.Sect[s].C = track
			t.Sect[s].H = head
			t.Sect[s].R = (ss + minSect + 4)
			t.Sect[s].N = 2
			t.Sect[s].SizeByte = 0x200
			sectorSize += t.Sect[s].SizeByte
			s++
		}
	}
	t.Data = make([]byte, sectorSize)
	for i := 0; i < len(t.Data); i++ {
		t.Data[i] = 0xE5
	}
	if len(d.Tracks) < int(track+1) {
		d.Tracks = append(d.Tracks, t)
		d.Entry.NbTracks++
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
		offset := make([]byte, (0x100 - (52 + uint(d.Entry.NbHeads*d.Entry.NbTracks))))
		if err := binary.Write(w, binary.LittleEndian, &offset); err != nil {
			fmt.Fprintf(os.Stderr, "Cannot write CPCEmuEnt padding 0x100 error :%v\n", err)
			return err
		}
	}
	var i uint8
	for i = 0; i < d.Entry.NbTracks; i++ {
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
	// if d.Entry.NbHeads == 1 {
	minSectFirst := d.GetMinSect()
	if minSectFirst != 0x41 && minSectFirst != 0xc1 && minSectFirst != 0x01 {
		fmt.Fprintf(os.Stderr, "Bad sector %.2x\n", minSectFirst)
		return ErrorBadSectorNumber
	}
	if d.Entry.NbTracks > 80 {
		d.Entry.NbTracks = 80
	}
	var track uint8
	for track = 0; track < d.Entry.NbTracks; track++ {
		tr := d.Tracks[track]
		if !d.Extended {
			if tr.NbSect != 9 {
				fmt.Fprintf(os.Stderr, "Warning : track :%d has %d sectors ! wanted 9\n", track, tr.NbSect)
			}
		}
		var minSect, maxSect, s uint8
		minSect = 0xFF
		maxSect = 0
		nbSecteur := int(tr.NbSect)
		if nbSecteur > len(tr.Sect) {
			nbSecteur = len(tr.Sect)
		}
		for s = 0; s < uint8(nbSecteur); s++ {
			if minSect > tr.Sect[s].R {
				minSect = tr.Sect[s].R
			}
			if maxSect < tr.Sect[s].R {
				maxSect = tr.Sect[s].R
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

// Recherche le plus petit secteur d'une piste
func (d *DSK) GetMinSect() uint8 {
	var Sect uint8 = 0xFF
	var s uint8
	tr := d.Tracks[0]
	// fmt.Fprintf(os.Stdout, "Track 0 nbSect :%d \n", tr.NbSect)
	for s = 0; s < tr.NbSect; s++ {
		//	fmt.Fprintf(os.Stdout, "Sector %d, R %d\n", s, tr.Sect[s].R)
		if Sect > tr.Sect[s].R {
			Sect = tr.Sect[s].R
		}
	}
	return Sect
}

// Retourne la position d'un secteur dans le fichier DSK, position dans la structure Data
func (d *DSK) GetPosData(track, sect uint8, SectPhysique bool) uint16 {
	// Recherche position secteur
	tr := d.Tracks[track]
	var SizeByte uint16
	var Pos uint16
	var s uint8
	// Pos += 256

	// fmt.Fprintf(os.Stdout,"Track:%d,Secteur:%d\n",track,sect)

	// Pos += 256
	for s = 0; s < tr.NbSect; s++ {
		if (tr.Sect[s].R == sect && SectPhysique) || (s == sect && !SectPhysique) {
			break
		}
		SizeByte = tr.Sect[s].SizeByte
		if SizeByte != 0 {
			Pos += SizeByte
		} else {
			Pos += (128 << tr.Sect[s].N)
		}
		// fmt.Fprintf(os.Stderr, "sizebyte:%d, t:%d,s:%d,tr.Sect[s].SizeByte:%d, tr->Sect[ s ].N:%d, Pos:%d\n", SizeByte,t,s,tr.Sect[s].SizeByte,tr.Sect[ s ].N,Pos)
	}

	return Pos
}

func (d *DSK) GetFile(path string, indice int) error {
	i := indice
	current := make([]byte, 16)
	nomIndice := make([]byte, 16)
	lMax := 0x1000000
	cumul := 0
	err := d.GetCatalogue()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error while getting the catalogue, error :%v\n", err)
	}
	copy(nomIndice, d.Catalogue[i].Nom[:])
	copy(nomIndice, d.Catalogue[i].Ext[:])
	fw, err := os.Create(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot open file (%s), error :%v\n", path, err)
		return err
	}
	defer fw.Close()
	for {
		// Longueur du fichier
		var j uint8
		l := (d.Catalogue[i].NbPages + 7) >> 3
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
		copy(current, d.Catalogue[i].Nom[:])
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
	err := d.GetCatalogue()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error while getting the catalogue, error :%v\n", err)
	}
	// var size uint16
	// size = len(d.Tracks) *
	// for _, t := range d.Tracks {
	// 	for _, s := range t.Sect {
	// 		size += uint16(s.N)
	// 	}
	// 	//	size /= uint16(t.NbSect)
	// }
	return d.Entry.DataSize * uint16(d.Entry.NbTracks)
}

func GetNomAmsdos(masque string) string {
	amsdosFile := make([]byte, 12)
	for i := 0; i < 12; i++ {
		amsdosFile[i] = ' '
	}
	file := strings.ToUpper(strings.TrimSuffix(filepath.Base(masque), filepath.Ext(filepath.Base(masque))))
	filenameSize := len(file)
	if filenameSize > 8 {
		filenameSize = 8
	}
	copy(amsdosFile[0:filenameSize], file[0:filenameSize])
	amsdosFile[8] = '.'
	ext := strings.ToUpper(filepath.Ext(masque))
	copy(amsdosFile[9:12], ext[1:])
	return string(amsdosFile)
}

func (d *DSK) PutFile(masque string, typeModeImport uint8, loadAddress, exeAddress, userNumber uint16, isSystemFile, readOnly bool) error {
	buff := make([]byte, 0x20000)
	cFileName := GetNomAmsdos(masque)
	header := &amsdos.StAmsdos{}
	var addHeader bool
	var err error

	err = d.GetCatalogue()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error while getting the catalogue, error :%v\n", err)
	}
	fr, err := os.Open(masque)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot read file (%s) error :%v\n", masque, err)
		return err
	}
	fileLength, err := fr.Read(buff)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot read the content of the file (%s) with error %v\n", masque, err)
		return err
	}
	fmt.Fprintf(os.Stderr, "file (%s) read (%d bytes).\n", masque, fileLength)
	_, err = fr.Seek(0, io.SeekStart)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error while seeking in file error :%v\n", err)
	}

	if err = binary.Read(fr, binary.LittleEndian, header); err != nil {
		fmt.Fprintf(os.Stderr, "No header found for file :%s, error :%v\n", masque, err)
	}

	if typeModeImport == MODE_ASCII && fileLength%128 != 0 {
		buff[fileLength] = 0x1A
	}

	if typeModeImport == MODE_PROTECTED && fileLength%128 != 0 {
		buff[fileLength] = 0x1A
	}

	var isAmsdos bool
	//
	// Regarde si le fichier contient une en-tete ou non
	//
	if err == nil && header.Checksum == header.ComputedChecksum16() {
		isAmsdos = true
	}
	if !isAmsdos {
		// Creer une en-tete amsdos par defaut
		fmt.Fprintf(os.Stderr, "Create header... (%s)\n", masque)
		header = &amsdos.StAmsdos{}
		header.User = byte(userNumber)
		header.Size = uint16(fileLength)
		header.Size2 = uint16(fileLength)
		header.LogicalSize = uint16(fileLength)
		copy(header.Filename[:], []byte(cFileName[0:12]))
		header.Address = loadAddress
		if loadAddress != 0 {
			typeModeImport = MODE_BINAIRE
		}
		header.Exec = exeAddress
		if exeAddress != 0 || loadAddress != 0 {
			typeModeImport = MODE_BINAIRE
		}
		header.Type = typeModeImport

		// Il faut recalculer le checksum en comptant es adresses !
		header.Checksum = header.ComputedChecksum16()

	} else {
		fmt.Fprintf(os.Stderr, "File has already header...(%s)\n", masque)
	}
	//
	// En fonction du mode d'importation...
	//
	switch typeModeImport {
	case MODE_ASCII:
		//
		// Importation en mode ASCII
		//
		if isAmsdos {
			// Supprmier en-tete si elle existe
			fmt.Fprintf(os.Stderr, "Removing header...(%s)\n", masque)
			copy(buff[0:], buff[binary.Size(amsdos.StAmsdos{}):])
		}
	case MODE_BINAIRE:
		//
		// Importation en mode BINAIRE
		//

		if !isAmsdos {
			//
			// Indique qu'il faudra ajouter une en-tete
			//
			addHeader = true
		}
	}
	//
	// Si fichier ok pour etre import
	//
	if addHeader {
		// Ajoute l'en-tete amsdos si necessaire

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
		//	memmove( &Buff[ sizeof( StAmsdos ) ], Buff, Lg );
		//         	memcpy( Buff, e, sizeof( StAmsdos ) );
		//       	Lg += sizeof( StAmsdos );
	}
	if fileLength > 65536 {
		return ErrorFileSizeExceed
	}
	// if (MODE_BINAIRE) ClearAmsdos(Buff); //Remplace les octets inutilises par des 0 dans l'en-tete
	return d.CopyFile(buff, cFileName, uint16(fileLength), 256, userNumber, isSystemFile, readOnly)
}

// Copie un fichier sur le DSK
//
// la taille est determine par le nombre de NbPages
// regarder pourquoi different d'une autre DSK
func (d *DSK) CopyFile(bufFile []byte, fileName string, fileLength, maxBloc, userNumber uint16, isSystemFile, readOnly bool) error {
	var nbPages, taillePage int
	d.FillBitmap()
	dirLoc := d.GetNomDir(fileName)
	var posFile uint16                       // Construit l'entree pour mettre dans le catalogue
	for posFile = 0; posFile < fileLength; { // Pour chaque bloc du fichier
		posDir, err := d.RechercheDirLibre() // Trouve une entree libre dans le CAT
		if err == nil {
			dirLoc.User = uint8(userNumber) // Remplit l'entree : User 0
			if isSystemFile {
				dirLoc.Ext[0] |= 0x80
			}
			if readOnly {
				dirLoc.Ext[0] |= 0x80
			}
			dirLoc.NumPage = uint8(nbPages) // Numero de l'entree dans le fichier
			nbPages++
			taillePage = ((int(fileLength) - int(posFile) + 127) >> 7) // Taille de la page (on arrondit par le haut)
			if taillePage > 128 {                                      // Si y'a plus de 16k il faut plusieurs pages
				taillePage = 128
			}

			dirLoc.NbPages = uint8(taillePage)
			l := (dirLoc.NbPages + 7) >> 3 // Nombre de blocs=TaillePage/8 arrondi par le haut
			for i := 0; i < 16; i++ {
				dirLoc.Blocks[i] = 0
			}
			var j uint8
			for j = 0; j < l; j++ { // Pour chaque bloc de la page
				bloc := d.RechercheBlocLibre(int(maxBloc)) // Met le fichier sur la disquette
				//	fmt.Fprintf(os.Stdout,"Bloc:%d, MaxBloc:%d\n",bloc,maxBloc)
				if bloc != 0 {
					dirLoc.Blocks[j] = bloc
					err = d.WriteBloc(int(bloc), bufFile, posFile)
					if err != nil {
						fmt.Fprintf(os.Stdout, "error while writing bloc %v\n", err)
					}
					posFile += 1024 // Passe au bloc suivant

				} else {
					return ErrorNoBloc
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

func (d *DSK) FillBitmap() int {
	for i := 0; i < len(d.BitMap); i++ {
		d.BitMap[i] = 0
	}
	d.BitMap[0] = 1
	d.BitMap[1] = 1
	var nbKo int
	for i := 0; i < 64; i++ {
		dir, _ := d.GetInfoDirEntry(uint8(i))
		if dir.User != USER_DELETED {
			for j := 0; j < 16; j++ {
				b := dir.Blocks[j]
				if b > 1 && d.BitMap[b] != 1 {
					d.BitMap[b] = 1
					nbKo++
				}
			}
		}
	}
	return nbKo
}

func (d *DSK) GetNomDir(nomFile string) StDirEntry {
	e := StDirEntry{}
	for i := 0; i < 8; i++ {
		e.Nom[i] = ' '
	}

	for i := 0; i < 3; i++ {
		e.Ext[i] = ' '
	}
	copy(e.Ext[:], []byte(nomFile[9:12]))
	copy(e.Nom[:], []byte(nomFile[0:8]))
	return e
}

func (d *DSK) CopyRawFile(bufFile []byte, fileLength uint16, track, sector int) (int, int, error) {
	d.FillBitmap()

	var posFile uint16 // Construit l'entree pour mettre dans le catalogue
	var err error
	var written int
	for posFile = 0; posFile < fileLength; { // Pour chaque bloc du fichier
		track, sector, written, err = d.WriteAtTrackSector(track, sector, bufFile, posFile)
		if err != nil {
			return track, sector, err
		}
		posFile += uint16(written) // Passe à la position suivante
	}
	return track, sector, nil
}

func (d *DSK) WriteAtTrackSector(track int, sect int, bufBloc []byte, offset uint16) (int, int, int, error) {
	var dataWritten int
	minSect := d.GetMinSect()
	//
	// Ajuste le nombre de pistes si depassement capacite
	//
	if track > int(d.Entry.NbTracks-1) {
		if d.Entry.NbHeads == 1 {
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
	sectorSize := uint16(d.Tracks[track].Sect[sect].SizeByte)
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
	if track > int(d.Entry.NbTracks-1) {
		if d.Entry.NbHeads == 1 {
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
	sectorSize = uint16(d.Tracks[track].Sect[sect].SizeByte)
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
	minSect := d.GetMinSect()
	if minSect == 0x41 {
		track += 2
	} else {
		if minSect == 0x01 {
			track++
		}
	}
	//
	// Ajuste le nombre de pistes si depassement capacite
	//
	if track > int(d.Entry.NbTracks-1) {
		if d.Entry.NbHeads == 1 {
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
	if track > int(d.Entry.NbTracks-1) {
		if d.Entry.NbHeads == 1 {
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
	d.FillBitmap()
	content := make([]byte, 0)
	var posFile uint16 // Construit l'entree pour mettre dans le catalogue
	var buf []byte

	for posFile = 0; posFile < fileLength; { // Pour chaque bloc du fichier
		track, sector, buf = d.ReadAtTrackSector(track, sector)
		posFile += uint16(len(buf)) // Passe à la position suivante
		content = append(content, buf...)
	}
	return track, sector, content
}

func (d *DSK) ReadAtTrackSector(track, sect int) (int, int, []byte) {
	minSect := d.GetMinSect()
	if minSect == 0x41 {
		track += 2
	} else {
		if minSect == 0x01 {
			track++
		}
	}
	sectorSize := uint16(d.Tracks[track].Sect[sect].SizeByte)
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
	sectorSize = uint16(d.Tracks[track].Sect[sect].SizeByte)

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
	minSect := d.GetMinSect()
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

//
// Recherche un bloc libre et le remplit
//

func (d *DSK) RechercheBlocLibre(maxBloc int) uint8 {
	for i := 2; i < maxBloc; i++ {
		if d.BitMap[i] == 0 {
			d.BitMap[i] = 1
			return uint8(i)
		}
	}
	return 0
}

//
// Recherche une entree de repertoire libre
//

func (d *DSK) RechercheDirLibre() (uint8, error) {
	for i := 0; i < 64; i++ {
		dir, _ := d.GetInfoDirEntry(uint8(i))
		if dir.User == USER_DELETED {
			return uint8(i), nil
		}
	}
	return 0, ErrorNoDirEntry
}

func (d *DSK) DisplayCatalogue() {
	err := d.GetCatalogue()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error while getting the catalogue error :%v\n", err)
	}
	for i := 0; i < 64; i++ {
		entry := d.Catalogue[i]
		if entry.User != USER_DELETED && entry.NumPage != 0 {
			fmt.Fprintf(os.Stderr, "%s.%s : %d\n", entry.Nom, entry.Ext, entry.User)
		}
	}
}

func (d *DSK) GetEntryyNameInCatalogue(num int) string {
	err := d.GetCatalogue()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error while getting the catalogue error :%v\n", err)
	}
	var nom string
	for i := 0; i < 64; i++ {
		entry := d.Catalogue[i]
		if entry.User != USER_DELETED && entry.NumPage != 0 && i == num {
			nom = fmt.Sprintf("%.8s.%.3s", entry.Nom, entry.Ext)
			//	fmt.Fprintf(os.Stdout,"%s.%s : %d\n",entry.Nom,entry.Ext,entry.User )
			return nom
		}
	}
	return nom
}

func (d *DSK) GetEntrySizeInCatalogue(num int) string {
	err := d.GetCatalogue()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error while getting the catalogue error :%v\n", err)
	}
	for i := 0; i < 64; i++ {
		entry := d.Catalogue[i]
		if entry.User != USER_DELETED && entry.NumPage != 0 && i == num {
			var p, t int
			for {
				if d.Catalogue[p+i].User == entry.User {
					t += int(d.Catalogue[p+i].NbPages)
				}
				p++
				if d.Catalogue[p+i].NumPage != 0 || p+i >= 64 {
					break
				}
			}
			return fmt.Sprintf("%d ko", (t+7)>>3)
		}
	}
	return ""
}

func (d *DSK) GetFilesize(s StDirEntry) int {
	t := 0
	for i := 0; i < 64; i++ {
		if d.Catalogue[i].User != USER_DELETED {
			if d.Catalogue[i].Nom == s.Nom &&
				d.Catalogue[i].Ext == s.Ext {
				t += int(d.Catalogue[i].NbPages)
			}
		}
	}
	return (t + 7) >> 3
}

func (d *DSK) GetFilesIndices() []int {
	indices := make([]int, 0)
	cache := make(map[string]bool)
	for i := 0; i < 64; i++ {
		if d.Catalogue[i].User != USER_DELETED {
			filename := fmt.Sprintf("%s.%s", d.Catalogue[i].Nom, d.Catalogue[i].Ext)
			if !cache[filename] {
				cache[filename] = true
				indices = append(indices, i)
			}
		}
	}

	return indices
}

func (d *DSK) GetCatalogue() error {
	if d.catalogueLoaded {
		return nil
	}
	for i := 0; i < 64; i++ {
		dirEntry, err := d.GetInfoDirEntry(uint8(i))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error while reading catalogue error :%v\n", err)
		}
		d.Catalogue[i] = dirEntry
	}
	d.catalogueLoaded = true
	return nil
}

func (d *DSK) SetInfoDirEntry(numDir uint8, e StDirEntry) error {
	minSect := d.GetMinSect()
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
		//	fmt.Fprintf(os.Stdout, "t:%d,s:%d,pos:%d\n", t, s, pos)
		//	fmt.Fprintf(os.Stdout,"offset:%d\n",((uint16(numDir)&15)<<5) + d.GetPosData(t, s, true))
		copy(d.Tracks[t].Data[((uint16(numDir)&15)<<5)+pos:((uint16(numDir)&15)<<5)+pos+uint16(binary.Size(entry))], entry[:])
	}
	return nil
}

func (d *DSK) GetInfoDirEntry(numDir uint8) (StDirEntry, error) {
	dir := StDirEntry{}
	minSect := d.GetMinSect()
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

func (d *DSK) GetType(langue int, ams *amsdos.StAmsdos) string {
	if ams.Checksum == ams.ComputedChecksum16() {
		switch ams.Type {
		case 0:
			return "BASIC"
		case 1:
			return "BASIC(P)"
		case 2:
			return "BINAIRE"
		case 3:
			return "BINAIRE(P)"
		default:
			return "INCONNU"

		}
	}
	return "ASCII"
}

func (d *DSK) FileExists(entry StDirEntry) int {
	for i := 0; i < 64; i++ {
		dir, err := d.GetInfoDirEntry(uint8(i))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error while getting info dir entry (%d) error :%v\n", i, err)
		} else {
			for q := 0; q < 8; q++ {
				dir.Nom[q] &= 127
			}
			for q := 0; q < 3; q++ {
				dir.Ext[q] &= 127
			}
			if dir.User != USER_DELETED && dir.Nom == entry.Nom && dir.Ext == entry.Ext {
				return i
			}
		}

	}
	return NOT_FOUND
}

func (d *DSK) GetFileIn(filename string, indice int) ([]byte, error) {
	i := indice
	lMax := 0x1000000
	b := make([]byte, 0)
	firstBlock := true
	/*	tabDir := make([]StDirEntry, 64)
		for j := 0; j < 64; j++ {
			tabDir[j], _ = d.GetInfoDirEntry(uint8(j))
		}*/
	err := d.GetCatalogue()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error while getting the catalogue error :%v\n", err)
	}
	entryIndice := d.Catalogue[i]
	var cumul, tailleFichier int
	var isAmsdos bool
	for {
		l := (d.Catalogue[i].NbPages + 7) >> 3
		for j := 0; uint8(j) < l; j++ {
			tailleBloc := 1024
			bloc := d.ReadBloc(int(d.Catalogue[i].Blocks[j]))
			if firstBlock {
				var header *cpc.CpcHead
				isAmsdos, header = amsdos.CheckAmsdos(bloc)
				if isAmsdos {
					tailleFichier = int(header.Size) + 0x80
				}
				firstBlock = false
			}
			var nbOctets int
			if lMax > tailleBloc {
				nbOctets = tailleBloc
			} else {
				nbOctets = lMax
			}
			if nbOctets > 0 {
				b = append(b, bloc...)
				cumul += nbOctets
			}
			lMax -= 1024
		}
		i++
		if i >= 64 {
			return b, errors.New("cannot get the file, Exceed catalogue indice")
		}
		if entryIndice.Nom != d.Catalogue[i].Nom || entryIndice.Ext != d.Catalogue[i].Ext {
			break
		}
	}
	if tailleFichier <= 0 || tailleFichier <= cumul {
		tailleFichier = cumul
	}

	/*if !isAmsdos {
		keepOn := true
		for i = tailleFichier - 1; i >= 0; i-- {
			if b[i] == 0 {
				tailleFichier--
			} else {
				keepOn = false
			}
			if !keepOn {
				break
			}
		}
	}*/

	return b[0:tailleFichier], nil
}

func (d *DSK) ViewFile(indice int) ([]byte, int, error) {
	i := indice
	lMax := 0x1000000
	b := make([]byte, 0)
	firstBlock := true
	err := d.GetCatalogue()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error while getting the catalogue error :%v\n", err)
	}
	entryIndice := d.Catalogue[i]
	var tailleFichier, cumul int
	for {
		l := (d.Catalogue[i].NbPages + 7) >> 3
		var j uint8
		for j = 0; j < l; j++ {
			tailleBloc := 1024
			bloc := d.ReadBloc(int(d.Catalogue[i].Blocks[j]))
			if firstBlock {
				isAmsdos, header := amsdos.CheckAmsdos(bloc)
				if isAmsdos {
					t := make([]byte, len(bloc))
					copy(t, bloc[HeaderSize:])
					bloc = t
					tailleBloc -= HeaderSize
					tailleFichier = int(header.Size)
				}
				firstBlock = false
			}
			var nbOctets int
			if lMax > tailleBloc {
				nbOctets = tailleBloc
			} else {
				nbOctets = lMax
			}
			if nbOctets > 0 {
				b = append(b, bloc...)
				cumul += nbOctets
			}
			lMax -= 1024
		}
		i++
		if i >= 64 {
			return b, cumul, errors.New("cannot get the file, Exceed catalogue indice")
		}
		if entryIndice.Nom != d.Catalogue[i].Nom || entryIndice.Ext != d.Catalogue[i].Ext {
			break
		}
	}
	if tailleFichier == 0 {
		tailleFichier = cumul
	}
	for i := len(b) - 1; i >= 0; i-- {
		if b[i] == 0xE5 {
			tailleFichier = i
		} else {
			break
		}
	}
	return b, tailleFichier, nil
}

func (d *DSK) RemoveFile(indice uint8) error {
	err := d.GetCatalogue()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error while getting the catalogue error :%v\n", err)
	}
	entryIndice := d.Catalogue[indice]

	for {

		entry, err := d.GetInfoDirEntry(indice)
		if err != nil {
			return ErrorNoDirEntry
		}
		if entryIndice.Nom != d.Catalogue[indice].Nom || entryIndice.Ext != d.Catalogue[indice].Ext {
			break
		}
		d.Catalogue[indice].User = USER_DELETED
		entry.User = USER_DELETED
		if err := d.SetInfoDirEntry(indice, entry); err != nil {
			return ErrorNoDirEntry
		}
		indice++

	}
	return nil
}
