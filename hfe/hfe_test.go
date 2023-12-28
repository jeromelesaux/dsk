package hfe_test

import (
	"os"
	"testing"

	"github.com/jeromelesaux/dsk/hfe"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadHfe(t *testing.T) {
	t.Run("read_header", func(t *testing.T) {
		f, err := os.Open("../testdata/fighter_bomber.hfe")
		require.NoError(t, err)

		h, err := hfe.ReadHeader(f)
		assert.NoError(t, err)
		assert.NotEmpty(t, h)
	})

	t.Run("read_hfe", func(t *testing.T) {
		f, err := os.Open("../testdata/fighter_bomber.hfe")
		require.NoError(t, err)

		h, err := hfe.Read(f)
		assert.NoError(t, err)
		assert.NotEmpty(t, h)
		assert.Equal(t, hfe.CPC_DD_FLOPPYMODE, h.Header.FloppyInterfaceMode)
	})

}
