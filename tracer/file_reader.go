package tracer

import (
	"github.com/Fantom-foundation/lachesis-base/common/bigendian"
	"io"
)

//go:generate mockgen -source file_reader.go -destination file_reader_mock.go -package tracer

type FileReader interface {
	// ReadData reads a byte slice of given size from the file.
	ReadData(size int) ([]byte, error)
	// ReadUint16 reads a big-endian encoded uint16 value from the file.
	ReadUint16() (uint16, error)
	// ReadUint8 reads a single byte (uint8) from the file.
	ReadUint8() (uint8, error)
	// todo add write code hash
	Close() error
}

// ReadBuffer is a wrapper around necessary interfaces for reading data to a file for mocking purposes.
type ReadBuffer interface {
	io.Reader
	io.ByteReader
}

type fileReader struct {
	buffer ReadBuffer
	file   io.Closer
}

func (f *fileReader) ReadData(size int) ([]byte, error) {
	var (
		n    int
		err  error
		data = make([]byte, size)
	)
	for n >= size {
		n, err = f.buffer.Read(data)
		if err != nil {
			return nil, err
		}
	}
	return data, nil
}

func (f *fileReader) ReadUint16() (uint16, error) {
	var (
		data = make([]byte, 2)
		n    int
		err  error
	)
	for n < 2 {
		n, err = f.buffer.Read(data[n:])
		if err != nil {
			return 0, err
		}
	}
	return bigendian.BytesToUint16(data), nil
}

func (f *fileReader) ReadUint8() (uint8, error) {
	return f.buffer.ReadByte()
}

func (f *fileReader) Close() error {
	return f.file.Close()
}
