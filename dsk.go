package dsk

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/jeromelesaux/m4client/cpc"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var USER_DELETED uint8 = 0xE5
var SECTSIZE uint16 = 512
var (
	ErrorUnsupportedDskFormat    = errors.New("Unsupported DSK Format.")
	ErrorUnsupportedMultiHeadDsk = errors.New("Multi-side dsk ! Expected 1 head")
	ErrorBadSectorNumber         = errors.New("DSK has wrong sector number!")
	ErrorCatalogueExceed         = errors.New("Catalogue indice exceed.")
	ErrorNoBloc                  = errors.New("Error no more block available.")
	ErrorNoDirEntry              = errors.New("Error no more dir entry available.")
	ErrorFileSizeExceed          = errors.New("Filesize exceed.")
)
var (
	MODE_ASCII   uint8 = 0
	MODE_BINAIRE uint8 = 1
)

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

func (c *CPCEMUSect) ToString() string {
	return fmt.Sprintf("C:%d,H:%d,R:%d,N:%d,Un1:%d:SizeByte:%d", //,DataSize:%d",
		c.C, c.H, c.R, c.N, c.Un1, c.SizeByte) //, len(c.Data))
}

func (c *CPCEMUSect) Write(w io.Writer) error {
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
	for i = 0; i < c.NbSect; i++ {
		sect := &CPCEMUSect{}
		if err := sect.Read(r); err != nil {
			fmt.Fprintf(os.Stderr, "Cannot read sector (%d) error :%v\n", i, err)
			return err
		}
		c.Sect[i] = *sect
	}
	for i = c.NbSect; i < 29; i++ {
		sect := &CPCEMUSect{}
		sect.Read(r)
	}
	c.Data = make([]byte, 0x200*uint16(c.NbSect))
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
	Entry           CPCEMUEnt
	Tracks          []CPCEMUTrack
	BitMap          [256]byte
	Catalogue       [64]StDirEntry
	catalogueLoaded bool
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
	d.Tracks = make([]CPCEMUTrack, d.Entry.NbTracks)
	var i uint8
	for i = 0; i < d.Entry.NbTracks; i++ {
		//	fmt.Fprintf(os.Stdout,"Loading track %d, total: %d\n", i, cpcEntry.NbTracks)
		track := &CPCEMUTrack{}
		if err := track.Read(r); err != nil {
			fmt.Fprintf(os.Stderr, "Error track (%d) error :%v\n", i, err)
		}
		d.Tracks[i] = *track
		//fmt.Fprintf(os.Stdout, "Track %d %s\n", i, d.Tracks[i].ToString())
	}
	return nil
}

func FormatDsk(nbSect, nbTrack uint8) *DSK {
	dsk := &DSK{}
	entry := CPCEMUEnt{}
	copy(entry.Debut[:], "MV - CPCEMU Disk-File\r\nDisk-Info\r\n")
	entry.DataSize = 0x100 + (SECTSIZE * 9)
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
	for s = 0; s < nbSect; {
		t.Sect[s].C = track
		t.Sect[s].H = 0
		t.Sect[s].R = (ss + minSect)
		t.Sect[s].N = 2
		t.Sect[s].SizeByte = 0x200
		ss++
		s++
		if s < nbSect {
			t.Sect[s].C = track
			t.Sect[s].H = 0
			t.Sect[s].R = (ss + minSect + 4)
			t.Sect[s].N = 2
			t.Sect[s].SizeByte = 0x200
			s++
		}
	}
	t.Data = make([]byte, 0x200*uint16(nbSect))
	for i := 0; i < len(t.Data); i++ {
		t.Data[i] = 0xE5
	}
	if len(d.Tracks) < int(track+1) {
		d.Tracks = append(d.Tracks, t)
		d.Entry.NbTracks++
	} else {
		d.Tracks[track] = t
	}
}

