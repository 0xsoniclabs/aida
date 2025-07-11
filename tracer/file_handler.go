package tracer

import (
	"bufio"
	"compress/gzip"
	"errors"
	"fmt"
	"github.com/Fantom-foundation/lachesis-base/common/bigendian"
	"io"
	"os"
)

//go:generate mockgen -source file_handler.go -destination file_handler_mock.go -package tracer

type FileHandler interface {
	WriteData(data []byte)
	WriteUint16(data uint16)
	WriteUint8(idx uint8)
	Close() error
}

func NewFileHandler(filename string) (*fileHandler, error) {
	file, err := os.Create(filename)
	if err != nil {
		return nil, err
	}

	return &fileHandler{
		file:   gzip.NewWriter(file),
		buffer: bufio.NewWriter(file),
	}, nil
}

type fileHandler struct {
	buffer *bufio.Writer
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
