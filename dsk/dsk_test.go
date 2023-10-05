package dsk

import (
	"testing"
)

func TestOpenDsk(t *testing.T) {
	formated := FormatDsk(9, 40, 1, DataFormat, 0)
	if err := WriteDsk("../testdata/test.dsk", formated); err != nil {
		t.Fatalf("error :%v", err)
	}
	t.Logf("(%s)=(%s)\n", "../testdata/sonic-pa.bas", GetNomAmsdos("../testdata/sonic-pa.bas"))
	if err := formated.PutFile("../testdata/ironman.scr", MODE_BINAIRE, 0, 0, 0, false, false); err != nil {
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
