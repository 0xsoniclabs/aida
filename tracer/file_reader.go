package tracer

type FileReader interface {
	// ReadData reads a byte slice of any size to the file.
	ReadData() ([]byte, error)
	// ReadUint16 reads a big-endian encoded uint16 value to the file.
	ReadUint16() (uint16, error)
	// ReadUint8 reads a single byte (uint8) to the file.
	ReadUint8() (uint8, error)
	// todo add write code hash
	Close() error
}