func (d *DSK) Write(w io.Writer) error {
	if err := binary.Write(w, binary.LittleEndian, &d.Entry); err != nil {
		fmt.Fprintf(os.Stderr, "Cannot read CPCEmuEnt error :%v\n", err)
		return err
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
	//fmt.Fprintf(os.Stdout, "Track 0 nbSect :%d \n", tr.NbSect)
	for s = 0; s < tr.NbSect; s++ {
		//	fmt.Fprintf(os.Stdout, "Sector %d, R %d\n", s, tr.Sect[s].R)
		if Sect > tr.Sect[s].R {
			Sect = tr.Sect[s].R
		}
	}
	return Sect
}

//
// Retourne la position d'un secteur dans le fichier DSK, position dans la structure Data
//
func (d *DSK) GetPosData(track, sect uint8, SectPhysique bool) uint16 {
	// Recherche position secteur
	tr := d.Tracks[track]
	var SizeByte uint16
	var Pos uint16
	var s uint8
	//Pos += 256

	//fmt.Fprintf(os.Stdout,"Track:%d,Secteur:%d\n",track,sect)

	//Pos += 256
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
		//fmt.Fprintf(os.Stderr, "sizebyte:%d, t:%d,s:%d,tr.Sect[s].SizeByte:%d, tr->Sect[ s ].N:%d, Pos:%d\n", SizeByte,t,s,tr.Sect[s].SizeByte,tr.Sect[ s ].N,Pos)
	}

	return Pos
}

func (d *DSK) GetFile(path string, indice int) error {
	i := indice
	current := make([]byte, 16)
	nomIndice := make([]byte, 16)
	lMax := 0x1000000
	cumul := 0
	d.GetCatalogue()
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
	} //while (! strncmp( NomIndice, current , max( strlen( NomIndice ), strlen( current ) )));

	return nil
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
	copy(amsdosFile[9:12], ext[1:4])
	return string(amsdosFile)
}

func (d *DSK) PutFile(masque string, typeModeImport uint8, loadAdress, exeAdress, userNumber uint16, isSystemFile, readOnly bool) error {
	buff := make([]byte, 0x20000)
	cFileName := GetNomAmsdos(masque)
	header := &StAmsdos{}
	var addHeader bool
	d.GetCatalogue()
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
	rbuff := bytes.NewReader(buff)
	if err := binary.Read(rbuff, binary.LittleEndian, header); err != nil {
		fmt.Fprintf(os.Stdout, "No header found for file :%s, error :%v\n", masque, err)
	}

	if typeModeImport == MODE_ASCII {
		for i := 0; i < 0x20000; i++ {
			if buff[i] > 136 {
				buff[i] = '?'
			}
		}
	}

	var isAmsdos bool
	//
	// Regarde si le fichier contient une en-tete ou non
	//
	if header.Checksum == header.ComputedChecksum16() {
		isAmsdos = true
	}
	if !isAmsdos {
		// Creer une en-tete amsdos par defaut
		fmt.Fprintf(os.Stdout, "Create header... (%s)\n", masque)
		header = &StAmsdos{Size: uint16(len(buff)), Size2: uint16(len(buff))}
		copy(header.Filename[:], []byte(cFileName[0:14]))
		if loadAdress != 0 {
			header.Address = loadAdress
			typeModeImport = MODE_BINAIRE
		}
		if exeAdress != 0 {
			header.Exec = exeAdress
			typeModeImport = MODE_BINAIRE
		}
		// Il faut recalculer le checksum en comptant es adresses !
		header.Checksum = header.ComputedChecksum16()
	} else {
		fmt.Fprintf(os.Stdout, "File has already header...(%s)\n", masque)
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
			fmt.Fprintf(os.Stdout, "Removing header...(%s)\n", masque)
			copy(buff[0:], buff[binary.Size(StAmsdos{}):])
		}
		break
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
		break
	}
	//
	// Si fichier ok pour etre import
	//
	if addHeader {
		// Ajoute l'en-tete amsdos si necessaire

		var rbuff bytes.Buffer
		binary.Write(&rbuff, binary.LittleEndian, header)
		binary.Write(&rbuff, binary.LittleEndian, buff)
		buff = rbuff.Bytes()
		//	memmove( &Buff[ sizeof( StAmsdos ) ], Buff, Lg );
		//         	memcpy( Buff, e, sizeof( StAmsdos ) );
		//       	Lg += sizeof( StAmsdos );
	}
	if fileLength > 65536 {
		return ErrorFileSizeExceed
	}
	//if (MODE_BINAIRE) ClearAmsdos(Buff); //Remplace les octets inutilises par des 0 dans l'en-tete
	return d.CopyFile(buff, cFileName, uint16(fileLength), 256, userNumber, isSystemFile, readOnly)

}

