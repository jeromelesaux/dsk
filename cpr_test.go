package dsk

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestEmptyCprSave(t *testing.T) {
	cpr := NewCpr("empty.cpr")
	err := cpr.Save()
	if err != nil {
		t.Fatalf("Expected not error and gets %v\n", err)
	}
}

func TestCopyData(t *testing.T) {
	TestEmptyCprSave(t)
	cpr := NewCpr("empty.cpr")
	if err := cpr.Open(); err != nil {
		t.Fatalf("Expected not error and gets %v\n", err)
	}
	f, err := os.Open("sonic-pa.BAS")
	if err != nil {
		t.Fatalf("Expected not error and gets %v\n", err)
	}
	defer f.Close()
	b, err := ioutil.ReadAll(f)
	if err != nil {

	}
	if err := cpr.Copy(b, 0); err != nil {
		t.Fatalf("Expected not error and gets %v\n", err)
	}
	cpr.FilePath = "test.cpr"
	cpr.Save()
}
