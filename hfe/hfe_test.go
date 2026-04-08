package hfe

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"

	extdsk "github.com/jeromelesaux/dsk/dsk"
	"github.com/stretchr/testify/require"
)

// --- helpers ---

func makeDSK(numTracks, numSides int) *extdsk.DSK {
	d := extdsk.FormatDsk(9, uint8(numTracks), uint8(numSides), extdsk.DataFormat, extdsk.DSK_TYPE)
	for t := 0; t < len(d.Tracks); t++ {
		for i := range d.Tracks[t].Data {
			d.Tracks[t].Data[i] = byte((t + i) & 0xFF)
		}
	}
	return d
}

func writeHFE(t *testing.T, d *extdsk.DSK) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.hfe")
	if err := FromDSK(d, path); err != nil {
		t.Fatalf("FromDSK failed: %v", err)
	}
	return path
}

// --- crc16 ---

func TestCRC16_Deterministic(t *testing.T) {
	data := []byte{0xFE, 0x00, 0x00, 0x02}
	if crc16(data) != crc16(data) {
		t.Error("crc16 is not deterministic")
	}
}

func TestCRC16_EmptyInput(t *testing.T) {
	if got := crc16(nil); got != 0xFFFF {
		t.Errorf("crc16(nil) = 0x%04X, want 0xFFFF", got)
	}
}

func TestCRC16_DifferentInputs(t *testing.T) {
	if crc16([]byte{0x00}) == crc16([]byte{0xFF}) {
		t.Error("crc16 collision on different inputs")
	}
}

// --- mfmEncode / mfmDecode ---

func TestMFMRoundTrip(t *testing.T) {
	cases := [][]byte{
		{0x00},
		{0xFF},
		{0xAA, 0x55},
		{0x4E, 0x4E, 0x4E},
		{0xA1, 0xFE, 0xFB},
		make([]byte, 32),
	}
	for _, input := range cases {
		encoded := mfmEncode(input)
		decoded := mfmDecode(encoded)
		if len(decoded) < len(input) {
			t.Errorf("decoded too short for input %v", input)
			continue
		}
		if !bytes.Equal(decoded[:len(input)], input) {
			t.Errorf("round-trip failed for %v: got %v", input, decoded[:len(input)])
		}
	}
}

func TestMFMEncode_OutputSize(t *testing.T) {
	input := make([]byte, 10)
	// 10 bytes * 8 bits * 2 (clock+data) = 160 bits = 20 bytes
	if got := len(mfmEncode(input)); got != 20 {
		t.Errorf("expected 20 bytes, got %d", got)
	}
}

func TestMFMDecode_Empty(t *testing.T) {
	if got := mfmDecode(nil); len(got) != 0 {
		t.Errorf("expected empty, got %d bytes", len(got))
	}
}

// --- interleave ---

func TestInterleave_Layout(t *testing.T) {
	side0 := bytes.Repeat([]byte{0xAA}, 256)
	side1 := bytes.Repeat([]byte{0xBB}, 256)
	out := interleave(side0, side1)

	if len(out) != blockSize {
		t.Fatalf("expected %d bytes, got %d", blockSize, len(out))
	}
	for i := 0; i < 256; i++ {
		if out[i] != 0xAA {
			t.Errorf("side0 byte %d: got 0x%02X", i, out[i])
			break
		}
	}
	for i := 256; i < 512; i++ {
		if out[i] != 0xBB {
			t.Errorf("side1 byte %d: got 0x%02X", i, out[i])
			break
		}
	}
}

func TestInterleave_MultiBlock(t *testing.T) {
	side0 := make([]byte, 512)
	side1 := make([]byte, 512)
	if got := len(interleave(side0, side1)); got != 2*blockSize {
		t.Errorf("expected %d, got %d", 2*blockSize, got)
	}
}

func TestInterleave_AsymmetricSides(t *testing.T) {
	side0 := make([]byte, 256)
	side1 := make([]byte, 512)
	out := interleave(side0, side1)
	if len(out) != 2*blockSize {
		t.Errorf("expected %d, got %d", 2*blockSize, len(out))
	}
}

// --- countSectors ---

func TestCountSectors_None(t *testing.T) {
	if n := countSectors(make([]byte, 100)); n != 0 {
		t.Errorf("expected 0, got %d", n)
	}
}

