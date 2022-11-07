package operation

import (
	"bytes"
	"fmt"
	"github.com/Fantom-foundation/Aida/tracer/dict"
	"github.com/ethereum/go-ethereum/common"
	"io"
	"os"
	"testing"
)

func initExist(t *testing.T) (*dict.DictionaryContext, *Exist, common.Address) {
	addr := getRandomAddress(t)
	// create dictionary context
	dict := dict.NewDictionaryContext()
	cIdx := dict.EncodeContract(addr)

	// create new operation
	op := NewExist(cIdx)
	if op == nil {
		t.Fatalf("failed to create operation")
	}
	// check id
	if op.GetId() != ExistID {
		t.Fatalf("wrong ID returned")
	}

	return dict, op, addr
}

// TestExistReadWrite writes a new Exist object into a buffer, reads from it,
// and checks equality.
func TestExistReadWrite(t *testing.T) {
	_, op1, _ := initExist(t)
	testOperationReadWrite(t, op1, ReadExist)
}

// TestExistDebug creates a new Exist object and checks its Debug message.
func TestExistDebug(t *testing.T) {
	dict, op, addr := initExist(t)

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
	label, f := operationLabels[ExistID]
	if !f {
		t.Fatalf("label for %d not found", ExistID)
	}

	if buf.String() != fmt.Sprintf("\t%s: %s\n", label, addr) {
		t.Fatalf("wrong debug message: %s", buf.String())
	}
}

// TestExistExecute
func TestExistExecute(t *testing.T) {
	dict, op, addr := initExist(t)

	// check execution
	mock := NewMockStateDB()
	op.Execute(mock, dict)

	// check whether methods were correctly called
	expected := []Record{{ExistID, []any{addr}}}
	mock.compareRecordings(expected, t)
}
