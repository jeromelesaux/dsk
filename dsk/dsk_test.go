package dsk

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"os"
	"testing"

	"github.com/jeromelesaux/dsk/amsdos"
	"github.com/stretchr/testify/assert"
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

	fmt.Printf("%d", v|mask)
	fmt.Printf("%d", v0|mask)
	fmt.Printf("%d", v1|mask)
	fmt.Printf("%d", byte(0b00000011))
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

func TestCleanBitmap(t *testing.T) {
	d := FormatDsk(9, 40, 1, DataFormat, 0)
	d.BitMap[0] = 0xFF
	d.BitMap[255] = 0xAA
	d.CleanBitmap()
	for i, b := range d.BitMap {
		assert.Equalf(t, byte(0), b, "bitmap index %d", i)
	}
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
	track.Sect[0] = CPCEMUSect{C: 0, H: 0, R: 1, N: 2, Un1: 0, SizeByte: 0x0200}
	track.Data = make([]byte, int(track.Sect[0].SizeByte))
	for i := range track.Data {
		track.Data[i] = 0xE5
	}

	var buf bytes.Buffer
	assert.NoError(t, track.Write(&buf))
	var track2 CPCEMUTrack
	assert.NoError(t, track2.Read(&buf))
	assert.Equal(t, track.Track, track2.Track)
	assert.Equal(t, track.Head, track2.Head)
	assert.Equal(t, track.NbSect, track2.NbSect)
	assert.Equal(t, track.Sect[0].R, track2.Sect[0].R)
	assert.Equal(t, track.Data[:len(track2.Data)], track2.Data)
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
	err := readDsk.Read(bad)
	assert.ErrorIs(t, err, ErrorUnsupportedDskFormat)
}

func TestFormatDskVendorFormat(t *testing.T) {
	d := FormatDsk(9, 40, 1, VendorFormat, 0)
	assert.NotNil(t, d)
	assert.Equal(t, uint8(40), d.Entry.NbTracks)
	assert.Equal(t, uint8(1), d.Entry.NbHeads)
	assert.Equal(t, 40, len(d.TrackSizeTable))
	assert.Equal(t, uint16(0x1300), d.Entry.DataSize)
}

func TestFormatDskExtended(t *testing.T) {
	d := FormatDsk(9, 2, 2, DataFormat, EXTENDED_DSK_TYPE)
	assert.NotNil(t, d)
	assert.True(t, d.Extended)
	assert.Equal(t, 4, len(d.TrackSizeTable))
	assert.Equal(t, uint8(2), d.Entry.NbHeads)
	assert.Equal(t, uint8(2), d.Entry.NbTracks)
}

func TestWriteReadDskFileRoundTrip(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "dsk-roundtrip-*.dsk")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	d := FormatDsk(9, 1, 1, DataFormat, 0)
	assert.NoError(t, WriteDsk(tmpFile.Name(), d))

	readDsk, err := ReadDsk(tmpFile.Name())
	assert.NoError(t, err)
	assert.Equal(t, d.Entry.NbTracks, readDsk.Entry.NbTracks)
	assert.Equal(t, d.Entry.NbHeads, readDsk.Entry.NbHeads)
	assert.Equal(t, len(d.TrackSizeTable), len(readDsk.TrackSizeTable))
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
		fmt.Fprintf(os.Stdout, "error while writing in header %v\n", err)
	}
	err = binary.Write(&rbuff, binary.LittleEndian, data)
	if err != nil {
		fmt.Fprintf(os.Stdout, "error while writing in content %v\n", err)
	}
	return rbuff.Bytes()
}
