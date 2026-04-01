package action

import (
	"os"
	"testing"

	"github.com/jeromelesaux/dsk/dsk"
	"github.com/stretchr/testify/assert"
)

func TestOptionsBuilder(t *testing.T) {
	op := NewOptions().
		WithQuiet(true).
		WithFormat(true).
		WithForce(true).
		WithAnalyze(true).
		WithDataFormat(false).
		WithVendorFormat(true).
		WithStdout(true).
		WithHidden(true).
		WithRemoveHeader(true).
		WithRawImport(true).
		WithRawExport(true)

	assert.True(t, op.quiet)
	assert.True(t, op.format)
	assert.True(t, op.force)
	assert.True(t, op.analyze)
	assert.False(t, op.dataFormat)
	assert.True(t, op.vendorFormat)
	assert.True(t, op.stdout)
	assert.True(t, op.hidden)
	assert.True(t, op.removeHeader)
	assert.True(t, op.rawImport)
	assert.True(t, op.rawExport)
}

func TestDskDescriptorDefaultsAndBuilder(t *testing.T) {
	d := NewDskDescriptor()
	assert.Equal(t, 9, d.Sector)
	assert.Equal(t, 39, d.Track)
	assert.Equal(t, 2, d.Head)
	assert.Equal(t, dsk.DataFormat, d.Type)
	assert.Equal(t, "./", d.FolderPath)

	d = d.WithSector(13).WithTrack(10).WithHead(1).WithPath("/tmp").WithType(1)
	assert.Equal(t, 13, d.Sector)
	assert.Equal(t, 10, d.Track)
	assert.Equal(t, 1, d.Head)
	assert.Equal(t, "/tmp", d.Path)
	assert.Equal(t, 1, d.Type)
}

func TestAmsdosFileDescriptorDefaultsAndBuilder(t *testing.T) {
	a := NewAmsdosFileDescriptor()
	assert.Equal(t, AmsdosTypeAscii, a.Type)

	a = a.WithAddHeader(true).WithType(AmsdosTypeBinary).WithUser(5).WithPath("hello")
	a.WithExec(0x1234)
	a.WithLoad(0x5678)
	assert.True(t, a.addHeader)
	assert.Equal(t, AmsdosTypeBinary, a.Type)
	assert.Equal(t, uint16(5), a.User)
	assert.Equal(t, uint16(0x1234), a.Exec)
	assert.Equal(t, uint16(0x5678), a.Load)
	assert.Equal(t, "hello", a.Path)
}

func TestAmsdosFileDescriptorAddExecAddLoad(t *testing.T) {
	a := NewAmsdosFileDescriptor()
	a.AddExec("#100").AddLoad("0x200")
	assert.Equal(t, uint16(0x100), a.Exec)
	assert.Equal(t, uint16(0x200), a.Load)
	assert.Equal(t, AmsdosTypeBinary, a.Type)
	assert.True(t, a.addHeader)
}

func TestAmsdosFileDescriptorAddExecAddLoadInvalid(t *testing.T) {
	a := NewAmsdosFileDescriptor()
	a.AddExec("#GGGG")
	assert.Equal(t, uint16(0), a.Exec)
	a.AddLoad("0xZZZZ")
	assert.Equal(t, uint16(0), a.Load)
}

func TestActionNewAndDskIsSet(t *testing.T) {
	a := NewAction("foo.dsk")
	assert.Equal(t, "foo.dsk", a.Path)
	assert.True(t, a.DskIsSet())

	a2 := NewAction("", "")
	assert.False(t, a2.DskIsSet())
}

func TestSnaActionBuilders(t *testing.T) {
	s := NewSnaAction("test.sna").
		WithFiles("file1", "file2").
		WithCPCType(4).
		WithVersion(2).
		WithScreemode(1).
		WithSnaFormatAction(true).
		WithSnaHexaListAction(false).
		WithSnaPutAction(true).
		WithSnaGetAction(false).
		WithSnaInfoAction(true)

	assert.Equal(t, "test.sna", s.Path)
	assert.Equal(t, "file2", s.File)
	assert.Equal(t, 4, s.CPCType)
	assert.Equal(t, 2, s.Version)
	assert.Equal(t, 1, s.Screenmode)
}

func TestFormatSnaAndInfoSna(t *testing.T) {
	tmpfile := t.TempDir() + "/test.sna"

	onError, _, _ := FormatSna(tmpfile, 1)
	assert.False(t, onError)

	onError, _, _ = InfoSna(tmpfile)
	assert.False(t, onError)

	// invalid version path: should fail
	onError, msg, _ := FormatSna(t.TempDir()+"/bad.sna", 99)
	assert.True(t, onError)
	assert.Contains(t, msg, "Cannot create Sna file")
}

func TestSnaActionDoSnaActionsFormat(t *testing.T) {
	tmpfile := t.TempDir() + "/test2.sna"
	s := NewSnaAction(tmpfile).WithVersion(1).WithSnaFormatAction(true)
	onError, _, _ := s.DoSnaActions()
	assert.False(t, onError)

	// The SNA file should have been created
	exists := false
	if _, err := os.Stat(tmpfile); err == nil {
		exists = true
	}
	assert.True(t, exists)
}

