package tracer

import (
	"bufio"
	"compress/gzip"
	"errors"
	"fmt"
	"github.com/0xsoniclabs/carmen/go/common"
	"github.com/Fantom-foundation/lachesis-base/common/bigendian"
	"io"
	"os"
)

// NewFileHandler creates a new FileHandler that writes to a gzip-compressed file using a buffer.
func NewFileHandler(filename string) (FileHandler, error) {
	file, err := os.Create(filename)
	if err != nil {
		return nil, err
	}

	return &fileHandler{
		file:   gzip.NewWriter(file),
		buffer: bufio.NewWriter(file),
	}, nil
}

//go:generate mockgen -source file_handler.go -destination file_handler_mock.go -package tracer

type FileHandler interface {
	// WriteData writes a byte slice of any size to the file.
	WriteData(data []byte)
	// WriteUint16 writes a big-endian encoded uint16 value to the file.
	WriteUint16(data uint16)
	// WriteUint8 writes a single byte (uint8) to the file.
	WriteUint8(idx uint8)
	// todo add write code hash
	Close() error
}

// Buffer is a wrapper around necessary interfaces for writing data to a file for mocking purposes.
type Buffer interface {
	io.Writer
	io.ByteWriter
	common.Flusher
}

type fileHandler struct {
	buffer Buffer
	file   io.Closer
}

func (f *fileHandler) WriteData(data []byte) {
	_, err := f.buffer.Write(data)
	if err != nil {
		panic(fmt.Errorf("error writing []byte to buffer: %v", err))
	}
}

func (f *fileHandler) WriteUint16(data uint16) {
	_, err := f.buffer.Write(bigendian.Uint16ToBytes(data))
	if err != nil {
		panic(fmt.Errorf("error writing uint16 to buffer: %v", err))
	}
}

func (f *fileHandler) WriteUint8(idx uint8) {
	err := f.buffer.WriteByte(idx)
	if err != nil {
		panic(fmt.Errorf("error writing uint8 to buffer: %v", err))
	}
}

func (f *fileHandler) Close() error {
	// Flush the buffer to ensure all data is written to the file
	// then close the file
	return errors.Join(f.buffer.Flush(), f.file.Close())
}
