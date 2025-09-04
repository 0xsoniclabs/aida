package operation

import (
	"fmt"
	"testing"

	"github.com/0xsoniclabs/aida/tracer/context"
	"github.com/ethereum/go-ethereum/common"
)

func initGetStateAndCommittedStateLcls(t *testing.T) (*context.Replay, *GetStateAndCommittedStateLcls, common.Address, common.Hash) {
	// create context context
	ctx := context.NewReplay()

	// create new operation
	op := NewGetStateAndCommittedStateLcls()
	if op == nil {
		t.Fatalf("failed to create operation")
	}
	// check id
	if op.GetId() != GetStateAndCommittedStateLclsID {
		t.Fatalf("wrong ID returned")
	}

	addr := getRandomAddress(t)
	ctx.EncodeContract(addr)

	storage := getRandomHash(t)
	ctx.EncodeKey(storage)

	return ctx, op, addr, storage
}

// TestGetStateLclsReadWrite writes a new GetStateLcls object into a buffer, reads from it,
// and checks equality.
func TestGetStateAndCommittedStateLclsReadWrite(t *testing.T) {
	_, op1, _, _ := initGetStateAndCommittedStateLcls(t)
	testOperationReadWrite(t, op1, ReadGetStateAndCommittedStateLcls)
}

// TestGetStateAndCommittedStateLclsDebug creates a new GetStateLcls object and checks its Debug message.
func TestGetStateAndCommittedStateLclsDebug(t *testing.T) {
	ctx, op, addr, storage := initGetStateAndCommittedStateLcls(t)
	testOperationDebug(t, ctx, op, fmt.Sprint(addr, storage))
}

// TestGetStateAndCommittedStateLclsExecute
func TestGetStateAndCommittedStateLclsExecute(t *testing.T) {
	ctx, op, addr, storage := initGetStateAndCommittedStateLcls(t)

	// check execution
	mock := NewMockStateDB()
	op.Execute(mock, ctx)

	// check whether methods were correctly called
	expected := []Record{{GetStateID, []any{addr, storage}}, {GetCommittedStateID, []any{addr, storage}}}
	mock.compareRecordings(expected, t)
}
