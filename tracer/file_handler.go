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

func NewFileHandler(filename string) (*FileHandler, error) {
	file, err := os.Create(filename)
	if err != nil {
		return nil, err
	}

	return &FileHandler{
		file:   gzip.NewWriter(file),
		buffer: bufio.NewWriter(file),
	}, nil
}

type FileHandler struct {
	buffer *bufio.Writer
	file   io.Closer
}

func (f *FileHandler) Close() error {
	// Flush the buffer to ensure all data is written to the file
	// then close the file
	return errors.Join(f.buffer.Flush(), f.file.Close())
}

func (f *FileHandler) WriteData(data []byte) {
	_, err := f.buffer.Write(data)
	if err != nil {
		panic(fmt.Errorf("error writing []byte to buffer: %v", err))
	}
}

func (f *FileHandler) WriteUint16(data uint16) {
	_, err := f.buffer.Write(bigendian.Uint16ToBytes(data))
	if err != nil {
		panic(fmt.Errorf("error writing uint16 to buffer: %v", err))
	}
}

func (f *FileHandler) WriteUint8(idx uint8) {
	err := f.buffer.WriteByte(idx)
	if err != nil {
		panic(fmt.Errorf("error writing uint8 to buffer: %v", err))
	}
}
