package dsk

import (
	"fmt"
	"testing"
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
