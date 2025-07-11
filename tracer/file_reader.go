package tracer

import (
	"bufio"
	"fmt"
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
	return &fileReader{
		reader: bufio.NewReader(gzipReader),
		closer: gzipReader,
	}, nil
}

//go:generate mockgen -source file_reader.go -destination file_reader_mock.go -package tracer

type FileReader interface {
	ReadAddr() (common.Address, error)
	ReadHash() (common.Hash, error)
	ReadBalanceChange() (*uint256.Int, tracing.BalanceChangeReason, error)
	ReadUint64() (uint64, error)
	ReadUint32() (uint32, error)
	ReadUnknownSizeData() ([]byte, error)
	ReadNonceChange() (uint64, tracing.NonceChangeReason, error)
	ReadBool() (bool, error)
	// ReadData reads a byte slice of given size from the file.
	ReadData(size int) ([]byte, error)
	// ReadUint16 reads a big-endian encoded uint16 value from the file.
	ReadUint16() (uint16, error)
	// ReadUint8 reads a single byte (uint8) from the file.
	ReadUint8() (uint8, error)

	// todo add write code hash
	Close() error
	ReadAccessList() (types.AccessList, error)
	ReadRules() (params.Rules, error)
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

func (f *fileReader) ReadAccessList() (types.AccessList, error) {
	//TODO implement me
	panic("implement me")
}

func (f *fileReader) ReadRules() (params.Rules, error) {
	//TODO implement me
	panic("implement me")
}

func (f *fileReader) ReadBool() (bool, error) {
	b, err := f.reader.ReadByte()
	if err != nil {
		return false, err
	}
	return b != 0, nil
}

func (f *fileReader) ReadUnknownSizeData() ([]byte, error) {
	// We first need to read the size
	size, err := f.ReadUint32()
	if err != nil {
		return nil, fmt.Errorf("cannot read size: %w", err)
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
	data, err := f.ReadData(9)
	if err != nil {
		return 0, 0, err
	}
	uintData := make([]byte, 8)
	copy(uintData, data[:len(data)-1])
	// Reason is the last byte in the data slice, and the rest is the uint256 value.
	return bigendian.BytesToUint64(uintData), tracing.NonceChangeReason(data[len(data)-1]), nil
}

func (f *fileReader) ReadAddr() (common.Address, error) {
	data, err := f.ReadData(common.AddressLength)
	if err != nil {
		return common.Address{}, err
	}
	return common.Address(data), nil
}

func (f *fileReader) ReadHash() (common.Hash, error) {
	data, err := f.ReadData(common.HashLength)
	if err != nil {
		return common.Hash{}, err
	}
	return common.Hash(data), nil
}

func (f *fileReader) ReadUint64() (uint64, error) {
	data, err := f.ReadData(8)
	if err != nil {
		return 0, err
	}
	return bigendian.BytesToUint64(data), nil
}

func (f *fileReader) ReadUint32() (uint32, error) {
	data, err := f.ReadData(4)
	if err != nil {
		return 0, err
	}
	return bigendian.BytesToUint32(data), nil
}

func (f *fileReader) ReadData(size int) ([]byte, error) {
	var (
		err  error
		data = make([]byte, size)
	)
	if _, err = io.ReadFull(f.reader, data); err != nil {
		return nil, err
	}
	return data, nil
}

func (f *fileReader) ReadUint16() (uint16, error) {
	data, err := f.ReadData(2)
	if err != nil {
		return 0, err
	}
	return bigendian.BytesToUint16(data), nil
}

func (f *fileReader) ReadUint8() (uint8, error) {
	return f.reader.ReadByte()
}

func (f *fileReader) Close() error {
	return f.closer.Close()
}
