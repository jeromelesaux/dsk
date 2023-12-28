package dsk

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestOpenDsk(t *testing.T) {
	formatted := FormatDsk(9, 40, 1, DataFormat, 0)
	if err := WriteDsk("../testdata/test.dsk", formatted); err != nil {
		t.Fatal(err)
	}
	t.Logf("(%s)=(%s)\n", "../testdata/sonic-pa.bas", GetAmsDosName("../testdata/sonic-pa.bas"))
	if err := formatted.PutFile("../testdata/ironman.scr", SaveModeBinary, 0, 0, 0, false, false); err != nil {
		t.Fatal(err)
	}
	err := WriteDsk("../testdata/test.dsk", formatted)
	if err != nil {
		t.Fatal(err)
	}

	dsk, err := ReadDsk("../testdata/ironman.dsk")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, uint8(42), dsk.Entry.Tracks, "bad number of tracks: %v", dsk.Entry.Tracks)
	assert.Equal(t, uint8(1), dsk.Entry.Heads, "bad number of heads: %v", dsk.Entry.Heads)
	if err := dsk.CheckDsk(); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "IRONMAN .SCR", dsk.GetCatEntryName(0))
	assert.Equal(t, "16 KiB", dsk.GetCatEntrySizeStr(0))
}

func TestFormatDsk(t *testing.T) {
	formatted := FormatDsk(9, 40, 1, DataFormat, 0)
	if err := WriteDsk("test.dsk", formatted); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 40, len(formatted.Tracks))
	for tr := range formatted.Tracks {
		track := formatted.Tracks[tr] // reference to track
		assert.Equal(t, uint8(9), track.SectorCount)
		for s := uint8(0); s < track.SectorCount; s++ {
			sector := track.Sect[s]
			assert.Equal(t, uint16(512), sector.Size)
			buffer := bytes.NewBuffer(make([]byte, 1024))
			err := sector.Write(buffer)
			if err != nil {
				t.Fatal(err)
				return
			}
			for b := uint16(0); b < sector.Size; b++ {
				b, err := buffer.ReadByte()
				if err != nil {
					t.Fatal(err)
					return
				}
				assert.Equal(t, uint8(0), b)
			}
		}
	}
}

func TestFileContentDsk(t *testing.T) {
	disk := FormatDsk(9, 40, 1, DataFormat, 0)
	currentWorkingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	ironmanPath := currentWorkingDirectory + "/../testdata/ironman.scr"
	if err := disk.PutFile(ironmanPath, SaveModeBinary, 0, 0, 0, false, false); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "IRONMAN .SCR", disk.GetCatEntryName(0))
	assert.Equal(t, "16 KiB", disk.GetCatEntrySizeStr(0))
	disk.NewDirEntry("IRONMAN .SCR")
	fileBytes, err := disk.GetFileIn("IRONMAN .SCR", 0)
	if err != nil {
		t.Fatal(err)
	}
	original, err := os.ReadFile(ironmanPath)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, original, fileBytes)
}
