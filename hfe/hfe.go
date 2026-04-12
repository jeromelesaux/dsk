package hfe

import (
	"encoding/binary"
	"fmt"
	"os"

	extdsk "github.com/jeromelesaux/dsk/dsk"
)

const blockSize = 512

type Header struct {
	Signature       string
	FormatRevision  byte
	NumTracks       byte
	NumSides        byte
	TrackEncoding   byte
	BitRate         uint16
	FloppyRPM       uint16
	FloppyInterface byte
	MCUVersion      byte // byte at index 17
	TrackListOffset uint16
}

type TrackEntry struct {
	Offset uint16 // in blocks of 512 bytes
	Length uint16 // MFM stream length in bytes
}

type Track struct {
	Side0 []byte
	Side1 []byte
}

type HFE struct {
	Header  Header
	Entries []TrackEntry
	Tracks  []Track
}

func Open(path string) (*HFE, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(data) < blockSize {
		return nil, fmt.Errorf("file too small")
	}
	if string(data[:8]) != "HXCPICFE" {
		return nil, fmt.Errorf("invalid HFE signature")
	}

	h := &HFE{}
	h.Header = Header{
		Signature:       string(data[:8]),
		FormatRevision:  data[8],
		NumTracks:       data[9],
		NumSides:        data[10],
		TrackEncoding:   data[11],
		BitRate:         binary.LittleEndian.Uint16(data[12:14]),
		FloppyRPM:       binary.LittleEndian.Uint16(data[14:16]),
		FloppyInterface: data[16],
		MCUVersion:      data[17],
		TrackListOffset: binary.LittleEndian.Uint16(data[18:20]),
	}

	lutOffset := int(h.Header.TrackListOffset) * blockSize
	numTracks := int(h.Header.NumTracks)

	for i := range numTracks {
		off := lutOffset + i*4
		if off+4 > len(data) {
			return nil, fmt.Errorf("LUT entry %d out of bounds", i)
		}
		h.Entries = append(h.Entries, TrackEntry{
			Offset: binary.LittleEndian.Uint16(data[off:]),
			Length: binary.LittleEndian.Uint16(data[off+2:]),
		})
	}

	numSides := int(h.Header.NumSides)
	if numSides < 1 {
		numSides = 1
	}

	for _, entry := range h.Entries {
		trackOffset := int(entry.Offset) * blockSize
		// entry.Length is the total interleaved length; per-side = Length/2
		perSideLen := int(entry.Length) / 2
		nBlocks := (int(entry.Length) + 255) / 256

		side0 := make([]byte, 0, nBlocks*256)
		side1 := make([]byte, 0, nBlocks*256)

		for i := range nBlocks {
			start := trackOffset + i*blockSize
			if start+blockSize > len(data) {
				break
			}
			side0 = append(side0, data[start:start+256]...)
			side1 = append(side1, data[start+256:start+blockSize]...)
		}

		// truncate to actual per-side MFM length
		if perSideLen < len(side0) {
			side0 = side0[:perSideLen]
		}
		if perSideLen < len(side1) {
			side1 = side1[:perSideLen]
		}

		t := Track{Side0: side0}
		if numSides > 1 {
			t.Side1 = side1
		}
		h.Tracks = append(h.Tracks, t)
	}

	return h, nil
}

// reverseBits reverses the bit order within a byte (LSB↔MSB)
func reverseBits(b byte) byte {
	var r byte
	for range 8 {
		r = (r << 1) | (b & 1)
		b >>= 1
	}
	return r
}

// mfmDecode extracts data bits from an HFE MFM bitstream.
// HFE stores bits LSB-first per byte: bit 0 is clock, bit 1 is data, etc.
func mfmDecode(mfm []byte) []byte {
	totalBits := len(mfm) * 8
	out := make([]byte, 0, totalBits/16)
	var current byte
	bitCount := 0
	for i := 1; i < totalBits; i += 2 { // start at 1: skip clock bits, take data bits
		byteIdx := i / 8
		bitPos := i % 8 // LSB-first
		if byteIdx >= len(mfm) {
			break
		}
		bit := (mfm[byteIdx] >> uint(bitPos)) & 1
		current = (current << 1) | bit
		bitCount++
		if bitCount == 8 {
			out = append(out, current)
			current = 0
			bitCount = 0
		}
	}
	return out
}

// extractSectorData scans a decoded MFM byte stream and returns the raw sector data in order
func extractSectorData(raw []byte) []byte {
	var result []byte
	i := 0
	for i < len(raw)-10 {
		// IDAM: A1 A1 A1 FE
		if raw[i] == 0xA1 && i+3 < len(raw) && raw[i+1] == 0xA1 && raw[i+2] == 0xA1 && raw[i+3] == 0xFE {
			i += 4
			if i+4 > len(raw) {
				break
			}
			sectorN := raw[i+3]
			sectorSize := int(128) << sectorN
			i += 4 + 2 // skip C H R N + CRC

			// find DAM: A1 A1 A1 FB/F8
			found := false
			for j := 0; j < 60 && i+j+3 < len(raw); j++ {
				if raw[i+j] == 0xA1 && raw[i+j+1] == 0xA1 && raw[i+j+2] == 0xA1 &&
					(raw[i+j+3] == 0xFB || raw[i+j+3] == 0xF8) {
					i = i + j + 4
					found = true
					break
				}
			}
			if !found {
				continue
			}
			if i+sectorSize <= len(raw) {
				result = append(result, raw[i:i+sectorSize]...)
			}
			i += sectorSize
		} else {
			i++
		}
	}
	return result
}

