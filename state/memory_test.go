// Copyright 2025 Sonic Labs
// This file is part of Aida Testing Infrastructure for Sonic
//
// Aida is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Aida is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with Aida. If not, see <http://www.gnu.org/licenses/>.

package state

import (
	"testing"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
	gomock "go.uber.org/mock/gomock"

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

func TestInMemoryStateDB_GetLogs_AddsLogsWithCorrectTimestamp(t *testing.T) {
	txHash := common.Hash{0x1, 0x2, 0x3}
	blkNumber := uint64(10)
	blkHash := common.Hash{0x4, 0x5, 0x6}
	blkTimestamp := uint64(11)
	sdb := &inMemoryStateDB{state: &snapshot{
		parent: &snapshot{
			logs: []*types.Log{{Index: 1, BlockTimestamp: blkTimestamp}},
		},
		logs: []*types.Log{{Index: 0}},
	}}
	logs := sdb.GetLogs(txHash, blkNumber, blkHash, blkTimestamp)
	assert.Len(t, logs, 1) // No logs added yet
	assert.Equal(t, blkTimestamp, logs[0].BlockTimestamp)
	assert.Equal(t, uint(1), logs[0].Index)
}

func TestInMemoryStateDB_GetStateAndCommittedState_Returns(t *testing.T) {
	ctrl := gomock.NewController(t)
	db := NewMockVmStateDB(ctrl)

	address := common.Address{1}
	key := common.Hash{2}
	state := common.Hash{3}
	committed := common.Hash{4}

	db.EXPECT().GetStateAndCommittedState(address, key).Return(state, committed)
	gotState, gotCommitted := db.GetStateAndCommittedState(address, key)
	assert.Equal(t, state, gotState)
	assert.Equal(t, committed, gotCommitted)
}
