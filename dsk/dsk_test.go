package dsk

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/jeromelesaux/dsk/amsdos"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenDsk(t *testing.T) {
	formated := FormatDsk(9, 40, 1, DataFormat, 0)
	if err := WriteDsk("../testdata/test.dsk", formated); err != nil {
		t.Fatalf("error :%v", err)
	}
	t.Logf("(%s)=(%s)\n", "../testdata/sonic-pa.bas", GetNomAmsdos("../testdata/sonic-pa.bas"))
	if err := formated.PutFile("../testdata/ironman.scr", MODE_BINAIRE, 0, 0, 0, false, false, false); err != nil {
		t.Fatalf("Error:%v", err)
	}
	err := WriteDsk("../testdata/test.dsk", formated)
	if err != nil {
		t.Fatal(err)
	}

	dsk, err := ReadDsk("../testdata/ironman.dsk")
	if err != nil {
		t.Fatalf("error %v", err)
	}
	t.Logf("NBtracks:%d\n", dsk.Entry.NbTracks)
	t.Logf("Head :%d\n", dsk.Entry.NbHeads)
	if err := dsk.CheckDsk(); err != nil {
		t.Fatalf("error %v", err)
	}
	t.Logf("%s\n", dsk.GetEntryyNameInCatalogue(1))
	t.Logf("%s\n", dsk.GetEntrySizeInCatalogue(1))
}

func TestMaskBit7(t *testing.T) {
	v := byte(0b00000001)
	v0 := byte(0b00000000)
	v1 := byte(0b00000010)
	mask := byte(0b00000001)

	assert.Equal(t, byte(1), v|mask)
	assert.Equal(t, byte(1), v0|mask)
	assert.Equal(t, byte(3), v1|mask)
}

func TestFormatDsk(t *testing.T) {
	d := FormatDsk(9, 40, 1, DataFormat, 0)
	assert.NotNil(t, d)
	assert.Equal(t, uint8(40), d.Entry.NbTracks)
	assert.Equal(t, uint8(1), d.Entry.NbHeads)
}

