package dsk

import "testing"

func TestEmptyCprSave(t *testing.T) {
	cpr := NewCpr("empty.cpr")
	err := cpr.Save()
	if err != nil {
		t.Fatalf("Expected not error and gets %v\n", err)
	}
}