// ToDSK converts the HFE image into a *dsk.DSK using the external jeromelesaux/dsk package.
// It decodes the MFM bitstream of each track and injects the sector data.
func (h *HFE) ToDSK() (*extdsk.DSK, error) {
	numTracks := int(h.Header.NumTracks)
	numSides := max(int(h.Header.NumSides), 1)

	// Detect sector count from first track
	nbSect := uint8(9)
	if numTracks > 0 && len(h.Tracks) > 0 {
		raw := mfmDecode(h.Tracks[0].Side0)
		count := countSectors(raw)
		if count > 0 {
			nbSect = uint8(count)
		}
	}

	d := extdsk.FormatDsk(nbSect, uint8(numTracks), uint8(numSides), extdsk.DataFormat, extdsk.DSK_TYPE)

	for t := 0; t < numTracks && t < len(h.Tracks); t++ {
		raw0 := mfmDecode(h.Tracks[t].Side0)
		sectorData := extractSectorData(raw0)
		if len(sectorData) > 0 && len(sectorData) <= len(d.Tracks[t].Data) {
			copy(d.Tracks[t].Data, sectorData)
		}

		if numSides > 1 && len(h.Tracks[t].Side1) > 0 {
			raw1 := mfmDecode(h.Tracks[t].Side1)
			sectorData1 := extractSectorData(raw1)
			idx1 := t*2 + 1
			if idx1 < len(d.Tracks) && len(sectorData1) > 0 && len(sectorData1) <= len(d.Tracks[idx1].Data) {
				copy(d.Tracks[idx1].Data, sectorData1)
			}
		}
	}

	return d, nil
}

// mfmEncode encodes a raw byte stream into an HFE MFM bitstream (LSB-first per byte).
func mfmEncode(data []byte) []byte {
	var bits []uint8 // flat list of bits in transmission order
	prevBit := byte(0)
	for _, b := range data {
		for i := 7; i >= 0; i-- {
			dataBit := (b >> uint(i)) & 1
			clockBit := byte(0)
			if dataBit == 0 && prevBit == 0 {
				clockBit = 1
			}
			bits = append(bits, clockBit, dataBit)
			prevBit = dataBit
		}
	}
	// pack bits LSB-first into bytes
	out := make([]byte, (len(bits)+7)/8)
	for i, b := range bits {
		if b != 0 {
			out[i/8] |= 1 << uint(i%8) // LSB-first
		}
	}
	return out
}

func crc16(data []byte) uint16 {
	crc := uint16(0xFFFF)
	for _, b := range data {
		crc ^= uint16(b) << 8
		for range 8 {
			if crc&0x8000 != 0 {
				crc = (crc << 1) ^ 0x1021
			} else {
				crc <<= 1
			}
		}
	}
	return crc
}

// buildMFMTrack encodes a DSK track (sectors + data) into a raw MFM byte stream
func buildMFMTrack(track extdsk.CPCEMUTrack) []byte {
	var raw []byte

	// GAP4a
	for range 80 {
		raw = append(raw, 0x4E)
	}
	// Sync
	for range 12 {
		raw = append(raw, 0x00)
	}
	// IAM
	raw = append(raw, 0xC2, 0xC2, 0xC2, 0xFC)
	// GAP1
	for range 50 {
		raw = append(raw, 0x4E)
	}

	dataOffset := 0
	for s := 0; s < int(track.NbSect); s++ {
		sec := track.Sect[s]
		sectorSize := int(sec.SizeByte)
		if sectorSize == 0 {
			sectorSize = int(128) << sec.N
		}

		// Sync
		for range 12 {
			raw = append(raw, 0x00)
		}
		// IDAM
		raw = append(raw, 0xA1, 0xA1, 0xA1, 0xFE)
		idam := []byte{sec.C, sec.H, sec.R, sec.N}
		raw = append(raw, idam...)
		crc := crc16(append([]byte{0xA1, 0xA1, 0xA1, 0xFE}, idam...))
		raw = append(raw, byte(crc>>8), byte(crc))
		// GAP2
		for range 22 {
			raw = append(raw, 0x4E)
		}
		for range 12 {
			raw = append(raw, 0x00)
		}
		// DAM
		raw = append(raw, 0xA1, 0xA1, 0xA1, 0xFB)

		// sector data
		sdata := make([]byte, sectorSize)
		end := dataOffset + sectorSize
		if end > len(track.Data) {
			end = len(track.Data)
		}
		if dataOffset < len(track.Data) {
			copy(sdata, track.Data[dataOffset:end])
		}
		raw = append(raw, sdata...)
		dataOffset += sectorSize

		crc = crc16(append([]byte{0xA1, 0xA1, 0xA1, 0xFB}, sdata...))
		raw = append(raw, byte(crc>>8), byte(crc))
		// GAP3
		for i := 0; i < 54; i++ {
			raw = append(raw, 0x4E)
		}
	}
	// GAP4b
	for len(raw) < 6254 {
		raw = append(raw, 0x4E)
	}
	return mfmEncode(raw)
}

