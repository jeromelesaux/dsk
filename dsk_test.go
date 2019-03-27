package dsk

import (
    "testing"
)

func TestOpenDsk(t *testing.T) {

	formated := FormatDsk(9,40)
	if err := WriteDsk("test.dsk",formated); err != nil {
		t.Fatalf("error :%v",err)
	}

	dsk, err := NewDsk("ironman.dsk")
	if err != nil {
		t.Fatalf("error %v",err)
	}
	t.Logf("NBtracks:%d\n",dsk.Entry.NbTracks)
	t.Logf("Head :%d\n",dsk.Entry.NbHeads)
	if err := dsk.CheckDsk(); err != nil {
		t.Fatalf("error %v",err)
	} 
	t.Logf("%s\n",dsk.GetEntryyNameInCatalogue(1))
	t.Logf("%s\n",dsk.GetEntrySizeInCatalogue(1))
	
}