package rpc

import (
	"compress/gzip"
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func mockGzipFile(t *testing.T, path string) {
	file, err := os.Create(path)
	if err != nil {
		t.Fatal("failed to create gzip file:", err)
	}
	defer func(file *os.File) {
		err = file.Close()
		if err != nil {
			t.Fatal("failed to close file:", err)
		}
	}(file)

	gzipWriter := gzip.NewWriter(file)
	defer func(gzipWriter *gzip.Writer) {
		err = gzipWriter.Close()
		if err != nil {
			t.Fatal("failed to close gzip writer:", err)
		}
	}(gzipWriter)

	testData := `{"test": "data", "message": "hello world"}`
	_, err = gzipWriter.Write([]byte(testData))
	if err != nil {
		t.Fatal("failed to write test data to gzip file:", err)
	}
}

func TestRpc_NewFileReader(t *testing.T) {

	t.Run("success", func(t *testing.T) {
		tempDir := t.TempDir()
		path := tempDir + "/test.json.gz"

		mockGzipFile(t, path)

		reader, err := NewFileReader(context.TODO(), path)
		if err != nil {
			t.Fatalf("failed to create file reader: %v", err)
		}
		defer reader.Close()

		assert.NotNil(t, reader)
	})

	t.Run("not exist", func(t *testing.T) {
		tempDir := t.TempDir()
		path := tempDir + "/nonexistent.json.gz"

		reader, err := NewFileReader(context.TODO(), path)
		assert.Error(t, err)
		assert.Nil(t, reader)
	})

	t.Run("invalid gzip", func(t *testing.T) {
		tempDir := t.TempDir()
		path := tempDir + "/nonexistent.json.gz"
		file, err := os.Create(path)
		if err != nil {
			t.Fatal("failed to create file:", err)
		}
		err = file.Close()
		if err != nil {
			t.Fatal("failed to close file:", err)
		}

		reader, err := NewFileReader(context.TODO(), path)
		assert.Error(t, err)
		assert.Nil(t, reader)
	})
}
