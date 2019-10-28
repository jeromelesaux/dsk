package m4

import (
	"testing"
)

func TestNodeParsing(t *testing.T) {
	content := "Jeux/\n" +
		"Ishido.dsk,1,190K\n" +
		"Doomsday_Lost_Echoes_v1.0,0,0\n" +
		"GalacticTomb_128K,0,0\n" +
		"ImperialMahjong,0,0\n" +
		"Orion Prime (FR) (Cargosoft),0,0\n" +
		"The Shadows Of Sergoth v1.0 (F,UK,S) (128K) (Face A) (2018) [Original].dsk,1,190K\n" +
		"The Shadows Of Sergoth v1.0 (F,UK,S) (128K) (Face B) (2018) [Original].dsk,1,190K\n" +
		"Ishido,0,0"

	d := NewM4Dir(content)
	if d.CurrentPath != "Jeux/" {
		t.Fatalf("Expected currentpath value Jeux/ and gets %s", d.CurrentPath)
	}
	if len(d.Nodes) != 8 {
		t.Fatalf("Expected 8 nodes and gets :%d", len(d.Nodes))
	}
}
