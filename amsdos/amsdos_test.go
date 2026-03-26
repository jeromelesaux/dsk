package amsdos

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/jeromelesaux/m4client/cpc"
	"github.com/stretchr/testify/assert"
)

func TestCheckAmsdos(t *testing.T) {
	// Create a valid AMSDOS header
	header := &cpc.CpcHead{
		User:       0,
		Size:       100,
		Size2:      100,
		LogicalSize: 100,
		Type:       2, // binary
		Address:    0x4000,
		Exec:       0x4000,
	}
	header.Checksum = header.ComputedChecksum16()

	var buf bytes.Buffer
	err := binary.Write(&buf, binary.LittleEndian, header)
	assert.NoError(t, err)

	// Test valid header
	valid, parsedHeader := CheckAmsdos(buf.Bytes())
	assert.True(t, valid)
	assert.Equal(t, header, parsedHeader)

	// Test invalid header (wrong checksum)
	header.Checksum = 0
	buf.Reset()
	err = binary.Write(&buf, binary.LittleEndian, header)
	assert.NoError(t, err)

	valid, _ = CheckAmsdos(buf.Bytes())
	assert.False(t, valid)

	// Test empty buffer
	valid, _ = CheckAmsdos([]byte{})
	assert.False(t, valid)
}