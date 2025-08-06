package tracer

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/0xsoniclabs/aida/utils/bigendian"
	"github.com/0xsoniclabs/carmen/go/common"
	"github.com/klauspost/compress/gzip"
)

// NewFileWriter creates a new FileWriter that writes to a gzip-compressed file using a buffer.
func NewFileWriter(filename string) (FileWriter, error) {
	_, err := os.Stat(filename)
	if err == nil {
		return nil, fmt.Errorf("file %s already exists", filename)
	}

	file, err := os.Create(filename)
	if err != nil {
		return nil, err
	}

	gzipWriter := gzip.NewWriter(file)
	return &fileWriter{
		buffer: bufio.NewWriter(gzipWriter),
		closer: gzipWriter,
	}, nil
}

//go:generate mockgen -source file_writer.go -destination file_writer_mock.go -package tracer

type FileWriter interface {
	// WriteData writes a byte slice of any size to the file.
	WriteData(data []byte) error
	// WriteUint16 writes a big-endian encoded uint16 value to the file.
	WriteUint16(data uint16) error
	// WriteUint8 writes a single byte (uint8) to the file.
	WriteUint8(idx uint8) error
	// todo add write code hash
	Close() error
}

// WriteBuffer is a wrapper around necessary interfaces for writing data to a file for mocking purposes.
type WriteBuffer interface {
	io.Writer
	io.ByteWriter
	common.Flusher
}

type fileWriter struct {
	buffer WriteBuffer
	closer io.Closer
}

func (f *fileWriter) WriteData(data []byte) error {
	_, err := f.buffer.Write(data)
	if err != nil {
		return fmt.Errorf("error writing []byte to buffer: %w", err)
	}
	return nil
}

func (f *fileWriter) WriteUint16(data uint16) error {
	_, err := f.buffer.Write(bigendian.Uint16ToBytes(data))
	if err != nil {
		return fmt.Errorf("error writing uint16 to buffer: %w", err)
	}
	return nil
}

func (f *fileWriter) WriteUint8(idx uint8) error {
	err := f.buffer.WriteByte(idx)
	if err != nil {
		return fmt.Errorf("error writing uint8 to buffer: %w", err)
	}
	return nil
}

func (f *fileWriter) Close() error {
	// Flush the buffer to ensure all data is written to the file
	// then close the file
	return errors.Join(f.buffer.Flush(), f.closer.Close())
}
