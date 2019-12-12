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
	fos, err := os.Open("os.rom")
	if err != nil {
		t.Fatalf("Expected not error and gets %v\n", err)
	}
	defer fos.Close()
	bos, err := ioutil.ReadAll(fos)
	if err != nil {

	}
	if err := cpr.Copy(bos, 0); err != nil {
		t.Fatalf("Expected not error and gets %v\n", err)
	}
	fbas, err := os.Open("basic.rom")
	if err != nil {
		t.Fatalf("Expected not error and gets %v\n", err)
	}
	defer fbas.Close()
	bas, err := ioutil.ReadAll(fbas)
	if err != nil {

	}
	if err := cpr.Copy(bas, 1); err != nil {
		t.Fatalf("Expected not error and gets %v\n", err)
	}
	fams, err := os.Open("amsdos.rom")
	if err != nil {
		t.Fatalf("Expected not error and gets %v\n", err)
	}
	defer fams.Close()
	bams, err := ioutil.ReadAll(fams)
	if err != nil {

	}
	if err := cpr.Copy(bams, 2); err != nil {
		t.Fatalf("Expected not error and gets %v\n", err)
	}
	if err := cpr.Copy(b, 3); err != nil {
		t.Fatalf("Expected not error and gets %v\n", err)
	}
	cpr.FilePath = "test.cpr"
	cpr.Save()
}

func TestRasmoutRead(t *testing.T) {
	cpr := NewCpr("rasmoutput.cpr")
	err := cpr.Open()
	if err != nil {
		t.Fatalf("Expected not error and gets %v\n", err)
	}

}