func TestGetNomAmsdos(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"file.bin", "FILE    .BIN"},
		{"test.bas", "TEST    .BAS"},
		{"longfilename.txt", "LONGFILE.TXT"},
		{"a.txt", "A       .TXT"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := GetNomAmsdos(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPut(t *testing.T) {
	d := FormatDsk(9, 40, 1, DataFormat, 0)
	data := generateData(40136)
	name := GetNomAmsdos("FILE.BIN")
	data = addHeader(data, name)
	assert.NoError(t, d.CopyFile(data, name, uint16(len(data)), 256, uint16(MODE_BINAIRE), false, false, false))
	_, err := d.GetFileIn(name, 0)
	assert.NoError(t, err)
}

// --- nouveaux tests ---

func TestReadDskInvalidPath(t *testing.T) {
	_, err := ReadDsk("/nonexistent/file.dsk")
	assert.Error(t, err)
}

func TestReadDskInvalidFormat(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "bad.dsk")
	require.NoError(t, os.WriteFile(tmpFile, make([]byte, 512), 0644))
	dsk, _ := ReadDsk(tmpFile)
	assert.Equal(t, uint8(0), dsk.Entry.NbTracks)
}

func TestWriteReadRoundtrip(t *testing.T) {
	d := FormatDsk(9, 40, 1, DataFormat, 0)
	path := filepath.Join(t.TempDir(), "roundtrip.dsk")
	require.NoError(t, WriteDsk(path, d))
	loaded, err := ReadDsk(path)
	require.NoError(t, err)
	assert.Equal(t, d.Entry.NbTracks, loaded.Entry.NbTracks)
	assert.Equal(t, d.Entry.NbHeads, loaded.Entry.NbHeads)
}

func TestFormatDskExtended(t *testing.T) {
	d := FormatDsk(9, 2, 2, DataFormat, EXTENDED_DSK_TYPE)
	assert.True(t, d.Extended)
	assert.Equal(t, 4, len(d.TrackSizeTable))
	assert.Equal(t, uint8(2), d.Entry.NbHeads)
	assert.Equal(t, uint8(2), d.Entry.NbTracks)
}

func TestFormatDskVendor(t *testing.T) {
	d := FormatDsk(9, 40, 1, VendorFormat, 0)
	assert.True(t, d.Extended)
	assert.Equal(t, uint8(40), d.Entry.NbTracks)
	assert.Equal(t, 40, len(d.TrackSizeTable))
}

func TestCheckDsk(t *testing.T) {
	d := FormatDsk(9, 40, 1, DataFormat, 0)
	assert.NoError(t, d.CheckDsk())
}

func TestDskSize(t *testing.T) {
	d := FormatDsk(9, 40, 1, DataFormat, 0)
	assert.Greater(t, d.DskSize(), uint16(0))
}

func TestGetMinSect(t *testing.T) {
	d := FormatDsk(9, 40, 1, DataFormat, 0)
	assert.Equal(t, uint8(0xC1), d.GetMinSect())
}

func TestGetMinSectVendor(t *testing.T) {
	d := FormatDsk(9, 40, 1, VendorFormat, 0)
	assert.Equal(t, uint8(0x41), d.GetMinSect())
}

func TestCleanBitmap(t *testing.T) {
	d := FormatDsk(9, 40, 1, DataFormat, 0)
	d.BitMap[0] = 0xFF
	d.BitMap[255] = 0xAA
	d.CleanBitmap()
	for i, b := range d.BitMap {
		assert.Equalf(t, byte(0), b, "bitmap index %d", i)
	}
}

func TestFillBitmap(t *testing.T) {
	d := FormatDsk(9, 40, 1, DataFormat, 0)
	nbKo := d.FillBitmap()
	assert.Equal(t, 0, nbKo)
	assert.Equal(t, byte(1), d.BitMap[0])
	assert.Equal(t, byte(1), d.BitMap[1])
}

func TestRechercheDirLibre(t *testing.T) {
	d := FormatDsk(9, 40, 1, DataFormat, 0)
	require.NoError(t, d.GetCatalogue())
	pos, err := d.RechercheDirLibre()
	assert.NoError(t, err)
	assert.True(t, pos < 64)
}

func TestRechercheBlocLibre(t *testing.T) {
	d := FormatDsk(9, 40, 1, DataFormat, 0)
	d.FillBitmap()
	assert.Greater(t, d.RechercheBlocLibre(256), uint8(1))
}

func TestFileExistsNotFound(t *testing.T) {
	d := FormatDsk(9, 40, 1, DataFormat, 0)
	require.NoError(t, d.GetCatalogue())
	entry := StDirEntry{}
	copy(entry.Nom[:], "NOTEXIST")
	copy(entry.Ext[:], "BIN")
	assert.Equal(t, NOT_FOUND, d.FileExists(entry))
}

func TestGetFilesIndicesEmpty(t *testing.T) {
	d := FormatDsk(9, 40, 1, DataFormat, 0)
	require.NoError(t, d.GetCatalogue())
	assert.Empty(t, d.GetFilesIndices())
}

func TestCopyFileAndGetFileIn(t *testing.T) {
	d := FormatDsk(9, 40, 1, DataFormat, 0)
	data := generateData(40136)
	name := GetNomAmsdos("FILE.BIN")
	data = addHeader(data, name)
	require.NoError(t, d.CopyFile(data, name, uint16(len(data)), 256, 0, false, false, false))
	result, err := d.GetFileIn(name, 0)
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
}

func TestRemoveFile(t *testing.T) {
	d := FormatDsk(9, 40, 1, DataFormat, 0)
	data := generateData(40136)
	name := GetNomAmsdos("DEL.BIN")
	data = addHeader(data, name)
	require.NoError(t, d.CopyFile(data, name, uint16(len(data)), 256, 0, false, false, false))

	entry := d.GetNomDir(name, false)
	idx := d.FileExists(entry)
	require.NotEqual(t, NOT_FOUND, idx)

	require.NoError(t, d.RemoveFile(uint8(idx)))
	assert.Equal(t, NOT_FOUND, d.FileExists(entry))
}

func TestGetNomDirMethod(t *testing.T) {
	d := FormatDsk(9, 40, 1, DataFormat, 0)
	entry := d.GetNomDir("TESTFILE.BAS", false)
	assert.Equal(t, byte('T'), entry.Nom[0])
	assert.Equal(t, byte('B'), entry.Ext[0])
}

func TestGetNomDirHidden(t *testing.T) {
	d := FormatDsk(9, 40, 1, DataFormat, 0)
	entry := d.GetNomDir("TESTFILE.BAS", true)
	assert.True(t, entry.Ext[1]&0x80 != 0)
}

func TestCPCEMUEntToString(t *testing.T) {
	e := CPCEMUEnt{NbTracks: 40, NbHeads: 1, DataSize: 0x1300}
	s := e.ToString()
	assert.Contains(t, s, "nbTracks:40")
	assert.Contains(t, s, "nbHeads:1")
}

func TestCPCEMUSectToString(t *testing.T) {
	s := CPCEMUSect{C: 1, H: 0, R: 0xC1, N: 2, SizeByte: 0x200}
	assert.Contains(t, s.ToString(), "C:1")
	assert.Contains(t, s.ToString(), "R:193")
}

func TestCPCEMUTrackToString(t *testing.T) {
	tr := CPCEMUTrack{Track: 5, Head: 0, SectSize: 2, NbSect: 9, Gap3: 0x4E}
	result := tr.ToString()
	assert.Contains(t, result, "Track:5")
	assert.Contains(t, result, "nbSect:9")
}

func TestGetPosData(t *testing.T) {
	d := FormatDsk(9, 40, 1, DataFormat, 0)
	assert.Equal(t, uint16(0), d.GetPosData(2, 0xC1, true))
}

func TestExtendedDskWriteReadRoundtrip(t *testing.T) {
	d := FormatDsk(9, 40, 1, DataFormat, EXTENDED_DSK_TYPE)
	path := filepath.Join(t.TempDir(), "extended.dsk")
	require.NoError(t, WriteDsk(path, d))
	loaded, err := ReadDsk(path)
	require.NoError(t, err)
	assert.True(t, loaded.Extended)
	assert.Equal(t, d.Entry.NbTracks, loaded.Entry.NbTracks)
}

func TestCPCEMUSectRoundTrip(t *testing.T) {
	sect := CPCEMUSect{C: 1, H: 0, R: 2, N: 3, Un1: 0x1234, SizeByte: 0x0200}
	var buf bytes.Buffer
	assert.NoError(t, sect.Write(&buf))
	var sect2 CPCEMUSect
	assert.NoError(t, sect2.Read(&buf))
	assert.Equal(t, sect, sect2)
}

func TestCPCEMUTrackRoundTrip(t *testing.T) {
	track := CPCEMUTrack{Track: 0, Head: 0, SectSize: 2, NbSect: 1, Gap3: 0x4E, OctRemp: 0xE5}
	copy(track.ID[:], "Track-Info\r\n")
	track.Sect[0] = CPCEMUSect{C: 0, H: 0, R: 1, N: 2, SizeByte: 0x0200}
	track.Data = make([]byte, int(track.Sect[0].SizeByte))
	for i := range track.Data {
		track.Data[i] = 0xE5
	}
	var buf bytes.Buffer
	assert.NoError(t, track.Write(&buf))
	var track2 CPCEMUTrack
	assert.NoError(t, track2.Read(&buf))
	assert.Equal(t, track.Track, track2.Track)
	assert.Equal(t, track.NbSect, track2.NbSect)
	assert.Equal(t, track.Sect[0].R, track2.Sect[0].R)
	assert.Equal(t, track.Data, track2.Data)
}

func TestDSKReadWriteRoundTrip(t *testing.T) {
	d := FormatDsk(9, 1, 1, DataFormat, 0)
	var buf bytes.Buffer
	assert.NoError(t, d.Write(&buf))
	readDsk := &DSK{}
	assert.NoError(t, readDsk.Read(bytes.NewReader(buf.Bytes())))
	assert.Equal(t, d.Entry.NbTracks, readDsk.Entry.NbTracks)
	assert.Equal(t, d.Entry.NbHeads, readDsk.Entry.NbHeads)
	assert.Len(t, readDsk.Tracks, 1)
	assert.Equal(t, d.Tracks[0].Track, readDsk.Tracks[0].Track)
}

func TestDSKReadUnsupportedFormat(t *testing.T) {
	bad := &bytes.Buffer{}
	entry := CPCEMUEnt{}
	copy(entry.Debut[:], "XXXX")
	assert.NoError(t, binary.Write(bad, binary.LittleEndian, entry))
	readDsk := &DSK{}
	assert.ErrorIs(t, readDsk.Read(bad), ErrorUnsupportedDskFormat)
}

func TestPutFileAscii(t *testing.T) {
	d := FormatDsk(9, 40, 1, DataFormat, 0)
	tmpfile := filepath.Join(t.TempDir(), "hello.txt")
	content := []byte("hello cpc\n")
	require.NoError(t, os.WriteFile(tmpfile, content, 0644))
	require.NoError(t, d.PutFile(tmpfile, MODE_ASCII, 0, 0, 0, false, false, false))

	dskPath := filepath.Join(t.TempDir(), "ascii.dsk")
	require.NoError(t, WriteDsk(dskPath, d))

	reread, err := ReadDsk(dskPath)
	require.NoError(t, err)
	read, err := reread.GetFileIn(GetNomAmsdos(tmpfile), 0)
	assert.NoError(t, err)
	assert.True(t, bytes.Equal(read[:len(content)], content))
}

func generateData(len int) []byte {
	data := make([]byte, len)
	rand.Read(data)
	return data
}

func addHeader(data []byte, name string) []byte {
	header := &amsdos.StAmsdos{}
	header.User = MODE_BINAIRE
	header.Size = uint16(len(data))
	header.Size2 = uint16(len(data))
	header.LogicalSize = uint16(len(data))
	header.Type = MODE_BINAIRE
	copy(header.Filename[:], []byte(name[0:12]))
	header.Address = 0x0000
	header.Checksum = header.ComputedChecksum16()
	var rbuff bytes.Buffer
	err := binary.Write(&rbuff, binary.LittleEndian, header)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stdout, "error while writing in header %v\n", err)
	}
	err = binary.Write(&rbuff, binary.LittleEndian, data)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stdout, "error while writing in content %v\n", err)
	}
	return rbuff.Bytes()
}
