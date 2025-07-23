package tracer

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/0xsoniclabs/aida/txcontext"
	"github.com/Fantom-foundation/lachesis-base/common/bigendian"
	"github.com/cockroachdb/errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/holiman/uint256"
	"github.com/klauspost/compress/gzip"
	"io"
	"os"
)

func NewFileReader(filename string) (FileReader, error) {
	stat, err := os.Stat(filename)
	if err != nil {
		return nil, fmt.Errorf("could not stat file: %s, does it exist? %w", filename, err)
	}
	if stat.IsDir() {
		return nil, errors.New("given path to trace file is a directory")
	}
	if stat.Size() == 0 {
		return nil, errors.New("given trace file is empty")
	}
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("could not open trace file: %s, %w", filename, err)
	}
	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return nil, fmt.Errorf("could not create gzip reader for trace file: %s, %w", filename, err)
	}

	// Register the types we will be decoding from the trace file.
	gob.Register(txcontext.NewNilAccount())
	gob.Register(params.Rules{})
	gob.Register(types.AccessList{})
	gob.Register(types.Log{})

	return &fileReader{
		reader: bufio.NewReader(gzipReader),
		closer: gzipReader,
	}, nil
}

//go:generate mockgen -source file_reader.go -destination file_reader_mock.go -package tracer

type FileReader interface {
	// ReadAddr reads an Ethereum address (20 bytes) from the file.
	ReadAddr() (common.Address, error)
	// ReadHash reads a 32-byte hash from the file.
	ReadHash() (common.Hash, error)
	// ReadUint64 reads a big-endian encoded uint64 value from the file.
	ReadUint64() (uint64, error)
	// ReadUint32 reads a big-endian encoded uint32 value from the file.
	ReadUint32() (uint32, error)
	// ReadUint16 reads a big-endian encoded uint16 value from the file.
	ReadUint16() (uint16, error)
	// ReadUint8 reads a single byte (uint8) from the file.
	ReadUint8() (uint8, error)
	// ReadFixedSizeData reads a byte slice of given size from the file.
	ReadFixedSizeData(size int) ([]byte, error)
	// ReadVariableSizeData reads a byte slice of a variable size from the file.
	ReadVariableSizeData() ([]byte, error)
	// ReadBalanceChange reads a balance change from the file.
	ReadBalanceChange() (*uint256.Int, tracing.BalanceChangeReason, error)
	// ReadNonceChange reads a nonce change from the file.
	ReadNonceChange() (uint64, tracing.NonceChangeReason, error)
	// ReadBool reads a boolean value from the file.
	ReadBool() (bool, error)
	// ReadRules reads the rules from the file.
	ReadRules() (params.Rules, error)
	// ReadAccessList reads an access list from the file.
	ReadAccessList() (types.AccessList, error)
	// ReadWorldState reads the world state from the file.
	ReadWorldState() (txcontext.WorldState, error)
	// ReadLog reads a log entry from the file.
	ReadLog() (*types.Log, error)
	// Close closes the file reader.
	Close() error
}

// ReadBuffer is a wrapper around necessary interfaces for reading data to a file for mocking purposes.
type ReadBuffer interface {
	io.Reader
	io.ByteReader
}

type fileReader struct {
	reader ReadBuffer
	closer io.Closer
}

func (f *fileReader) ReadAddr() (common.Address, error) {
	data, err := f.ReadFixedSizeData(common.AddressLength)
	if err != nil {
		return common.Address{}, err
	}
	return common.Address(data), nil
}

func (f *fileReader) ReadHash() (common.Hash, error) {
	data, err := f.ReadFixedSizeData(common.HashLength)
	if err != nil {
		return common.Hash{}, err
	}
	return common.Hash(data), nil
}

func (f *fileReader) ReadUint64() (uint64, error) {
	data, err := f.ReadFixedSizeData(8)
	if err != nil {
		return 0, err
	}
	return bigendian.BytesToUint64(data), nil
}

func (f *fileReader) ReadUint32() (uint32, error) {
	data, err := f.ReadFixedSizeData(4)
	if err != nil {
		return 0, err
	}
	return bigendian.BytesToUint32(data), nil
}
func (f *fileReader) ReadUint16() (uint16, error) {
	data, err := f.ReadFixedSizeData(2)
	if err != nil {
		return 0, err
	}
	return bigendian.BytesToUint16(data), nil
}