func TestCountSectors_One(t *testing.T) {
	raw := []byte{0xA1, 0xA1, 0xA1, 0xFE, 0x00, 0x00, 0x00, 0x00}
	if n := countSectors(raw); n != 1 {
		t.Errorf("expected 1, got %d", n)
	}
}

func TestCountSectors_Nine(t *testing.T) {
	var raw []byte
	for i := 0; i < 9; i++ {
		raw = append(raw, 0xA1, 0xA1, 0xA1, 0xFE, 0x00, 0x00, 0x00, 0x00)
		raw = append(raw, make([]byte, 512)...)
	}
	if n := countSectors(raw); n != 9 {
		t.Errorf("expected 9, got %d", n)
	}
}

// --- extractSectorData ---

func TestExtractSectorData_Empty(t *testing.T) {
	if got := extractSectorData(nil); len(got) != 0 {
		t.Errorf("expected empty, got %d bytes", len(got))
	}
}

func TestExtractSectorData_OneSector(t *testing.T) {
	sectorData := bytes.Repeat([]byte{0x42}, 512)
	var raw []byte
	raw = append(raw, 0xA1, 0xA1, 0xA1, 0xFE)
	raw = append(raw, 0x00, 0x00, 0x01, 0x02) // C H R N (N=2 → 512 bytes)
	raw = append(raw, 0x00, 0x00)             // CRC
	raw = append(raw, 0xA1, 0xA1, 0xA1, 0xFB)
	raw = append(raw, sectorData...)

	got := extractSectorData(raw)
	if !bytes.Equal(got, sectorData) {
		t.Errorf("sector data mismatch: got %d bytes", len(got))
	}
}

// --- buildMFMTrack ---

func TestBuildMFMTrack_MinLength(t *testing.T) {
	d := makeDSK(1, 1)
	mfm := buildMFMTrack(d.Tracks[0])
	// mfmEncode(6250 raw bytes) = 12500 MFM bytes minimum
	if len(mfm) < 12500 {
		t.Errorf("MFM track too short: %d bytes", len(mfm))
	}
}

func TestBuildMFMTrack_ContainsSectors(t *testing.T) {
	d := makeDSK(1, 1)
	mfm := buildMFMTrack(d.Tracks[0])
	decoded := mfmDecode(mfm)
	n := countSectors(decoded)
	if n != int(d.Tracks[0].NbSect) {
		t.Errorf("expected %d sectors in MFM stream, got %d", d.Tracks[0].NbSect, n)
	}
}

func TestBuildMFMTrack_DataRoundTrip(t *testing.T) {
	d := makeDSK(1, 1)
	// fill with known pattern
	for i := range d.Tracks[0].Data {
		d.Tracks[0].Data[i] = byte(i & 0xFF)
	}
	mfm := buildMFMTrack(d.Tracks[0])
	decoded := mfmDecode(mfm)
	recovered := extractSectorData(decoded)

	if !bytes.Equal(recovered, d.Tracks[0].Data) {
		t.Errorf("sector data round-trip failed: got %d bytes, want %d", len(recovered), len(d.Tracks[0].Data))
	}
}

// --- FromDSK ---

func TestFromDSK_CreatesFile(t *testing.T) {
	path := writeHFE(t, makeDSK(2, 1))
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("output file not found: %v", err)
	}
	if info.Size() < int64(blockSize) {
		t.Errorf("output too small: %d bytes", info.Size())
	}
}

func TestFromDSK_ValidSignature(t *testing.T) {
	path := writeHFE(t, makeDSK(2, 1))
	raw, _ := os.ReadFile(path)
	if string(raw[:8]) != "HXCPICFE" {
		t.Errorf("bad signature: %q", string(raw[:8]))
	}
}

func TestFromDSK_HeaderFields(t *testing.T) {
	d := makeDSK(3, 1)
	path := writeHFE(t, d)
	raw, _ := os.ReadFile(path)

	if raw[9] != 3 {
		t.Errorf("NumTracks: expected 3, got %d", raw[9])
	}
	if raw[10] != 1 {
		t.Errorf("NumSides: expected 1, got %d", raw[10])
	}
	if binary.LittleEndian.Uint16(raw[12:14]) != 250 {
		t.Errorf("BitRate: expected 250")
	}
	if binary.LittleEndian.Uint16(raw[14:16]) != 300 {
		t.Errorf("RPM: expected 300")
	}
}

