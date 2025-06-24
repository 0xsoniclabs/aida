package utils

import (
	"testing"

	"github.com/0xsoniclabs/substate/db"
	"github.com/0xsoniclabs/substate/rlp"
	"github.com/0xsoniclabs/substate/substate"
	substatetypes "github.com/0xsoniclabs/substate/types"
	trlp "github.com/0xsoniclabs/substate/types/rlp"
	"github.com/stretchr/testify/assert"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/testutil"
	"go.uber.org/mock/gomock"
)

func TestWorldStateUpdate_GenerateUpdateSet(t *testing.T) {
	// TODO Protobuf encoding not supported yet
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	baseDb := db.NewMockBaseDB(ctrl)
	mockDb := db.NewMockDbAdapter(ctrl)

	input := getTestSubstate("default")
	input.Block = 0
	input.Transaction = 1
	encoded, err := trlp.EncodeToBytes(rlp.NewRLP(input))
	if err != nil {
		t.Fatalf("Failed to encode substate: %v", err)
	}

	expectedDestroyed := []substatetypes.Address{{1}, {2}}
	expectedResurrected := []substatetypes.Address{{3}}
	list := db.SuicidedAccountLists{DestroyedAccounts: expectedDestroyed, ResurrectedAccounts: expectedResurrected}
	value, _ := trlp.EncodeToBytes(list)

	kv := &testutil.KeyValue{}
	kv.PutU(db.SubstateDBKey(input.Block, input.Transaction), encoded)
	iter := iterator.NewArrayIterator(kv)
	mockDb.EXPECT().Get(gomock.Any(), gomock.Any()).Return(encoded, nil).AnyTimes()
	mockDb.EXPECT().NewIterator(gomock.Any(), gomock.Any()).Return(iter)
	baseDb.EXPECT().GetBackend().Return(mockDb)
	baseDb.EXPECT().Get(gomock.Any()).Return(value, nil).AnyTimes()

	set, i, err := GenerateUpdateSet(0, 2, &Config{
		Workers:          1,
		SubstateEncoding: "rlp",
	}, baseDb)
	assert.NoError(t, err)
	assert.NotNil(t, set)
	assert.Equal(t, 1, len(set))
	assert.Equal(t, 3, len(i))
}

func TestWorldStateUpdate_GenerateWorldStateFromUpdateDB(t *testing.T) {
	src := t.TempDir() + "/test.db"
	db, err := createTestUpdateDB(src)
	if err != nil {
		t.Fatalf("Failed to create test substate db: %v", err)
	}
	err = db.PutUpdateSet(testUpdateSet, testDeletedAccounts)
	if err != nil {
		t.Fatalf("Failed to put substate to test substate db: %v", err)
	}
	err = db.Close()
	if err != nil {
		t.Fatalf("Failed to close test substate db: %v", err)
	}
	dst := t.TempDir() + "/test2.db"
	db2, err := createTestUpdateDB(dst)
	if err != nil {
		t.Fatalf("Failed to create test substate db: %v", err)
	}
	err = db2.Close()
	if err != nil {
		t.Fatalf("Failed to close test substate db: %v", err)
	}

	// create mock db
	// case error
	cfg := &Config{
		AidaDb:     src,
		DeletionDb: dst,
		Workers:    1,
	}
	ws, err := GenerateWorldStateFromUpdateDB(cfg, uint64(100))
	assert.NoError(t, err)
	assert.NotNil(t, ws)
}

func TestWorldStateUpdate_ClearAccountStorage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	addr := substatetypes.BytesToAddress([]byte("test"))
	ws := substate.WorldState{}
	ws[addr] = &substate.Account{
		Nonce:   1,
		Balance: nil,
		Storage: map[substatetypes.Hash]substatetypes.Hash{
			substatetypes.BytesToHash([]byte("key1")): substatetypes.BytesToHash([]byte("value1")),
		},
		Code: nil,
	}
	ClearAccountStorage(ws, []substatetypes.Address{addr})
	assert.Equal(t, 0, len(ws[addr].Storage))
}
