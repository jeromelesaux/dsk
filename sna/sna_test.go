package sna_test

import (
	"bytes"
	"testing"

	"github.com/jeromelesaux/dsk/sna"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		{sna.Unknown, 3},
		{sna.CPCPlus6128, 4},
		{sna.CPCPlus464, 5},
		{sna.GX4000, 6},
	}
	for _, tt := range tests {
		result := sna.CPCValue(tt.cpc)
		assert.Equal(t, tt.expected, result)
	}
}

func TestCRTCValue(t *testing.T) {
	tests := []struct {
		crtc     sna.CRTC
		expected uint8
	}{
		{sna.HD6845S_UM6845, 0},
		{sna.UM6845R, 1},
		{sna.MC6845, 2},
		{sna.ASIC_6845, 3},
		{sna.Pre_ASIC, 4},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.expected, sna.CRTCValue(tt.crtc))
	}
}

func TestCPCType(t *testing.T) {
	assert.Equal(t, sna.CPC464, sna.CPCType(0))
	assert.Equal(t, sna.CPC664, sna.CPCType(1))
	assert.Equal(t, sna.CPC6128, sna.CPCType(2))
	assert.Equal(t, sna.Unknown, sna.CPCType(3))
	assert.Equal(t, sna.Unknown, sna.CPCType(99))
}

func TestSNAReadWriteRoundtrip(t *testing.T) {
	header := sna.NewSnaHeader()
	original := sna.NewSna(header)
	// write some data
	for i := 0; i < 100; i++ {
		original.Data[i] = byte(i)
	}

	var buf bytes.Buffer
	require.NoError(t, original.Write(&buf))

	loaded := &sna.SNA{}
	require.NoError(t, loaded.Read(bytes.NewReader(buf.Bytes())))

	assert.Equal(t, original.Header.Version, loaded.Header.Version)
	assert.Equal(t, original.Header.RegisterA, loaded.Header.RegisterA)
	assert.Equal(t, original.Data[:100], loaded.Data[:100])
}

func TestSNAV2ReadWriteRoundtrip(t *testing.T) {
	header := sna.NewSnaV2Header()
	original := sna.NewSna(header)

	var buf bytes.Buffer
	require.NoError(t, original.Write(&buf))

	loaded := &sna.SNA{}
	require.NoError(t, loaded.Read(bytes.NewReader(buf.Bytes())))

	assert.Equal(t, uint8(2), loaded.Header.Version)
	assert.Equal(t, uint8(128), loaded.Header.MemoryDumpSize)
}

func TestSNAPutGet(t *testing.T) {
	s := sna.NewSna(sna.NewSnaHeader())

	// Get works independently of Put
	for i := 0; i < 10; i++ {
		s.Data[0x100+i] = byte(i + 1)
	}
	result, err := s.Get(0x100, 10)
	require.NoError(t, err)
	for i := 0; i < 10; i++ {
		assert.Equal(t, byte(i+1), result[i])
	}
}

func TestSNAGetExceedSize(t *testing.T) {
	s := sna.NewSna(sna.NewSnaHeader())
	_, err := s.Get(0xFFF0, 0x1000)
	assert.Error(t, err)
}

func TestSNAPutNoHeaderNoAddress(t *testing.T) {
	s := sna.NewSna(sna.NewSnaHeader())
	err := s.Put([]byte{0x01, 0x02}, 0, 2)
	assert.ErrorIs(t, err, sna.ErrorNoHeaderOrStartAddress)
}

func TestSNACPCTypeString(t *testing.T) {
	s := sna.NewSna(sna.NewSnaHeader())
	assert.Equal(t, "CPC464", s.CPCType())
}

func TestSNACRTCTypeString(t *testing.T) {
	s := sna.NewSna(sna.NewSnaHeader())
	s.Header.CRTCType = 0
	assert.Equal(t, "HD6845S_UM6845", s.CRTCType())
}

func TestSNAHeaderString(t *testing.T) {
	h := sna.NewSnaHeader()
	result := h.String()
	assert.Contains(t, result, "Version:1")
	assert.Contains(t, result, "RegisterA")
}
