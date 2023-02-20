package dsk

import (
	"io"
	"os"
	"testing"
)

func TestCPR(t *testing.T) {
	cartridge := "empty.cpr"
	t.Run("emptyCPR", func(t *testing.T) {
		cpr := NewCpr(cartridge)
		err := cpr.Save()
		if err != nil {
			t.Fatalf("Expected not error and gets %v\n", err)
		}
	})

	t.Run("copyCPR", func(t *testing.T) {
		cpr := NewCpr(cartridge)
		if err := cpr.Open(); err != nil {
			t.Fatalf("Expected not error and gets %v\n", err)
		}
		f, err := os.Open("sonic-pa.BAS")
		if err != nil {
			t.Fatalf("Expected not error and gets %v\n", err)
		}
		defer f.Close()
		b, err := io.ReadAll(f)
		if err != nil {
			t.Fatal()
		}

		if err := cpr.Copy(b[128:], 3); err != nil {
			t.Fatalf("Expected not error and gets %v\n", err)
		}
		cpr.FilePath = "test.cpr"
		err = cpr.Save()
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("addData", func(t *testing.T) {
		//	TestEmptyCprSave(t)
		cpr := NewCpr(cartridge)
		if err := cpr.AddFile("ironman.scr", 3); err != nil {
			t.Fatalf("Expected not error and gets %v\n", err)
		}
		cpr.FilePath = "ironman.cpr"
		err := cpr.Save()
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("eraseCPR", func(t *testing.T) {
		os.Remove(cartridge)
		os.Remove("test.cpr")
	})
}
