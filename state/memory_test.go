package state

import (
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func TestInMemoryDb_SelfDestruct6780OnlyDeletesContractsCreatedInSameTransaction(t *testing.T) {
	a := common.Address{1}
	b := common.Address{2}

	db := MakeInMemoryStateDB(nil, 12)
	db.CreateContract(a)

	if want, got := false, db.HasSelfDestructed(a); want != got {
		t.Errorf("invalid self-destruct state of contract %x, want %v, got %v", a, want, got)
	}
	if want, got := false, db.HasSelfDestructed(b); want != got {
		t.Errorf("invalid self-destruct state of contract %x, want %v, got %v", b, want, got)
	}

	db.SelfDestruct6780(a) // < this should work

	if want, got := true, db.HasSelfDestructed(a); want != got {
		t.Errorf("invalid self-destruct state of contract %x, want %v, got %v", a, want, got)
	}
	if want, got := false, db.HasSelfDestructed(b); want != got {
		t.Errorf("invalid self-destruct state of contract %x, want %v, got %v", b, want, got)
	}

	db.SelfDestruct6780(b) // < this should be ignored

	if want, got := true, db.HasSelfDestructed(a); want != got {
		t.Errorf("invalid self-destruct state of contract %x, want %v, got %v", a, want, got)
	}
	if want, got := false, db.HasSelfDestructed(b); want != got {
		t.Errorf("invalid self-destruct state of contract %x, want %v, got %v", b, want, got)
	}
}

func TestInMemoryStateDB_GetLogs_ReturnEmptyLogsWithNilSnapshot(t *testing.T) {
	sdb := &inMemoryStateDB{state: nil}
	logs := sdb.GetLogs(common.Hash{}, 0, common.Hash{}, 0)
	assert.Empty(t, logs)
}

func TestInMemoryStateDB_GetLogs_AddsInfoAboutBlockAndTx(t *testing.T) {
	sdb := &inMemoryStateDB{state: &snapshot{
		parent: &snapshot{
			logs: []*types.Log{{Index: 1}},
		},
		logs: []*types.Log{{Index: 0}},
	}}
	txHash := common.Hash{0x1, 0x2, 0x3}
	blkNumber := uint64(10)
	blkHash := common.Hash{0x4, 0x5, 0x6}
	blkTimestamp := uint64(11)
	logs := sdb.GetLogs(txHash, blkNumber, blkHash, blkTimestamp)
	assert.Len(t, logs, 2) // No logs added yet
	assert.Equal(t, logs[0].TxHash, txHash)
	assert.Equal(t, blkNumber, logs[0].BlockNumber)
	assert.Equal(t, blkHash, logs[0].BlockHash)
	assert.Equal(t, blkTimestamp, logs[0].BlockTimestamp)
}
