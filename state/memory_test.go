package state

import (
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
