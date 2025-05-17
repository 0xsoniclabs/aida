package utils

import (
	"database/sql"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"os"
	"reflect"
	"testing"
)

func TestPrinter_NewPrinter(t *testing.T) {
	p := NewPrinters()
	assert.NotNil(t, p)
}

func TestPrinter_AddPrinter(t *testing.T) {
	p := &Printers{[]Printer{}}
	p1 := &PrinterToWriter{}
	p2 := &PrinterToWriter{}

	p.AddPrinter(p1)
	p.AddPrinter(p2)

	assert.Equal(t, 2, len(p.printers))
}

func TestPrinter_Print(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPrinter := NewMockPrinter(ctrl)
	p := &Printers{[]Printer{
		mockPrinter,
	}}
	mockPrinter.EXPECT().Print().Return(nil).Times(1)
	assert.NotPanics(t, p.Print)
}

func TestPrinter_Close(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPrinter := NewMockPrinter(ctrl)
	p := &Printers{[]Printer{
		mockPrinter,
	}}
	mockPrinter.EXPECT().Close().Return().Times(1)
	assert.NotPanics(t, p.Close)
}

func TestPrinters_AddPrinterToWriter(t *testing.T) {
	p := &Printers{}
	p.AddPrinterToWriter(os.Stdout, func() string {
		return "Hello, World!"
	})
	assert.Equal(t, 1, len(p.printers))
}

func TestPrinters_AddPrinterToConsole(t *testing.T) {
	p := &Printers{}
	p.AddPrinterToConsole(false, func() string {
		return "Hello, World!"
	})
	assert.Equal(t, 1, len(p.printers))

	p = &Printers{}
	p.AddPrinterToConsole(true, func() string {
		return "Hello, World!"
	})
	assert.Equal(t, 0, len(p.printers))
}

func TestPrinters_AddPrinterToFile(t *testing.T) {
	p := &Printers{}
	p.AddPrinterToFile("test.txt", func() string {
		return "Hello, World!"
	})
	assert.Equal(t, 1, len(p.printers))

	p = &Printers{}
	p.AddPrinterToFile("", func() string {
		return "Hello, World!"
	})
	assert.Equal(t, 0, len(p.printers))

}

func TestPrinters_AddPrinterToSqlite3(t *testing.T) {
	p := &Printers{}
	p.AddPrinterToSqlite3(":memory:", "", "", func() [][]any {
		return [][]any{}
	})
	assert.Equal(t, 1, len(p.printers))
}

func TestPrinterToWriter_Print(t *testing.T) {
	p := &PrinterToWriter{
		w: os.Stdout,
		f: func() string {
			return "Hello, World!"
		},
	}
	err := p.Print()
	assert.NoError(t, err)
}

func TestPrinterToWriter_Close(t *testing.T) {
	p := &PrinterToWriter{}
	assert.NotPanics(t, p.Close)
}

func TestPrinterToWriter_NewPrinterToWriter(t *testing.T) {
	p := NewPrinterToWriter(os.Stdout, func() string {
		return "Hello, World!"
	})
	assert.NotNil(t, p)
	assert.NotNil(t, p.w)
	assert.NotNil(t, p.f)
}

func TestPrinterToWriter_NewPrinterToConsole(t *testing.T) {
	p := NewPrinterToConsole(func() string {
		return "Hello, World!"
	})
	assert.NotNil(t, p)
	assert.Equal(t, reflect.ValueOf(os.Stdout).Pointer(), reflect.ValueOf(p.w).Pointer())
	assert.NotNil(t, p.w)
	assert.NotNil(t, p.f)
}

func TestPrinterToFile_Print(t *testing.T) {
	filePath := t.TempDir() + "/test.txt"
	p := &PrinterToFile{
		filepath: filePath,
		f: func() string {
			return "Hello, World!"
		},
	}
	err := p.Print()
	assert.NoError(t, err)
}

func TestPrinterToFile_Close(t *testing.T) {
	p := &PrinterToFile{
		filepath: t.TempDir() + "/test.txt",
		f: func() string {
			return "Hello, World!"
		},
	}
	assert.NotPanics(t, p.Close)
}

func TestPrinterToFile_NewPrinterToFile(t *testing.T) {
	filePath := t.TempDir() + "/test.txt"
	p := NewPrinterToFile(filePath, func() string {
		return "Hello, World!"
	})
	assert.NotNil(t, p)
	assert.Equal(t, filePath, p.filepath)
}

