package dsk

import (
    "testing"
)

func TestOpenDsk(t *testing.T) {

	/*formated := FormatDsk(9,40)
	if err := formated.Write("test.dsk"); err != nil {
		t.Fatalf("error :%v",err)
	}*/

	dsk, err := NewDsk("test.dsk")
	if err != nil {
		t.Fatalf("error %v",err)
	}
	t.Logf("NBtracks:%d\n",dsk.Entry.NbTracks)
	t.Logf("Head :%d\n",dsk.Entry.NbHeads)
	if err := dsk.CheckDsk(); err != nil {
		t.Fatalf("error %v",err)
	} 
}