package dsk

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestOpenDsk(t *testing.T) {
	formatted := FormatDsk(9, 40, 1, DataFormat, 0)
	if err := WriteDsk("test.dsk", formatted); err != nil {
		t.Fatal(err)
	}
	t.Logf("(%s)=(%s)\n", "/opt/data/sonic-pa.bas", GetAmsDosName("/opt/data/sonic-pa.bas"))
	if err := formatted.PutFile("ironman.scr", SaveModeBinary, 0, 0, 0, false, false); err != nil {
		t.Fatal(err)
	}
	err := WriteDsk("test.dsk", formatted)
	if err != nil {
		t.Fatal(err)
	}

	dsk, err := ReadDsk("ironman.dsk")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, uint8(42), dsk.Entry.Tracks, "bad number of tracks: %v", dsk.Entry.Tracks)
	assert.Equal(t, uint8(1), dsk.Entry.Heads, "bad number of heads: %v", dsk.Entry.Heads)
	if err := dsk.CheckDsk(); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "IRONMAN .SCR", dsk.GetEntryyNameInCatalogue(1))
	assert.Equal(t, "16 KiB", dsk.GetEntrySizeInCatalogue(1))
}