func (f *fileReader) ReadUint8() (uint8, error) {
	return f.reader.ReadByte()
}

func (f *fileReader) ReadFixedSizeData(size int) ([]byte, error) {
	var (
		err  error
		data = make([]byte, size)
	)
	if _, err = io.ReadFull(f.reader, data); err != nil {
		return nil, err
	}
	return data, nil
}

func (f *fileReader) ReadVariableSizeData() ([]byte, error) {
	// We first need to read the size
	size, err := f.ReadUint32()
	if err != nil {
		return nil, fmt.Errorf("cannot read size: %w", err)
	}
	if size == 0 {
		return []byte{}, nil
	}
	data := make([]byte, size)
	if _, err = io.ReadFull(f.reader, data); err != nil {
		return nil, err
	}
	return data, nil
}

func (f *fileReader) ReadBalanceChange() (*uint256.Int, tracing.BalanceChangeReason, error) {
	data, err := f.ReadVariableSizeData()
	if err != nil {
		return nil, 0, err
	}
	reason, err := f.reader.ReadByte()
	if err != nil {
		return nil, 0, err
	}
	// Reason is the last byte in the data slice, and the rest is the uint256 value.
	return uint256.NewInt(0).SetBytes(data), tracing.BalanceChangeReason(reason), nil
}

func (f *fileReader) ReadNonceChange() (uint64, tracing.NonceChangeReason, error) {
	// 8 bytes for uint64, 1 byte for reason
	data, err := f.ReadFixedSizeData(9)
	if err != nil {
		return 0, 0, err
	}
	uintData := make([]byte, 8)
	copy(uintData, data[:len(data)-1])
	// Reason is the last byte in the data slice, and the rest is the uint256 value.
	return bigendian.BytesToUint64(uintData), tracing.NonceChangeReason(data[len(data)-1]), nil
}

func (f *fileReader) ReadBool() (bool, error) {
	b, err := f.reader.ReadByte()
	if err != nil {
		return false, err
	}
	return b != 0, nil
}

func (f *fileReader) ReadRules() (params.Rules, error) {
	data, err := f.ReadVariableSizeData()
	if err != nil {
		return params.Rules{}, err
	}
	var rules params.Rules
	err = decodeGob[*params.Rules](&rules, data)
	if err != nil {
		return params.Rules{}, err
	}
	return rules, nil
}

func (f *fileReader) ReadAccessList() (types.AccessList, error) {
	data, err := f.ReadVariableSizeData()
	if err != nil {
		return types.AccessList{}, err
	}

	var accessList types.AccessList
	// If the data is empty, we return an empty access list.
	if len(data) == 0 {
		return accessList, nil
	}
	err = decodeGob[*types.AccessList](&accessList, data)
	if err != nil {
		return nil, err
	}
	return accessList, nil
}

func (f *fileReader) ReadWorldState() (txcontext.WorldState, error) {
	data, err := f.ReadVariableSizeData()
	if err != nil {
		return nil, err
	}
	// If the data is empty, we return an empty substate.
	if len(data) == 0 {
		return txcontext.AidaWorldState{}, nil
	}
	var ws txcontext.AidaWorldState
	err = decodeGob[*txcontext.AidaWorldState](&ws, data)
	if err != nil {
		return nil, err
	}
	return ws, nil
}

func (f *fileReader) ReadLog() (*types.Log, error) {
	data, err := f.ReadVariableSizeData()
	if err != nil {
		return nil, err
	}
	log := new(types.Log)
	// If the data is empty, we return an empty log.
	if len(data) == 0 {
		return log, nil
	}
	err = decodeGob[*types.Log](log, data)
	if err != nil {
		return nil, err
	}
	return log, nil
}

func (f *fileReader) Close() error {
	return f.closer.Close()
}

// decodeGob decodes a gob-encoded byte slice into the specified type T and returns it.
func decodeGob[T any](e T, data []byte) error {
	buf := bytes.NewBuffer(data)
	// No need to hold the buffer open after decoding, so we reset it.
	defer buf.Reset()
	err := gob.NewDecoder(buf).Decode(&e)
	if err != nil {
		return fmt.Errorf("error decoding gob data: %w", err)
	}
	return nil
}