func TestFromDSK_DoubleSided(t *testing.T) {
	path := writeHFE(t, makeDSK(2, 2))
	raw, _ := os.ReadFile(path)
	if raw[10] != 2 {
		t.Errorf("NumSides: expected 2, got %d", raw[10])
	}
}

func TestFromDSK_LUTAtBlock1(t *testing.T) {
	path := writeHFE(t, makeDSK(2, 1))
	raw, _ := os.ReadFile(path)
	lutOffset := binary.LittleEndian.Uint16(raw[18:20])
	if lutOffset != 1 {
		t.Errorf("LUT offset: expected 1, got %d", lutOffset)
	}
}

func TestFromDSK_FileSizeMultipleOfBlockSize(t *testing.T) {
	path := writeHFE(t, makeDSK(4, 1))
	info, _ := os.Stat(path)
	if info.Size()%int64(blockSize) != 0 {
		t.Errorf("file size %d is not a multiple of %d", info.Size(), blockSize)
	}
}

func TestFromDSK_InvalidPath(t *testing.T) {
	if err := FromDSK(makeDSK(1, 1), "/nonexistent/dir/out.hfe"); err == nil {
		t.Fatal("expected error for invalid path")
	}
}

// --- Open ---

func TestOpen_FileNotFound(t *testing.T) {
	if _, err := Open("/nonexistent/path.hfe"); err == nil {
		t.Fatal("expected error")
	}
}

func TestOpen_FileTooSmall(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tiny.hfe")
	os.WriteFile(path, make([]byte, 10), 0644)
	if _, err := Open(path); err == nil {
		t.Fatal("expected error for small file")
	}
}

func TestOpen_InvalidSignature(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bad.hfe")
	os.WriteFile(path, make([]byte, 1024), 0644)
	if _, err := Open(path); err == nil {
		t.Fatal("expected error for invalid signature")
	}
}

func TestOpen_ValidFile(t *testing.T) {
	path := writeHFE(t, makeDSK(2, 1))
	h, err := Open(path)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	if h.Header.NumTracks != 2 {
		t.Errorf("NumTracks: expected 2, got %d", h.Header.NumTracks)
	}
	if h.Header.NumSides != 1 {
		t.Errorf("NumSides: expected 1, got %d", h.Header.NumSides)
	}
	if len(h.Entries) != 2 {
		t.Errorf("LUT entries: expected 2, got %d", len(h.Entries))
	}
	if len(h.Tracks) != 2 {
		t.Errorf("Tracks: expected 2, got %d", len(h.Tracks))
	}
}

func TestOpen_TrackDataNotEmpty(t *testing.T) {
	path := writeHFE(t, makeDSK(2, 1))
	h, _ := Open(path)
	if len(h.Tracks[0].Side0) == 0 {
		t.Error("Side0 of track 0 is empty")
	}
}

// --- ToDSK ---

func TestToDSK_TrackCount(t *testing.T) {
	path := writeHFE(t, makeDSK(3, 1))
	h, _ := Open(path)
	d, err := h.ToDSK()
	if err != nil {
		t.Fatalf("ToDSK failed: %v", err)
	}
	if int(d.Entry.NbTracks) != 3 {
		t.Errorf("NbTracks: expected 3, got %d", d.Entry.NbTracks)
	}
}

func TestToDSK_SectorCount(t *testing.T) {
	path := writeHFE(t, makeDSK(2, 1))
	h, _ := Open(path)
	d, err := h.ToDSK()
	if err != nil {
		t.Fatalf("ToDSK failed: %v", err)
	}
	if d.Tracks[0].NbSect != 9 {
		t.Errorf("NbSect: expected 9, got %d", d.Tracks[0].NbSect)
	}
}

// --- HFE → DSK avec fichiers réels ---

