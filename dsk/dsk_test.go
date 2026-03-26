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
