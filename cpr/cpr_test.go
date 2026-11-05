package cpr_test

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/jeromelesaux/dsk/cpr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCPR(t *testing.T) {
	cartridge := filepath.Join(t.TempDir(), "empty.cpr")
	t.Run("emptyCPR", func(t *testing.T) {
		cpr := cpr.NewCpr(cartridge)
		err := cpr.Save()
		if err != nil {
			t.Fatalf("Expected not error and gets %v\n", err)
		}
	})

	t.Run("copyCPR", func(t *testing.T) {
		cpr := cpr.NewCpr(cartridge)
		if err := cpr.Open(); err != nil {
			t.Fatalf("Expected not error and gets %v\n", err)
		}
		f, err := os.Open("../testdata/sonic-pa.BAS")
		if err != nil {
			t.Fatalf("Expected not error and gets %v\n", err)
		}
		defer func() { _ = f.Close() }()
		b, err := io.ReadAll(f)
		if err != nil {
			t.Fatal()
		}

		if err := cpr.Copy(b[128:], 3); err != nil {
			t.Fatalf("Expected not error and gets %v\n", err)
		}
		cpr.FilePath = filepath.Join(t.TempDir(), "test.cpr")
		err = cpr.Save()
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("addData", func(t *testing.T) {
		//	TestEmptyCprSave(t)
		cpr := cpr.NewCpr(cartridge)
		if err := cpr.AddFile("../testdata/ironman.scr", 3); err != nil {
			t.Fatalf("Expected not error and gets %v\n", err)
		}
		cpr.FilePath = "ironman.cpr"
		err := cpr.Save()
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("eraseCPR", func(t *testing.T) {
		// cleanup handled by t.TempDir()
	})
}

func TestCopyBankOverflow(t *testing.T) {
	c := cpr.NewCpr("")
	err := c.Copy([]byte{1}, 32)
	assert.ErrorIs(t, err, cpr.ErrorBanknumber)
}

func TestCopyDataExceed(t *testing.T) {
	c := cpr.NewCpr("")
	big := make([]byte, 0x4001)
	err := c.Copy(big, 0)
	assert.ErrorIs(t, err, cpr.ErrorBankdataExceed)
}

func TestPatch(t *testing.T) {
	c := cpr.NewCpr("")
	require.NoError(t, c.Patch(0, 0x100, 0xAA))
	assert.Equal(t, byte(0xAA), c.DataZone.BankZone[0].BankData[0x100])
}

func TestPatchBankOverflow(t *testing.T) {
	c := cpr.NewCpr("")
	assert.ErrorIs(t, c.Patch(32, 0, 0), cpr.ErrorBanknumber)
}

func TestPatchOffsetOverflow(t *testing.T) {
	c := cpr.NewCpr("")
	assert.ErrorIs(t, c.Patch(0, 0x4001, 0), cpr.ErrorBankdataExceed)
}

func TestCopyOffset(t *testing.T) {
	c := cpr.NewCpr("")
	data := []byte{0xDE, 0xAD}
	require.NoError(t, c.CopyOffset(0, 0x10, data))
	assert.Equal(t, byte(0xDE), c.DataZone.BankZone[0].BankData[0x10])
	assert.Equal(t, byte(0xAD), c.DataZone.BankZone[0].BankData[0x11])
}

func TestCopyOffsetBankOverflow(t *testing.T) {
	c := cpr.NewCpr("")
	assert.ErrorIs(t, c.CopyOffset(32, 0, []byte{1}), cpr.ErrorBanknumber)
}

func TestCopyOffsetExceed(t *testing.T) {
	c := cpr.NewCpr("")
	assert.ErrorIs(t, c.CopyOffset(0, 0x4001, []byte{1}), cpr.ErrorBankdataExceed)
}

func TestAddBankOverflow(t *testing.T) {
	c := cpr.NewCpr("")
	_, err := c.Add([]byte{1}, 32)
	assert.ErrorIs(t, err, cpr.ErrorBanknumber)
}

func TestAddFileNotFound(t *testing.T) {
	c := cpr.NewCpr("")
	err := c.AddFile("/nonexistent/file.bin", 0)
	assert.Error(t, err)
}

func TestOpenNotFound(t *testing.T) {
	c := cpr.NewCpr("/nonexistent/file.cpr")
	err := c.Open()
	assert.Error(t, err)
}

func TestNewCprStructure(t *testing.T) {
	c := cpr.NewCpr("test.cpr")
	assert.Equal(t, [4]byte{'R', 'I', 'F', 'F'}, c.Riff)
	assert.Equal(t, [4]byte{'A', 'M', 'S', '!'}, c.DataZone.Ams)
	assert.Equal(t, uint32(0x4000), c.DataZone.BankZone[0].BankSize)
}
