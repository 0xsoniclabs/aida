package operation

import (
	"bytes"
	"fmt"
	"github.com/Fantom-foundation/Aida/tracer/dict"
	"github.com/ethereum/go-ethereum/common"
	"io"
	"math/rand"
	"os"
	"testing"
	"time"
)

func initSetNonce(t *testing.T) (*dict.DictionaryContext, *SetNonce, common.Address, uint64) {
	rand.Seed(time.Now().UnixNano())
	addr := getRandomAddress(t)
	nonce := rand.Uint64()

	// create dictionary context
	dict := dict.NewDictionaryContext()
	cIdx := dict.EncodeContract(addr)

	// create new operation
	op := NewSetNonce(cIdx, nonce)
	if op == nil {
		t.Fatalf("failed to create operation")
	}
	// check id
	if op.GetId() != SetNonceID {
		t.Fatalf("wrong ID returned")
	}

	return dict, op, addr, nonce
}

// TestSetNonceReadWrite writes a new SetNonce object into a buffer, reads from it,
// and checks equality.
func TestSetNonceReadWrite(t *testing.T) {
	_, op1, _, _ := initSetNonce(t)
	testOperationReadWrite(t, op1, ReadSetNonce)
}

// TestSetNonceDebug creates a new SetNonce object and checks its Debug message.
func TestSetNonceDebug(t *testing.T) {
	dict, op, addr, value := initSetNonce(t)

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
	label, f := operationLabels[SetNonceID]
	if !f {
		t.Fatalf("label for %d not found", SetNonceID)
	}

	if buf.String() != fmt.Sprintf("\t%s: %s, %d\n", label, addr, value) {
		t.Fatalf("wrong debug message: %s", buf.String())
	}
}

// TestSetNonceExecute
func TestSetNonceExecute(t *testing.T) {
	dict, op, addr, nonce := initSetNonce(t)

	// check execution
	mock := NewMockStateDB()
	op.Execute(mock, dict)

	// check whether methods were correctly called
	expected := []Record{{SetNonceID, []any{addr, nonce}}}
	mock.compareRecordings(expected, t)
}