func TestBRUTAL_HFEToDSK(t *testing.T) {
	h, err := Open(filepath.Join("testdata", "BRUTAL.HFE"))
	if err != nil {
		t.Fatalf("Open BRUTAL.HFE failed: %v", err)
	}

	converted, err := h.ToDSK()
	if err != nil {
		t.Fatalf("ToDSK failed: %v", err)
	}

	expected, err := extdsk.ReadDsk(filepath.Join("testdata", "BRUTAL.DSK"))
	if err != nil {
		t.Fatalf("ReadDsk BRUTAL.DSK failed: %v", err)
	}

	if converted.Entry.NbTracks != expected.Entry.NbTracks {
		t.Errorf("NbTracks: got %d, want %d", converted.Entry.NbTracks, expected.Entry.NbTracks)
	}
	if converted.Entry.NbHeads != expected.Entry.NbHeads {
		t.Errorf("NbHeads: got %d, want %d", converted.Entry.NbHeads, expected.Entry.NbHeads)
	}

	for i := 0; i < int(expected.Entry.NbTracks); i++ {
		if i >= len(converted.Tracks) {
			t.Errorf("track %d missing in converted DSK", i)
			continue
		}
		if converted.Tracks[i].NbSect != expected.Tracks[i].NbSect {
			t.Errorf("track %d: NbSect got %d, want %d", i, converted.Tracks[i].NbSect, expected.Tracks[i].NbSect)
		}
		if !bytes.Equal(converted.Tracks[i].Data, expected.Tracks[i].Data) {
			t.Errorf("track %d: sector data mismatch", i)
		}
	}

	f, err := os.Create("test.dsk")
	require.NoError(t, err)
	converted.Write(f)
	f.Close()
}

// --- Round-trip DSK → HFE → DSK (fichier réel) ---

func TestRoundTrip_EmptyDSK_File(t *testing.T) {
	const srcDSK = "empty.dsk"

	// 1. Lire le DSK original
	original, err := extdsk.ReadDsk(srcDSK)
	if err != nil {
		t.Fatalf("ReadDsk(%q) failed: %v", srcDSK, err)
	}

	dir := t.TempDir()
	hfePath := filepath.Join(dir, "empty.hfe")
	recoveredPath := filepath.Join(dir, "recovered.dsk")

	// 2. Convertir DSK → HFE
	if err := FromDSK(original, hfePath); err != nil {
		t.Fatalf("FromDSK failed: %v", err)
	}

	// 3. Relire le HFE et convertir → DSK
	h, err := Open(hfePath)
	if err != nil {
		t.Fatalf("Open HFE failed: %v", err)
	}
	recovered, err := h.ToDSK()
	if err != nil {
		t.Fatalf("ToDSK failed: %v", err)
	}

	// 4. Sauvegarder le DSK récupéré
	if err := extdsk.WriteDsk(recoveredPath, recovered); err != nil {
		t.Fatalf("WriteDsk failed: %v", err)
	}

	// 5. Comparer les métadonnées
	if recovered.Entry.NbTracks != original.Entry.NbTracks {
		t.Errorf("NbTracks: got %d, want %d", recovered.Entry.NbTracks, original.Entry.NbTracks)
	}
	if recovered.Entry.NbHeads != original.Entry.NbHeads {
		t.Errorf("NbHeads: got %d, want %d", recovered.Entry.NbHeads, original.Entry.NbHeads)
	}

	// 6. Comparer les données secteur piste par piste
	for i := 0; i < int(original.Entry.NbTracks); i++ {
		if i >= len(recovered.Tracks) {
			t.Errorf("track %d missing in recovered DSK", i)
			continue
		}
		if recovered.Tracks[i].NbSect != original.Tracks[i].NbSect {
			t.Errorf("track %d: NbSect got %d, want %d", i, recovered.Tracks[i].NbSect, original.Tracks[i].NbSect)
		}
		if !bytes.Equal(recovered.Tracks[i].Data, original.Tracks[i].Data) {
			t.Errorf("track %d: sector data mismatch", i)
		}
	}

	f, err := os.Create("test.dsk")
	require.NoError(t, err)
	recovered.Write(f)
	f.Close()
}

// --- Round-trip DSK → HFE → DSK ---

func TestRoundTrip_FromDSK_ToDSK(t *testing.T) {
	original := makeDSK(2, 1)
	path := writeHFE(t, original)

	h, err := Open(path)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	recovered, err := h.ToDSK()
	if err != nil {
		t.Fatalf("ToDSK failed: %v", err)
	}

	if recovered.Entry.NbTracks != original.Entry.NbTracks {
		t.Errorf("NbTracks: %d vs %d", recovered.Entry.NbTracks, original.Entry.NbTracks)
	}

	for i := 0; i < int(original.Entry.NbTracks); i++ {
		if !bytes.Equal(recovered.Tracks[i].Data, original.Tracks[i].Data) {
			t.Errorf("track %d data mismatch", i)
		}
	}
}