func TestPrinterToDb_Print(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db, mockDb, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)

	// case success
	p := &PrinterToDb{
		db:     db,
		insert: "",
		f: func() [][]any {
			return [][]any{}
		},
	}
	mockDb.ExpectBegin()
	mockDb.ExpectPrepare(p.insert).WillBeClosed()
	mockDb.ExpectCommit()

	err = p.Print()
	assert.NoError(t, err)

	// case Begin error
	mockErr := errors.New("mock error")
	mockDb.ExpectBegin().WillReturnError(mockErr)
	err = p.Print()
	assert.Error(t, err)

	// case Prepare error
	mockDb.ExpectBegin()
	mockDb.ExpectPrepare("").WillReturnError(mockErr)
	err = p.Print()
	assert.Error(t, err)

	// case Commit error
	mockDb.ExpectBegin()
	mockDb.ExpectPrepare("").WillBeClosed()
	mockDb.ExpectCommit().WillReturnError(mockErr)
	err = p.Print()
	assert.Error(t, err)

	if err = mockDb.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestPrinterToDb_Close(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db, mockDb, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)

	p := &PrinterToDb{
		db:     db,
		insert: "",
		f:      nil,
	}
	mockDb.ExpectClose()
	p.Close()
	if err = mockDb.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestPrinterToDb_Bufferize(t *testing.T) {
	p := &PrinterToDb{}
	buf, f := p.Bufferize(10)
	assert.NotNil(t, buf)
	assert.Equal(t, 10, buf.capacity)
	assert.NotNil(t, f)
	assert.Equal(t, 10, f.bf.capacity)
}

func TestPrinterToDb_NewPrinterToSqlite3(t *testing.T) {
	// case success
	db, err := NewPrinterToSqlite3(":memory:", "", "", func() [][]any {
		return [][]any{}
	})
	assert.NoError(t, err)
	assert.NotNil(t, db)

	// case error
	db, err = NewPrinterToSqlite3(":memory:", "asfd;asdf", "", func() [][]any {
		return [][]any{}
	})
	assert.Error(t, err)
	assert.Nil(t, db)
}

func TestPrinterToBuffer_Print(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFlusher := NewMockIFlusher(ctrl)
	p := &PrinterToBuffer{
		capacity: 10,
		f: func() [][]any {
			return [][]any{
				{"Hello", "World"},
			}
		},
		buffer:  make([][]any, 10),
		flusher: mockFlusher,
	}
	mockFlusher.EXPECT().Print().Return(nil).Times(1)
	err := p.Print()
	assert.NoError(t, err)
}
func TestPrinterToBuffer_Close(t *testing.T) {
	p := &PrinterToBuffer{
		capacity: 10,
		f:        nil,
		buffer:   make([][]any, 10),
		flusher:  &Flusher{},
	}
	assert.NotPanics(t, p.Close)
}
func TestPrinterToBuffer_Reset(t *testing.T) {
	p := &PrinterToBuffer{
		capacity: 10,
		f:        nil,
		buffer:   make([][]any, 10),
		flusher:  &Flusher{},
	}
	p.Reset()
	assert.Equal(t, 0, len(p.buffer))
}
func TestPrinterToBuffer_Length(t *testing.T) {
	p := &PrinterToBuffer{
		buffer: make([][]any, 10),
	}
	p.Length()
	assert.Equal(t, 10, p.Length())

}
func TestFlusher_Print(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db, mockDb, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)
	f := &Flusher{
		og: &PrinterToDb{
			db:     db,
			insert: "",
			f:      nil,
		},
		bf: &PrinterToBuffer{
			capacity: 0,
			f:        nil,
			buffer:   nil,
			flusher:  &Flusher{},
		},
	}
	mockDb.ExpectBegin()
	mockDb.ExpectPrepare("").WillBeClosed()
	mockDb.ExpectCommit()
	err = f.Print()
	assert.NoError(t, err)
	if err = mockDb.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestFlusher_Close(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db, mockDb, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)
	f := &Flusher{
		og: &PrinterToDb{
			db:     db,
			insert: "",
			f:      nil,
		},
		bf: &PrinterToBuffer{
			capacity: 0,
			f:        nil,
			buffer:   nil,
			flusher:  &Flusher{},
		},
	}
	assert.NotPanics(t, f.Close)
	if err := mockDb.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
