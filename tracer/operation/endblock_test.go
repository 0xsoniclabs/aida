package operation

import (
	"bytes"
	"fmt"
	"github.com/Fantom-foundation/Aida/tracer/dict"
	"io"
	"os"
	"testing"
)

func initEndBlock(t *testing.T) (*dict.DictionaryContext, *EndBlock) {
	// create dictionary context
	dict := dict.NewDictionaryContext()

	// create new operation
	op := NewEndBlock()
	if op == nil {
		t.Fatalf("failed to create operation")
	}
	// check id
	if op.GetId() != EndBlockID {
		t.Fatalf("wrong ID returned")
	}

	return dict, op
}

// TestEndBlockReadWrite writes a new EndBlock object into a buffer, reads from it,
// and checks equality.
func TestEndBlockReadWrite(t *testing.T) {
	_, op1 := initEndBlock(t)
	testOperationReadWrite(t, op1, ReadEndBlock)
}

// TestEndBlockDebug creates a new EndBlock object and checks its Debug message.
func TestEndBlockDebug(t *testing.T) {
	dict, op := initEndBlock(t)

	// divert stdout to a buffer
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// print debug message
	op.Debug(dict)

	// restore stdout
	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	io.Copy(&buf, r)

	// check debug message
	label, f := operationLabels[EndBlockID]
	if !f {
		t.Fatalf("label for %d not found", EndBlockID)
	}

	if buf.String() != fmt.Sprintf("\t%s\n", label) {
		t.Fatalf("wrong debug message: %s", buf.String())
	}
}

// TestEndBlockExecute
func TestEndBlockExecute(t *testing.T) {
	dict, op := initEndBlock(t)

	// check execution
	mock := NewMockStateDB()
	op.Execute(mock, dict)

	// check whether methods were correctly called
	mock.compareRecordings([]Record{}, t)
	// currently EndBlock isn't recorded
	//expected := []Record{{EndBlockID, []any{}}}
	//mock.compareRecordings(expected, t)
}