// interleave merges side0 and side1 into 512-byte blocks (256 per side)
func interleave(side0, side1 []byte) []byte {
	nBlocks := (len(side0) + 255) / 256
	if n := (len(side1) + 255) / 256; n > nBlocks {
		nBlocks = n
	}
	out := make([]byte, nBlocks*blockSize)
	for i := 0; i < nBlocks; i++ {
		for j := range 256 {
			if i*256+j < len(side0) {
				out[i*blockSize+j] = side0[i*256+j]
			}
			if i*256+j < len(side1) {
				out[i*blockSize+256+j] = side1[i*256+j]
			}
		}
	}
	return out
}

// FromDSK converts a *extdsk.DSK into an HFE file written at path.
// If header is provided, it uses those values instead of defaults.
func FromDSK(d *extdsk.DSK, path string, header ...*Header) error {
	numTracks := int(d.Entry.NbTracks)
	numSides := max(int(d.Entry.NbHeads), 1)

	type trackData struct {
		interleaved []byte
		mfmLen      uint16
	}
	tracks := make([]trackData, numTracks)

	for t := range numTracks {
		var side0, side1 []byte

		idx0 := t * numSides
		if idx0 < len(d.Tracks) {
			side0 = buildMFMTrack(d.Tracks[idx0])
		} else {
			side0 = mfmEncode(make([]byte, 6250))
		}

		if numSides > 1 {
			idx1 := t*numSides + 1
			if idx1 < len(d.Tracks) {
				side1 = buildMFMTrack(d.Tracks[idx1])
			} else {
				side1 = mfmEncode(make([]byte, 6250))
			}
		} else {
			side1 = make([]byte, len(side0))
		}

		mfmLen := max(len(side1), len(side0))
		tracks[t] = trackData{
			interleaved: interleave(side0, side1),
			mfmLen:      uint16(mfmLen * 2), // total interleaved length (both sides)
		}
	}

	// LUT at block 1
	lutBlocks := (numTracks*4 + blockSize - 1) / blockSize
	dataStartBlock := 1 + lutBlocks

	lut := make([]byte, lutBlocks*blockSize)
	currentBlock := dataStartBlock
	for t := range numTracks {
		binary.LittleEndian.PutUint16(lut[t*4:], uint16(currentBlock))
		binary.LittleEndian.PutUint16(lut[t*4+2:], tracks[t].mfmLen)
		currentBlock += (len(tracks[t].interleaved) + blockSize - 1) / blockSize
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	// Header block
	hdr := make([]byte, blockSize)
	for i := range hdr {
		hdr[i] = 0xff
	}
	copy(hdr[:8], "HXCPICFE")
	if len(header) > 0 {
		h := header[0]
		hdr[8] = h.FormatRevision
		hdr[9] = h.NumTracks
		hdr[10] = h.NumSides
		hdr[11] = h.TrackEncoding
		binary.LittleEndian.PutUint16(hdr[12:], h.BitRate)
		binary.LittleEndian.PutUint16(hdr[14:], h.FloppyRPM)
		hdr[16] = h.FloppyInterface
		hdr[17] = h.MCUVersion
		binary.LittleEndian.PutUint16(hdr[18:], h.TrackListOffset)
	} else {
		hdr[8] = 0 // revision
		hdr[9] = byte(numTracks)
		hdr[10] = byte(numSides)
		hdr[11] = 0                                   // ISOIBM_MFM
		binary.LittleEndian.PutUint16(hdr[12:], 0xFA) // bitrate kbps
		binary.LittleEndian.PutUint16(hdr[14:], 0)    // RPM
		hdr[16] = 0x06                                // generic shugart
		hdr[17] = 1                                   // MCU version
		binary.LittleEndian.PutUint16(hdr[18:], 1)    // LUT at block 1
	}

	if _, err := f.Write(hdr); err != nil {
		return err
	}
	if _, err := f.Write(lut); err != nil {
		return err
	}
	for t := range numTracks {
		padded := tracks[t].interleaved
		if rem := len(padded) % blockSize; rem != 0 {
			padded = append(padded, make([]byte, blockSize-rem)...)
		}
		if _, err := f.Write(padded); err != nil {
			return err
		}
	}
	return nil
}

func countSectors(raw []byte) int {
	count := 0
	for i := 0; i < len(raw)-3; i++ {
		if raw[i] == 0xA1 && raw[i+1] == 0xA1 && raw[i+2] == 0xA1 && raw[i+3] == 0xFE {
			count++
			i += 3
		}
	}
	return count
}