//
// Copie un fichier sur le DSK
//
// la taille est determine par le nombre de NbPages
// regarder pourquoi different d'une autre DSK
func (d *DSK) CopyFile(bufFile []byte, fileName string, fileLength, maxBloc, userNumber uint16, isSystemFile, readOnly bool) error {
	var nbPages, taillePage int
	d.FillBitmap()
	dirLoc := d.GetNomDir(fileName)
	var posFile uint16                       //Construit l'entree pour mettre dans le catalogue
	for posFile = 0; posFile < fileLength; { //Pour chaque bloc du fichier
		posDir, err := d.RechercheDirLibre() //Trouve une entree libre dans le CAT
		if err == nil {
			dirLoc.User = uint8(userNumber) //Remplit l'entree : User 0
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
			l := (dirLoc.NbPages + 7) >> 3 //Nombre de blocs=TaillePage/8 arrondi par le haut
			for i := 0; i < 16; i++ {
				dirLoc.Blocks[i] = 0
			}
			var j uint8
			for j = 0; j < l; j++ { //Pour chaque bloc de la page
				bloc := d.RechercheBlocLibre(int(maxBloc)) //Met le fichier sur la disquette
				//	fmt.Fprintf(os.Stdout,"Bloc:%d, MaxBloc:%d\n",bloc,maxBloc)
				if bloc != 0 {
					dirLoc.Blocks[j] = bloc
					d.WriteBloc(int(bloc), bufFile, posFile)
					posFile += 1024 // Passe au bloc suivant

				} else {
					return ErrorNoBloc
				}
			}
			//fmt.Fprintf(os.Stdout, "posDir:%d dirloc:%v\n", posDir, dirLoc)
			d.SetInfoDirEntry(posDir, dirLoc)
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
		d.FormatTrack(uint8(track), minSect, 9)
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
		d.FormatTrack(uint8(track), minSect, 9)
	}
	pos = d.GetPosData(uint8(track), uint8(sect)+minSect, true)
	copy(d.Tracks[track].Data[pos:], bufBloc[offset+SECTSIZE:offset+(SECTSIZE*2)])
	return nil
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
	pos := d.GetPosData(uint8(track), uint8(sect), true)
	copy(bufBloc, d.Tracks[track].Data[pos:pos+SECTSIZE])
	//int Pos = GetPosData( track, sect + MinSect, true );
	//memcpy( BufBloc, &ImgDsk[ Pos ], SECTSIZE );
	sect++
	if sect > 8 {
		track++
		sect = 0
	}
	pos = d.GetPosData(uint8(track), uint8(sect), true)
	copy(bufBloc, d.Tracks[track].Data[pos:pos+SECTSIZE])
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
	d.GetCatalogue()
	for i := 0; i < 64; i++ {
		entry := d.Catalogue[i]
		if entry.User != USER_DELETED && entry.NumPage != 0 {
			fmt.Fprintf(os.Stdout, "%s.%s : %d\n", entry.Nom, entry.Ext, entry.User)
		}
	}
}

func (d *DSK) GetEntryyNameInCatalogue(num int) string {
	d.GetCatalogue()
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
	d.GetCatalogue()
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
	data := d.Tracks[t].Data[((uint16(numDir)&15)<<5)+pos : ((uint16(numDir)&15)<<5)+pos+32]
	buffer := bytes.NewReader(data[:])
	if err := binary.Read(buffer, binary.LittleEndian, &dir); err != nil {
		return dir, err
	}
	//memcpy( &Dir
	//		, &ImgDsk[ ( ( NumDir & 15 ) << 5 ) + GetPosData( t, s, true ) ]
	//		, sizeof( StDirEntry )
	//		);
	return dir, nil
}

func (d *DSK) GetType(langue int, ams *StAmsdos) string {
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
