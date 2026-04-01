package sna_test

import (
	"bytes"
	"testing"

	"github.com/jeromelesaux/dsk/dsk"
	"github.com/jeromelesaux/dsk/sna"
	"github.com/stretchr/testify/assert"
)

func TestMemoryChunck(t *testing.T) {
	t.Run("uncompress", func(t *testing.T) {
		m := &sna.MemChunck{}
		buf := []byte{0xe5, 0xff, 0x2}
		m.Feed(buf)
		got := m.Data[0:0xff]
		for i := 0; i < len(got); i++ {
			assert.Equal(t, byte(2), got[i])
		}
	})

	t.Run("compress", func(t *testing.T) {
		m := &sna.MemChunck{}
		for i := 0; i < 0xff; i++ {
			m.Data[i] = 0x2
		}
		got := m.Export()[0:3]
		expected := []byte{0xe5, 0xff, 0x2}
		assert.Equal(t, expected, got)
	})
}

func TestNewSna(t *testing.T) {
	header := sna.NewSnaHeader()
	s := sna.NewSna(header)
	assert.NotNil(t, s)
	assert.Equal(t, header, s.Header)
}

func TestCPCValue(t *testing.T) {
	tests := []struct {
		cpc      sna.CPC
		expected uint8
	}{
		{sna.CPC464, 0},
		{sna.CPC664, 1},
		{sna.CPC6128, 2},
	}

	for _, tt := range tests {
		t.Run(string(tt.cpc), func(t *testing.T) {
			result := sna.CPCValue(tt.cpc)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCRTCValue(t *testing.T) {
	assert.Equal(t, uint8(0), sna.CRTCValue(sna.HD6845S_UM6845))
	assert.Equal(t, uint8(1), sna.CRTCValue(sna.UM6845R))
	assert.Equal(t, uint8(4), sna.CRTCValue(sna.Pre_ASIC))
}

func TestCPCTypeFunc(t *testing.T) {
	assert.Equal(t, sna.CPC464, sna.CPCType(0))
	assert.Equal(t, sna.Unknown, sna.CPCType(3))
	assert.Equal(t, sna.GX4000, sna.CPCType(6))
	assert.Equal(t, sna.Unknown, sna.CPCType(99))
}

func TestSnaPutGet(t *testing.T) {
	s := sna.NewSna(sna.NewSnaHeader())
	content := make([]byte, 20)
	for i := 0; i < len(content); i++ {
		content[i] = byte(0xAA + i)
	}
	assert.NoError(t, s.Put(content, 10, uint16(len(content))))
	got, err := s.Get(10, 10)
	assert.NoError(t, err)
	assert.Equal(t, content[:10], got)
}

func TestSnaPutWithoutStartAddress(t *testing.T) {
	s := sna.NewSna(sna.NewSnaHeader())
	content := []byte{0xAA, 0xBB, 0xCC}
	err := s.Put(content, 0, uint16(len(content)))
	assert.ErrorIs(t, err, sna.ErrorNoHeaderOrStartAddress)
}

func TestSnaGetRangeError(t *testing.T) {
	s := sna.NewSna(sna.NewSnaHeader())
	_, err := s.Get(65000, 1000)
	assert.ErrorIs(t, err, dsk.ErrorFileSizeExceed)
}

func TestSnaReadWriteRoundTrip(t *testing.T) {
	s := sna.NewSna(sna.NewSnaHeader())
	s.Data[0] = 0x11
	s.Data[1] = 0x22
	var buf bytes.Buffer
	assert.NoError(t, s.Write(&buf))

	s2 := &sna.SNA{}
	assert.NoError(t, s2.Read(bytes.NewReader(buf.Bytes())))
	assert.Equal(t, s.Header.Version, s2.Header.Version)
	assert.Equal(t, byte(0x11), s2.Data[0])
	assert.Equal(t, byte(0x22), s2.Data[1])
}

