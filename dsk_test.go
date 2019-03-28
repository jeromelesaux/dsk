package dsk

import (
    "testing"
)

func TestOpenDsk(t *testing.T) {

	formated := FormatDsk(9,40)
	if err := WriteDsk("test.dsk",formated); err != nil {
		t.Fatalf("error :%v",err)
	}
	t.Logf("(%s)=(%s)\n","/opt/data/sonic-pa.bas",GetNomAmsdos("/opt/data/sonic-pa.bas"))
	if err :=formated.PutFile("ironman.scr",MODE_BINAIRE,0,0,0,false,false); err != nil {
		t.Fatalf("Error:%v",err)
	}
	WriteDsk("test.dsk",formated)

	dsk, err := ReadDsk("ironman.dsk")
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