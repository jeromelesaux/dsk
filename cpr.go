package dsk

import (
	"encoding/binary"
	"errors"
	"fmt"
	"os"
)

var (
	ErrorBanknumber     = errors.New("Bad banknumber, overflow.")
	ErrorBankdataExceed = errors.New("Size exceed bank size #4000")
)

type Cpr struct {
	FilePath  string   // file path of the cardright
	Riff      [4]byte  // RIFF constant
	TotalSize uint32   // totalsize of the 32 banks -> 4 + ( 2 + 2 + 4 + 16384) * 32
	DataZone  DataZone // banks store
}

type DataZone struct {
	Ams      [4]byte      // AMS!
	BankZone [32]BankZone // 32 banks of the card
}

type BankZone struct {
	Cb         [2]byte // constant cb
	BankNumber [2]byte // bank number as string 00 to 31
	BankSize   uint32  // size of the bank 0X4000
	BankData   [0x4000]byte
}

func NewCpr(filePath string) *Cpr {
	cpr := &Cpr{
		FilePath:  filePath,
		Riff:      [4]byte{'R', 'I', 'F', 'F'},
		TotalSize: 4 + (2+2+4+16384)*32,
		DataZone: DataZone{
			Ams: [4]byte{'A', 'M', 'S', '!'},
		},
	}
	cb := [2]byte{'c', 'b'}
	for i := 0; i < 32; i++ {
		copy(cpr.DataZone.BankZone[i].Cb[:], cb[:])
		//second := i % 10
		//first := (i - second) / 10
		banknumber := fmt.Sprintf("%.2d", i)
		cpr.DataZone.BankZone[i].BankNumber[0] = banknumber[0]
		cpr.DataZone.BankZone[i].BankNumber[1] = banknumber[1]
		cpr.DataZone.BankZone[i].BankSize = 0x4000
	}
	return cpr
}

func (c *Cpr) Save() error {
	f, err := os.Create(c.FilePath)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := binary.Write(f, binary.LittleEndian, c.Riff); err != nil {
		return err
	}
	if err := binary.Write(f, binary.LittleEndian, c.TotalSize); err != nil {
		return err
	}

	if err := binary.Write(f, binary.LittleEndian, c.DataZone.Ams); err != nil {
		return err
	}

	for i := 0; i < 32; i++ {
		if err := binary.Write(f, binary.LittleEndian, c.DataZone.BankZone[i].Cb); err != nil {
			return err
		}
		if err := binary.Write(f, binary.LittleEndian, c.DataZone.BankZone[i].BankNumber); err != nil {
			return err
		}
		if err := binary.Write(f, binary.LittleEndian, c.DataZone.BankZone[i].BankSize); err != nil {
			return err
		}
		if err := binary.Write(f, binary.LittleEndian, c.DataZone.BankZone[i].BankData); err != nil {
			return err
		}
	}

	return nil
}

func (c *Cpr) Open() error {
	f, err := os.Open(c.FilePath)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := binary.Read(f, binary.LittleEndian, &c.Riff); err != nil {
		return err
	}
	if err := binary.Read(f, binary.LittleEndian, c.TotalSize); err != nil {
		return err
	}

	if err := binary.Read(f, binary.LittleEndian, &c.DataZone.Ams); err != nil {
		return err
	}

	for i := 0; i < 32; i++ {
		if err := binary.Read(f, binary.LittleEndian, &c.DataZone.BankZone[i].Cb); err != nil {
			return err
		}
		if err := binary.Read(f, binary.LittleEndian, &c.DataZone.BankZone[i].BankNumber); err != nil {
			return err
		}
		if err := binary.Read(f, binary.LittleEndian, c.DataZone.BankZone[i].BankSize); err != nil {
			return err
		}
		if err := binary.Read(f, binary.LittleEndian, &c.DataZone.BankZone[i].BankData); err != nil {
			return err
		}
	}

	return nil
}

func (c *Cpr) Copy(b []byte, banknumber int) error {
	if banknumber >= 32 {
		return ErrorBanknumber
	}
	if len(b) > 0x4000 {
		return ErrorBankdataExceed
	}
	copy(c.DataZone.BankZone[banknumber].BankData[:], b)
	return nil
}
