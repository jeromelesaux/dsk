package dsk

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/jeromelesaux/m4client/cpc"
	"io"
	"os"
)

var USER_DELETED = 0xE5
var SECTSIZE = 512
var ErrorUnsupportedDskFormat = errors.New("Unsupported DSK Format.")
var ErrorUnsupportedMultiHeadDsk = errors.New("Multi-side dsk ! Expected 1 head")
var ErrorBadSectorNumber = errors.New("DSK has wrong sector number!")

type StAmsdos = cpc.CpcHead

type CPCEMUEnt struct {
	Debut    [0x30]byte // "MV - CPCEMU Disk-File\r\nDisk-Info\r\n"
	NbTracks uint8
	NbHeads  uint8
	DataSize uint16 // 0x1300 = 256 + ( 512 * nbsecteurs )
	Unused   [0xCC]byte
}

func (e *CPCEMUEnt) ToString() string {
	return fmt.Sprintf("Debut:%s, nbTracks:%d, nbHeads:%d, DataSize:%d",
		e.Debut, e.NbTracks, e.NbHeads, e.DataSize)
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

func ( c *CPCEMUSect)ToString() string {
	return fmt.Sprintf("C:%d,H:%d,R:%d,N:%d,Un1:%d:SizeByte:%d", //,DataSize:%d",
	c.C,c.H,c.R,c.N,c.Un1,c.SizeByte ) //, len(c.Data))
}

func (c *CPCEMUSect)Write(w io.Writer) error {
	if err := binary.Write(w, binary.LittleEndian, &c.C); err != nil {
		fmt.Fprintf(os.Stderr, "Error while reading CPCEmuSect.C error :%v\n", err)
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, &c.H); err != nil {
		fmt.Fprintf(os.Stderr, "Error while reading CPCEmuSect.H error :%v\n", err)
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, &c.R); err != nil {
		fmt.Fprintf(os.Stderr, "Error while reading CPCEmuSect.R error :%v\n", err)
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, &c.N); err != nil {
		fmt.Fprintf(os.Stderr, "Error while reading CPCEmuSect.N error :%v\n", err)
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, &c.Un1); err != nil {
		fmt.Fprintf(os.Stderr, "Error while reading CPCEmuSect.Un1 error :%v\n", err)
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, &c.SizeByte); err != nil {
		fmt.Fprintf(os.Stderr, "Error while reading CPCEmuSect.SizeByte error :%v\n", err)
		return err
	}
/*	if err := binary.Write(w, binary.LittleEndian, &c.Data); err != nil {
		fmt.Fprintf(os.Stderr, "Error while reading CPCEmuSect.Data error :%v\n", err)
		return err
	} */ 
	fmt.Fprintf(os.Stdout,"Sector %s\n",c.ToString())
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
	
/*	c.Data = make([]byte, c.SizeByte)
	fmt.Fprintf(os.Stdout,"Sector:%s\n",c.ToString())
	nb, err := r.Read(c.Data)
	fmt.Fprintf(os.Stdout,"%d bytes read\n",nb)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error while reading CPCEmuSect.Data error :%v\n", err)
		return err
	} */
	fmt.Fprintf(os.Stdout,"Sector %s\n",c.ToString())
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
	
	fmt.Fprintf(os.Stdout,"Track:%s\n",c.ToString())
	var i uint8
	for i = 0; i < c.NbSect; i++ {
		sect := &CPCEMUSect{}
		if err := sect.Read(r); err != nil {
			fmt.Fprintf(os.Stderr, "Cannot read sector (%d) error :%v\n", i, err)
			return err
		}
		c.Sect[i] = *sect
	}
	for i = c.NbSect; i< 29 ; i++ {
		sect := &CPCEMUSect{}
		sect.Read(r)
	}
	c.Data = make([]byte,0x200*uint16(c.NbSect))
	if err := binary.Read(r, binary.LittleEndian, &c.Data); err != nil {
		fmt.Fprintf(os.Stderr, "Error while reading CPCEmuSect.Data error :%v\n", err)
		return err
	} 
	return nil
}


func (c *CPCEMUTrack) Write(w io.Writer) error {
	if err := binary.Write(w, binary.LittleEndian, &c.ID); err != nil {
		fmt.Fprintf(os.Stderr, "Error while reading CPCEMUTrack.ID error :%v\n", err)
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, &c.Track); err != nil {
		fmt.Fprintf(os.Stderr, "Error while reading CPCEMUTrack.Track error :%v\n", err)
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, &c.Head); err != nil {
		fmt.Fprintf(os.Stderr, "Error while reading CPCEMUTrack.Head error :%v\n", err)
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, &c.Unused); err != nil {
		fmt.Fprintf(os.Stderr, "Error while reading CPCEMUTrack.Unused error :%v\n", err)
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, &c.SectSize); err != nil {
		fmt.Fprintf(os.Stderr, "Error while reading CPCEMUTrack.SectSize error :%v\n", err)
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, &c.NbSect); err != nil {
		fmt.Fprintf(os.Stderr, "Error while reading CPCEMUTrack.NbSect error :%v\n", err)
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, &c.Gap3); err != nil {
		fmt.Fprintf(os.Stderr, "Error while reading CPCEMUTrack.Gap3 error :%v\n", err)
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, &c.OctRemp); err != nil {
		fmt.Fprintf(os.Stderr, "Error while reading CPCEMUTrack.OctRemp error :%v\n", err)
		return err
	}
	
	fmt.Fprintf(os.Stdout,"Track:%s\n",c.ToString())
	var i uint8
	for i = 0; i < c.NbSect; i++ {
		if err := c.Sect[i].Write(w); err != nil {
			fmt.Fprintf(os.Stderr, "Cannot read sector (%d) error :%v\n", i, err)
			return err
		}
	}
	for i = c.NbSect; i < 29 ; i++ {
		sect := &CPCEMUSect{}
		sect.Write(w)
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
	Entry  CPCEMUEnt
	Tracks []CPCEMUTrack
	BitMap [256]byte
}

func (d *DSK)Read(r io.Reader) error {
	if err := binary.Read(r, binary.LittleEndian, &d.Entry); err != nil {
		fmt.Fprintf(os.Stderr, "Cannot read CPCEmuEnt error :%v\n", err)
		return  err
	}
	mv := make([]byte, 4)
	extended := make([]byte, 16)
	copy(mv, d.Entry.Debut[0:4])
	copy(extended, d.Entry.Debut[0:16])
	if string(mv) != "MV -" && string(extended) != "EXTENDED CPC DSK" {
		return ErrorUnsupportedDskFormat
	}
	d.Tracks = make([]CPCEMUTrack,d.Entry.NbTracks)
	var i uint8
	for i = 0; i < d.Entry.NbTracks; i++ {
		//	fmt.Fprintf(os.Stdout,"Loading track %d, total: %d\n", i, cpcEntry.NbTracks)
		track := &CPCEMUTrack{}
		if err := track.Read(r); err != nil {
			fmt.Fprintf(os.Stderr, "Error track (%d) error :%v\n", i, err)
		} 
		d.Tracks[i] = *track
		fmt.Fprintf(os.Stdout, "Track %d %s\n", i, d.Tracks[i].ToString())
	}
	return nil
}

func FormatDsk(nbSect, nbTrack uint8) *DSK {
	dsk := &DSK{}
	entry := CPCEMUEnt{}
	copy(entry.Debut[:], "MV - CPCEMU Disk-File\r\nDisk-Info\r\n")
	entry.DataSize = uint16(0x21c * int(nbSect))
	entry.NbTracks = nbTrack
	entry.NbHeads = 1
	dsk.Entry = entry
	dsk.Tracks = make([]CPCEMUTrack, nbTrack)
	var i uint8
	for i = 0; i < nbTrack; i++ {
		dsk.FormatTrack(i, 0xC1, nbSect)

	}
	return dsk
}

func (d *DSK) FormatTrack(track, minSect, nbSect uint8) {
	t := CPCEMUTrack{}
	copy(t.ID[:], "Track-Info\r\n")
	t.Track = track
	t.Head = 0
	t.SectSize = 2
	t.NbSect = nbSect
	t.Gap3 = 0x4E
	t.OctRemp = 0xE5
	//
	// Gestion "entrelacement" des secteurs
	//
	var s uint8
	var ss uint8
	for s = 0; s < nbSect; s++ {
		t.Sect[s].C = track
		t.Sect[s].H = 0
		t.Sect[s].R = (ss + minSect)
		t.Sect[s].N = 2
		t.Sect[s].SizeByte = 0x200
	/*	var i uint16
		t.Sect[s].Data = make([]byte, t.Sect[s].SizeByte)
		for i = 0; i < t.Sect[s].SizeByte; i++ {
			t.Sect[s].Data[i] = 0xe5
		} */
		ss++
	}
	t.Data = make([]byte,0x200*uint16(nbSect))
	for i := 0; i < len(t.Data); i++ {
		t.Data[i] = 0xE5
	}
	d.Tracks[track] = t
}

func (d *DSK) Write(w io.Writer) error {
	if err := binary.Write(w, binary.LittleEndian, &d.Entry); err != nil {
		fmt.Fprintf(os.Stderr, "Cannot read CPCEmuEnt error :%v\n", err)
		return  err
	}
	var i uint8
	for i = 0; i < d.Entry.NbTracks; i++ {
		if err := d.Tracks[i].Write(w); err != nil {
			fmt.Fprintf(os.Stderr, "Error track (%d) error :%v\n", i, err)
		} 
	}
	return nil
}

func WriteDsk(filePath string, d *DSK) (error) {
	f, err := os.Create(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot open file (%s) error %v\n", filePath, err)
		return err
	}
	if err :=d.Write(f); err != nil {
		fmt.Fprintf(os.Stderr,"Error while reading (%s) error %v", filePath, err)
	}	
	f.Close()
	return nil
}

func NewDsk(filePath string) (*DSK, error) {
	f, err := os.Open(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot open file (%s) error %v\n", filePath, err)
		return &DSK{}, err
	}
	
	dsk := &DSK{}
	if err :=dsk.Read(f); err != nil {
		fmt.Fprintf(os.Stderr,"Error while reading (%s) error %v", filePath, err)
	}	
	f.Close()
	return dsk, nil
}

func (d *DSK) CheckDsk() error {
	if d.Entry.NbHeads == 1 {
		minSectFirst := d.GetMinSect()
		if minSectFirst != 0x41 && minSectFirst != 0xc1 && minSectFirst != 0x01 {
			fmt.Fprintf(os.Stderr, "Bad sector %.2x\n", minSectFirst)
			return ErrorBadSectorNumber
		}
		if d.Entry.NbTracks > 42 {
			d.Entry.NbTracks = 42
		}
		var track uint8
		for track = 0; track < d.Entry.NbTracks; track++ {
			tr := d.Tracks[track]
			if tr.NbSect != 9 {
				fmt.Fprintf(os.Stdout, "Warning : track :%d has %d sectors ! wanted 9\n", track, tr.NbSect)
			}
			var minSect, maxSect, s uint8
			minSect = 0xFF
			maxSect = 0
			for s = 0; s < tr.NbSect; s++ {
				if minSect > tr.Sect[s].R {
					minSect = tr.Sect[s].R
				}
				if maxSect < tr.Sect[s].R {
					maxSect = tr.Sect[s].R
				}
			}
			if maxSect-minSect != 8 {
				fmt.Fprintf(os.Stdout, "Warning : strange sector numbering in track %d!\n", track)
			}
			if minSect != minSectFirst {
				fmt.Fprintf(os.Stdout, "Warning : track %d start at sector %d while track 0 starts at %d\n", track, minSect, minSectFirst)
			}
		}
		return nil
	}
	return ErrorUnsupportedMultiHeadDsk
}

//
// Recherche le plus petit secteur d'une piste
//
func (d *DSK) GetMinSect() uint8 {
	var Sect uint8 = 0xFF
	var s uint8
	tr := d.Tracks[0]
	fmt.Fprintf(os.Stdout, "Track 0 nbSect :%d \n", tr.NbSect)
	for s = 0; s < tr.NbSect; s++ {
		fmt.Fprintf(os.Stdout, "Sector %d, R %d\n", s, tr.Sect[s].R)
		if Sect > tr.Sect[s].R {
			Sect = tr.Sect[s].R
		}
	}
	return Sect
}

//
// Retourne la position d'un secteur dans le fichier DSK
//
func (d *DSK) GetPosData(track, sect uint8, SectPhysique bool) uint16 {
	// Recherche position secteur
	var tr CPCEMUTrack = d.Tracks[track]
	var SizeByte uint16
	var Pos uint16

	if (tr.Sect[sect].R == sect && SectPhysique) || !SectPhysique {
		SizeByte = tr.Sect[sect].SizeByte
		if SizeByte == 0 {
			Pos += SizeByte
		} else {
			Pos += (128 << tr.Sect[sect].N)
		}
	}
	return Pos
}

//
// Recherche un bloc libre et le remplit
//
/*
func (d *DSK) RechercheBlocLibre(MaxBloc uint8) uint8 {
	var i uint8
	for i = 2; i < MaxBloc; i++ {
		if d.Bitmap[i] == 0 {
			d.Bitmap[i] = 1
			return i
		}
	}
	return 0
}
*/
//
// Recherche une entr�e de r�pertoire libre
//
/*
func (d *DSK) RechercheDirLibre() uint8 {
	for i := 0; i < 64; i++ {
		Dir = GetInfoDirEntry(i)
		if Dir.User == USER_DELETED {
			return i
		}
	}
	return -1
}
*/
