package dsk_test

import (
	"testing"

	"github.com/jeromelesaux/dsk"
	"github.com/stretchr/testify/assert"
)

func TestMemoryChunck(t *testing.T) {
	t.Run("uncompress", func(t *testing.T) {
		m := &dsk.MemChunck{}
		buf := []byte{0xe5, 0xff, 0x2}
		m.Feed(buf)
		got := m.Data[0:0xff]
		for i := 0; i < len(got); i++ {
			assert.Equal(t, byte(2), got[i])
		}
	})

	t.Run("compress", func(t *testing.T) {
		m := &dsk.MemChunck{}
		for i := 0; i < 0xff; i++ {
			m.Data[i] = 0x2
		}
		got := m.Export()[0:3]
		expected := []byte{0xe5, 0xff, 0x2}
		assert.Equal(t, expected, got)
	})
}
