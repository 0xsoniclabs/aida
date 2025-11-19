// Copyright 2025 Sonic Labs
// This file is part of Aida Testing Infrastructure for Sonic
//
// Aida is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Aida is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with Aida. If not, see <http://www.gnu.org/licenses/>.

package coverage

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

var (
	counterFileMagic = [4]byte{0x00, 0x63, 0x77, 0x6d}
)

const counterFileFooterSize = 16

type counterFileHeader struct {
	Magic     [4]byte
	Version   uint32
	MetaHash  [16]byte
	Flavor    uint8
	BigEndian bool
	_         [6]byte
}

type counterSegmentHeader struct {
	FcnEntries uint64
	StrTabLen  uint32
	ArgsLen    uint32
}

type counterFileFooter struct {
	Magic       [4]byte
	_           [4]byte
	NumSegments uint32
	_           [4]byte
}

func parseCounterFile(expectedHash [16]byte, data []byte) (map[CounterKey]uint32, error) {
	reader := bytes.NewReader(data)

	var hdr counterFileHeader
	if err := binary.Read(reader, binary.LittleEndian, &hdr); err != nil {
		return nil, fmt.Errorf("coverage: read counter header: %w", err)
	}
	if hdr.Magic != counterFileMagic {
		return nil, fmt.Errorf("coverage: invalid counter header magic")
	}
	if hdr.MetaHash != expectedHash {
		return nil, fmt.Errorf("coverage: counter/meta hash mismatch")
	}

	results := make(map[CounterKey]uint32)

	if reader.Len() < counterFileFooterSize {
		return nil, fmt.Errorf("coverage: truncated counter file")
	}

	for reader.Len() != counterFileFooterSize {
		var seg counterSegmentHeader
		if err := binary.Read(reader, binary.LittleEndian, &seg); err != nil {
			return nil, fmt.Errorf("coverage: read counter segment header: %w", err)
		}

		if seg.StrTabLen > 0 {
			if _, err := reader.Seek(int64(seg.StrTabLen), io.SeekCurrent); err != nil {
				return nil, fmt.Errorf("coverage: skip counter string table: %w", err)
			}
		}
		if seg.ArgsLen > 0 {
			if _, err := reader.Seek(int64(seg.ArgsLen), io.SeekCurrent); err != nil {
				return nil, fmt.Errorf("coverage: skip counter args: %w", err)
			}
		}

		if err := alignReader(reader, 4); err != nil {
			return nil, err
		}

		for funcEntry := uint64(0); funcEntry < seg.FcnEntries; funcEntry++ {
			numCounters, err := readUint32(reader, hdr)
			if err != nil {
				return nil, fmt.Errorf("coverage: read counter count: %w", err)
			}
			pkgIdx, err := readUint32(reader, hdr)
			if err != nil {
				return nil, fmt.Errorf("coverage: read counter pkg idx: %w", err)
			}
			funcIdx, err := readUint32(reader, hdr)
			if err != nil {
				return nil, fmt.Errorf("coverage: read counter func idx: %w", err)
			}
			for unitIdx := uint32(0); unitIdx < numCounters; unitIdx++ {
				value, err := readUint32(reader, hdr)
				if err != nil {
					return nil, fmt.Errorf("coverage: read counter value: %w", err)
				}
				results[CounterKey{Pkg: pkgIdx, Func: funcIdx, Unit: unitIdx}] = value
			}
		}
	}

	var ftr counterFileFooter
	if err := binary.Read(reader, binary.LittleEndian, &ftr); err != nil {
		return nil, fmt.Errorf("coverage: read counter footer: %w", err)
	}
	if ftr.Magic != counterFileMagic {
		return nil, fmt.Errorf("coverage: invalid counter footer magic")
	}

	return results, nil
}

func alignReader(r *bytes.Reader, alignment int64) error {
	pos, err := r.Seek(0, io.SeekCurrent)
	if err != nil {
		return fmt.Errorf("coverage: query position: %w", err)
	}
	mod := pos % alignment
	if mod == 0 {
		return nil
	}
	_, err = r.Seek(alignment-mod, io.SeekCurrent)
	if err != nil {
		return fmt.Errorf("coverage: align reader: %w", err)
	}
	return nil
}

func readUint32(r *bytes.Reader, hdr counterFileHeader) (uint32, error) {
	switch hdr.Flavor {
	case 0: // default to raw if unspecified
		return readUint32Raw(r, hdr.BigEndian)
	case 1: // CtrRaw
		return readUint32Raw(r, hdr.BigEndian)
	case 2: // CtrULeb128
		return readUint32ULEB(r)
	default:
		return 0, fmt.Errorf("coverage: unknown counter flavor %d", hdr.Flavor)
	}
}

func readUint32Raw(r *bytes.Reader, bigEndian bool) (uint32, error) {
	var buf [4]byte
	if _, err := io.ReadFull(r, buf[:]); err != nil {
		return 0, err
	}
	if bigEndian {
		return binary.BigEndian.Uint32(buf[:]), nil
	}
	return binary.LittleEndian.Uint32(buf[:]), nil
}

func readUint32ULEB(r *bytes.Reader) (uint32, error) {
	var (
		shift uint
		value uint32
	)
	for {
		var buf [1]byte
		if _, err := io.ReadFull(r, buf[:]); err != nil {
			return 0, err
		}
		b := buf[0]
		value |= uint32(b&0x7f) << shift
		if b&0x80 == 0 {
			break
		}
		shift += 7
	}
	return value, nil
}
