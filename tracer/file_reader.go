package tracer

import (
	"bufio"
	"fmt"
	"github.com/Fantom-foundation/lachesis-base/common/bigendian"
	"github.com/cockroachdb/errors"
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
	reader ReadBuffer
	closer io.Closer
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
	var (
		data = make([]byte, 2)
		err  error
	)
	if _, err = io.ReadFull(f.reader, data); err != nil {
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
