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

package utils

import (
	"errors"
	"math"
	"testing"

	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/substate/db"
	"github.com/0xsoniclabs/substate/substate"
	"github.com/Fantom-foundation/lachesis-base/common/bigendian"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/syndtr/goleveldb/leveldb"
	"go.uber.org/mock/gomock"
)

func TestDownloadPatchesJson(t *testing.T) {
	AidaDbRepositoryUrl = AidaDbRepositorySonicUrl

	patches, err := DownloadPatchesJson()
	if err != nil {
		t.Fatal(err)
	}

	if len(patches) == 0 {
		t.Fatal("patches.json are empty; are you connected to the internet?")
	}
}

func TestGetPatchFirstBlock_Positive(t *testing.T) {
	AidaDbRepositoryUrl = AidaDbRepositorySonicUrl

	patches, err := DownloadPatchesJson()
	if err != nil {
		t.Fatalf("cannot download patches.json; %v", err)
	}

	for _, p := range patches {
		firstBlock, err := getPatchFirstBlock(p.ToBlock)
		if err != nil {
			t.Fatalf("getPatchFirstBlock returned an err; %v", err)
		}

		// returned block needs to match the block in patch
		if firstBlock != p.FromBlock {
			t.Fatalf("first blocks are different; expected: %v, real: %v", firstBlock, p.FromBlock)
		}
	}
}

func TestAidaDbMetadata_SetDbHash(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockErr := errors.New("mock error")
	dbHash := []byte("hash123")

	// Case 1: Success
	mockDb := db.NewMockSubstateDB(ctrl)
	md := NewAidaDbMetadata(mockDb, "ERROR")
	mockDb.EXPECT().Put([]byte(DbHashPrefix), dbHash).Return(nil)
	err := md.SetDbHash(dbHash)
	assert.NoError(t, err)

	// Case 2: Error
	mockDb = db.NewMockSubstateDB(ctrl)
	md = NewAidaDbMetadata(mockDb, "ERROR")
	mockDb.EXPECT().Put([]byte(DbHashPrefix), dbHash).Return(mockErr)
	err = md.SetDbHash(dbHash)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), mockErr.Error())
}

func TestAidaDbMetadata_GetDbHash(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedHash := []byte("hash123")

	// Case 1: Success
	mockDb := db.NewMockSubstateDB(ctrl)
	md := NewAidaDbMetadata(mockDb, "ERROR")
	mockDb.EXPECT().Get([]byte(DbHashPrefix)).Return(expectedHash, nil)
	hash := md.GetDbHash()
	assert.Equal(t, expectedHash, hash)

	// Case 2: Not found error
	mockDb = db.NewMockSubstateDB(ctrl)
	md = NewAidaDbMetadata(mockDb, "ERROR")
	mockDb.EXPECT().Get([]byte(DbHashPrefix)).Return(nil, leveldb.ErrNotFound)
	hash = md.GetDbHash()
	assert.Nil(t, hash)

	// Case 3: Other error
	mockDb = db.NewMockSubstateDB(ctrl)
	md = NewAidaDbMetadata(mockDb, "ERROR")
	mockDb.EXPECT().Get([]byte(DbHashPrefix)).Return(nil, errors.New("other error"))
	hash = md.GetDbHash()
	assert.Nil(t, hash)
}

func TestAidaDbMetadata_DeleteMetadata(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Case 1: Success - all deletes succeed
	mockDb := db.NewMockSubstateDB(ctrl)
	md := NewAidaDbMetadata(mockDb, "ERROR")

	mockDb.EXPECT().Delete([]byte(ChainIDPrefix)).Return(nil)
	mockDb.EXPECT().Delete([]byte(FirstBlockPrefix)).Return(nil)
	mockDb.EXPECT().Delete([]byte(LastBlockPrefix)).Return(nil)
	mockDb.EXPECT().Delete([]byte(FirstEpochPrefix)).Return(nil)
	mockDb.EXPECT().Delete([]byte(LastEpochPrefix)).Return(nil)
	mockDb.EXPECT().Delete([]byte(TypePrefix)).Return(nil)
	mockDb.EXPECT().Delete([]byte(TimestampPrefix)).Return(nil)

	err := md.Delete()
	require.NoError(t, err)

	// Case 2: Some deletes fail - should log errors but not fail
	mockDb = db.NewMockSubstateDB(ctrl)
	md = NewAidaDbMetadata(mockDb, "ERROR")

	mockErr := errors.New("delete error")
	mockDb.EXPECT().Delete([]byte(ChainIDPrefix)).Return(mockErr)
	mockDb.EXPECT().Delete([]byte(FirstBlockPrefix)).Return(mockErr)
	mockDb.EXPECT().Delete([]byte(LastBlockPrefix)).Return(mockErr)
	mockDb.EXPECT().Delete([]byte(FirstEpochPrefix)).Return(mockErr)
	mockDb.EXPECT().Delete([]byte(LastEpochPrefix)).Return(mockErr)
	mockDb.EXPECT().Delete([]byte(TypePrefix)).Return(mockErr)
	mockDb.EXPECT().Delete([]byte(TimestampPrefix)).Return(mockErr)

	err = md.Delete()
	require.ErrorContains(t, err, mockErr.Error())
}

func TestAidaDbMetadata_AidaDbTyString(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	typ := GenType
	assert.Equal(t, "Generate", typ.String())

	typ = CloneType
	assert.Equal(t, "Clone", typ.String())

	typ = PatchType
	assert.Equal(t, "Patch", typ.String())

	typ = NoType
	assert.Equal(t, "NoType", typ.String())

	typ = 99
	assert.Equal(t, "unknown db type", typ.String())
}
func TestAidaDbMetadata_SetHasHashPatch(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := db.NewMockSubstateDB(ctrl)
	md := NewAidaDbMetadata(mockDb, "ERROR")

	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(nil)
	err := md.SetHasHashPatch()
	assert.NoError(t, err)
}

func TestAidaDbMetadata_SetUpdatesetInterval(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := db.NewMockSubstateDB(ctrl)
	md := NewAidaDbMetadata(mockDb, "ERROR")

	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(nil)
	err := md.SetUpdatesetInterval(uint64(99))
	assert.NoError(t, err)

	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(errors.New("mock error"))
	err = md.SetUpdatesetInterval(uint64(99))
	assert.Error(t, err)
}

func TestAidaDbMetadata_SetUpdatesetSize(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := db.NewMockSubstateDB(ctrl)
	md := NewAidaDbMetadata(mockDb, "ERROR")

	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(nil)
	err := md.SetUpdatesetSize(uint64(99))
	assert.NoError(t, err)

	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(errors.New("mock error"))
	err = md.SetUpdatesetSize(uint64(99))
	assert.Error(t, err)
}

func TestFindBlockRangeInSubstate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := db.NewMockSubstateDB(ctrl)

	// Case 1: Success
	mockDb.EXPECT().GetFirstSubstate().Return(&substate.Substate{
		Env: &substate.Env{
			Number: 0,
		},
	})
	mockDb.EXPECT().GetLastSubstate().Return(&substate.Substate{
		Env: &substate.Env{
			Number: 100,
		},
	}, nil)
	start, end, succeed := FindBlockRangeInSubstate(mockDb)
	assert.Equal(t, uint64(0), start)
	assert.Equal(t, uint64(100), end)
	assert.Equal(t, true, succeed)

	// case error
	mockDb = db.NewMockSubstateDB(ctrl)
	mockDb.EXPECT().GetFirstSubstate().Return(nil)
	start, end, succeed = FindBlockRangeInSubstate(mockDb)
	assert.Equal(t, uint64(0), start)
	assert.Equal(t, uint64(0), end)
	assert.Equal(t, false, succeed)

	// case error
	mockDb = db.NewMockSubstateDB(ctrl)
	mockDb.EXPECT().GetFirstSubstate().Return(&substate.Substate{
		Env: &substate.Env{
			Number: 0,
		},
	})
	mockDb.EXPECT().GetLastSubstate().Return(nil, errors.New("mock error"))
	start, end, succeed = FindBlockRangeInSubstate(mockDb)
	assert.Equal(t, uint64(0), start)
	assert.Equal(t, uint64(0), end)
	assert.Equal(t, false, succeed)

	// case error
	mockDb = db.NewMockSubstateDB(ctrl)
	mockDb.EXPECT().GetFirstSubstate().Return(&substate.Substate{
		Env: &substate.Env{
			Number: 0,
		},
	})
	mockDb.EXPECT().GetLastSubstate().Return(nil, nil)
	start, end, succeed = FindBlockRangeInSubstate(mockDb)
	assert.Equal(t, uint64(0), start)
	assert.Equal(t, uint64(0), end)
	assert.Equal(t, false, succeed)
}

func TestAidaDbMetadata_GetFirstBlock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := db.NewMockSubstateDB(ctrl)
	md := AidaDbMetadata{
		Db:  mockDb,
		log: logger.NewLogger("ERROR", "metadata-test"),
	}

	mockDb.EXPECT().Get([]byte(FirstBlockPrefix)).Return(nil, errors.New("mock error"))
	firstBlock := md.GetFirstBlock()
	assert.Equal(t, uint64(0), firstBlock)

	mockDb.EXPECT().Get([]byte(FirstBlockPrefix)).Return(bigendian.Uint64ToBytes(100), nil)
	firstBlock = md.GetFirstBlock()
	assert.Equal(t, uint64(100), firstBlock)

	// clear cache
	md.FirstBlock = nil
	// not found - get from substate
	mockDb.EXPECT().Get([]byte(FirstBlockPrefix)).Return(nil, leveldb.ErrNotFound)
	mockDb.EXPECT().GetFirstSubstate().Return(&substate.Substate{Block: uint64(100)})
	firstBlock = md.GetFirstBlock()
	assert.Equal(t, uint64(100), firstBlock)

	// cached - no mock call
	firstBlock = md.GetFirstBlock()
	assert.Equal(t, uint64(100), firstBlock)
}

func TestAidaDbMetadata_GetLastBlock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := db.NewMockSubstateDB(ctrl)
	md := AidaDbMetadata{
		Db:  mockDb,
		log: logger.NewLogger("ERROR", "metadata-test"),
	}

	mockDb.EXPECT().Get([]byte(LastBlockPrefix)).Return(nil, errors.New("mock error"))
	lastBlock := md.GetLastBlock()
	assert.Equal(t, uint64(0), lastBlock)

	mockDb.EXPECT().Get([]byte(LastBlockPrefix)).Return(bigendian.Uint64ToBytes(100), nil)
	lastBlock = md.GetLastBlock()
	assert.Equal(t, uint64(100), lastBlock)

	// clear cache
	md.LastBlock = nil
	// not found - get from substate
	mockDb.EXPECT().Get([]byte(LastBlockPrefix)).Return(nil, leveldb.ErrNotFound)
	mockDb.EXPECT().GetLastSubstate().Return(&substate.Substate{Block: uint64(100)}, nil)
	lastBlock = md.GetLastBlock()
	assert.Equal(t, uint64(100), lastBlock)

	// cached - no mock call
	lastBlock = md.GetLastBlock()
	assert.Equal(t, uint64(100), lastBlock)
}

func TestAidaDbMetadata_GetFirstEpoch(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := db.NewMockSubstateDB(ctrl)
	md := NewAidaDbMetadata(mockDb, "ERROR")

	mockDb.EXPECT().Get([]byte(FirstEpochPrefix)).Return(nil, errors.New("mock error"))
	data := md.GetFirstEpoch()
	assert.Equal(t, uint64(0), data)

	mockDb.EXPECT().Get([]byte(FirstEpochPrefix)).Return(nil, leveldb.ErrNotFound)
	data = md.GetFirstEpoch()
	assert.Equal(t, uint64(0), data)

	mockDb.EXPECT().Get([]byte(FirstEpochPrefix)).Return(bigendian.Uint64ToBytes(100), nil)
	data = md.GetFirstEpoch()
	assert.Equal(t, uint64(100), data)

	// cached - no mock call
	data = md.GetFirstEpoch()
	assert.Equal(t, uint64(100), data)
}

func TestAidaDbMetadata_GetLastEpoch(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := db.NewMockSubstateDB(ctrl)
	md := NewAidaDbMetadata(mockDb, "ERROR")

	mockDb.EXPECT().Get([]byte(LastEpochPrefix)).Return(nil, errors.New("mock error"))
	data := md.GetLastEpoch()
	assert.Equal(t, uint64(0), data)

	mockDb.EXPECT().Get([]byte(LastEpochPrefix)).Return(nil, leveldb.ErrNotFound)
	data = md.GetLastEpoch()
	assert.Equal(t, uint64(0), data)

	mockDb.EXPECT().Get([]byte(LastEpochPrefix)).Return(bigendian.Uint64ToBytes(100), nil)
	data = md.GetLastEpoch()
	assert.Equal(t, uint64(100), data)

	// cached - no mock call
	data = md.GetLastEpoch()
	assert.Equal(t, uint64(100), data)
}

func TestAidaDbMetadata_GetChainID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := db.NewMockSubstateDB(ctrl)
	md := NewAidaDbMetadata(mockDb, "ERROR")

	mockDb.EXPECT().Get([]byte(ChainIDPrefix)).Return(nil, errors.New("mock error"))
	data := md.GetChainID()
	assert.Equal(t, ChainID(0), data)

	mockDb.EXPECT().Get([]byte(ChainIDPrefix)).Return(nil, leveldb.ErrNotFound)
	data = md.GetChainID()
	assert.Equal(t, ChainID(0), data)

	mockDb.EXPECT().Get([]byte(ChainIDPrefix)).Return(bigendian.Uint64ToBytes(100), nil)
	data = md.GetChainID()
	assert.Equal(t, ChainID(100), data)

	// cached - no mock call
	data = md.GetChainID()
	assert.Equal(t, ChainID(100), data)
}

func TestAidaDbMetadata_GetTimestamp(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := db.NewMockSubstateDB(ctrl)
	md := NewAidaDbMetadata(mockDb, "ERROR")

	mockDb.EXPECT().Get([]byte(TimestampPrefix)).Return(nil, errors.New("mock error"))
	data := md.GetTimestamp()
	assert.Equal(t, uint64(0), data)

	mockDb.EXPECT().Get([]byte(TimestampPrefix)).Return(nil, leveldb.ErrNotFound)
	data = md.GetTimestamp()
	assert.Equal(t, uint64(0), data)

	mockDb.EXPECT().Get([]byte(TimestampPrefix)).Return(bigendian.Uint64ToBytes(100), nil)
	data = md.GetTimestamp()
	assert.Equal(t, uint64(100), data)

	// cached - no mock call
	data = md.GetTimestamp()
	assert.Equal(t, uint64(100), data)
}

func TestAidaDbMetadata_GetDbType(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := db.NewMockSubstateDB(ctrl)
	md := NewAidaDbMetadata(mockDb, "ERROR")
	mockDb.EXPECT().Get([]byte(TypePrefix)).Return(bigendian.Uint64ToBytes(100), nil)
	data := md.GetDbType()
	assert.Equal(t, AidaDbType(0), data)

	mockDb.EXPECT().Get([]byte(TypePrefix)).Return(nil, errors.New("mock error"))
	data = md.GetDbType()
	assert.Equal(t, AidaDbType(0), data)

	mockDb.EXPECT().Get([]byte(TypePrefix)).Return(nil, leveldb.ErrNotFound)
	data = md.GetDbType()
	assert.Equal(t, AidaDbType(0), data)
}

func Test_FindEpochNumber_IsSkippedForEthereumChainIDs(t *testing.T) {
	for chainID := range EthereumChainIDs {
		md := &AidaDbMetadata{ChainId: chainID}
		assert.NoError(t, md.findEpochs())
		// Epochs must be unchange
		assert.Equal(t, md.GetFirstEpoch(), uint64(0))
		assert.Equal(t, md.GetLastEpoch(), uint64(0))
	}
}

func TestMetadata_MergeOk(t *testing.T) {
	// Create two real SubstateDBs in temp dirs
	dir1 := t.TempDir()
	dir2 := t.TempDir()

	db1, err := db.NewDefaultSubstateDB(dir1)
	require.NoError(t, err)
	db2, err := db.NewDefaultSubstateDB(dir2)
	require.NoError(t, err)

	// Set up metadata for target
	md1 := NewAidaDbMetadata(db1, "ERROR")
	require.NoError(t, md1.SetChainID(SonicMainnetChainID))
	require.NoError(t, md1.SetFirstBlock(10))
	require.NoError(t, md1.SetLastBlock(20))
	require.NoError(t, md1.SetDbType(GenType))

	// Set up metadata for source
	md2 := NewAidaDbMetadata(db2, "ERROR")
	require.NoError(t, md2.SetChainID(SonicMainnetChainID))
	require.NoError(t, md2.SetFirstBlock(21))
	require.NoError(t, md2.SetLastBlock(30))
	require.NoError(t, md2.SetDbType(GenType))

	// Positive case: blocks align
	err = md1.Merge(md2)
	assert.NoError(t, err)
	assert.Equal(t, uint64(10), md1.GetFirstBlock())
	assert.Equal(t, uint64(30), md1.GetLastBlock())
	assert.Equal(t, GenType, md1.GetDbType())

	// Negative case: chain IDs differ
	md3 := NewAidaDbMetadata(db2, "ERROR")
	require.NoError(t, md3.SetChainID(200))
	require.NoError(t, md3.SetFirstBlock(21))
	require.NoError(t, md3.SetLastBlock(30))
	require.NoError(t, md3.SetDbType(GenType))

	err = md1.Merge(md3)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot merge dbs with different chainIDs")
}

func TestMetadata_MergeError(t *testing.T) {
	tests := []struct {
		name          string
		targetChainID ChainID
		targetFirst   uint64
		targetLast    uint64
		srcFirstBlock uint64
		srcLastBlock  uint64
		srcChainID    ChainID
		errMsg        string
	}{
		{
			name:          "source is subset of target",
			targetChainID: SonicMainnetChainID,
			srcChainID:    SonicMainnetChainID,
			targetFirst:   10,
			targetLast:    30,
			srcFirstBlock: 15,
			srcLastBlock:  20,
			errMsg:        "source db (15-20) is subset of target db (10-30)",
		},
		{
			name:          "target is subset of source",
			targetChainID: SonicMainnetChainID,
			srcChainID:    SonicMainnetChainID,
			targetFirst:   15,
			targetLast:    20,
			srcFirstBlock: 10,
			srcLastBlock:  30,
			errMsg:        "target db (15-20) is subset of source db (10-30)",
		},
		{
			name:          "gap before target",
			targetChainID: SonicMainnetChainID,
			srcChainID:    SonicMainnetChainID,
			targetFirst:   20,
			targetLast:    30,
			srcFirstBlock: 10,
			srcLastBlock:  18,
			errMsg:        "cannot merge dbs with gap; target db (20-30), source db (10-18)",
		},
		{
			name:          "gap after target",
			targetChainID: SonicMainnetChainID,
			srcChainID:    SonicMainnetChainID,
			targetFirst:   10,
			targetLast:    18,
			srcFirstBlock: 20,
			srcLastBlock:  30,
			errMsg:        "cannot merge dbs with gap; target db (10-18), source db (20-30)",
		},
		{
			name:          "blocks do not align (overlap)",
			targetChainID: SonicMainnetChainID,
			srcChainID:    SonicMainnetChainID,
			targetFirst:   10,
			targetLast:    20,
			srcFirstBlock: 15,
			srcLastBlock:  25,
			errMsg:        "blocks does not align; target db (10-20), source db (15-25)",
		},
		{
			name:          "different chainIDs",
			targetChainID: SonicMainnetChainID,
			srcChainID:    EthereumChainID,
			targetFirst:   10,
			targetLast:    20,
			srcFirstBlock: 21,
			srcLastBlock:  25,
			errMsg:        "cannot merge dbs with different chainIDs; target db chainID 146, source db chainID 1",
		},
		{
			name:          "unknown chainIDs",
			targetChainID: UnknownChainID,
			srcChainID:    UnknownChainID,
			targetFirst:   10,
			targetLast:    20,
			srcFirstBlock: 21,
			srcLastBlock:  25,
			errMsg:        "cannot merge dbs with no chainIDs in metadata; you can set chainID manually using the util-db insert cmd",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			db1, err := db.NewDefaultSubstateDB(t.TempDir())
			require.NoError(t, err)
			db2, err := db.NewDefaultSubstateDB(t.TempDir())
			require.NoError(t, err)

			md1 := NewAidaDbMetadata(db1, "ERROR")
			require.NoError(t, md1.SetChainID(test.targetChainID))
			require.NoError(t, md1.SetFirstBlock(test.targetFirst))
			require.NoError(t, md1.SetLastBlock(test.targetLast))
			require.NoError(t, md1.SetDbType(GenType))

			md2 := NewAidaDbMetadata(db2, "ERROR")
			require.NoError(t, md2.SetChainID(test.srcChainID))
			require.NoError(t, md2.SetFirstBlock(test.srcFirstBlock))
			require.NoError(t, md2.SetLastBlock(test.srcLastBlock))
			require.NoError(t, md2.SetDbType(GenType))

			err = md1.Merge(md2)
			assert.ErrorContains(t, err, test.errMsg)
		})
	}
}

func TestAidaDbMetadata_SetTimestamp(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Case 1: Success
	mockDb := db.NewMockSubstateDB(ctrl)
	md := NewAidaDbMetadata(mockDb, "ERROR")
	mockDb.EXPECT().Put([]byte(TimestampPrefix), gomock.Any()).Return(nil)
	err := md.SetTimestamp()
	assert.NoError(t, err)

	// Case 2: Error
	mockDb = db.NewMockSubstateDB(ctrl)
	md = NewAidaDbMetadata(mockDb, "ERROR")
	mockDb.EXPECT().Put([]byte(TimestampPrefix), gomock.Any()).Return(errors.New("mock error"))
	err = md.SetTimestamp()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mock error")
}

func TestAidaDbMetadata_GetDb(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := db.NewMockSubstateDB(ctrl)
	md := NewAidaDbMetadata(mockDb, "ERROR")
	assert.Equal(t, mockDb, md.GetDb())
}

func TestAidaDbMetadata_HasHashPatch(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := db.NewMockSubstateDB(ctrl)
	md := NewAidaDbMetadata(mockDb, "ERROR")

	// Case 1: Found
	mockDb.EXPECT().Get([]byte(HasStateHashPatchPrefix)).Return([]byte{1}, nil)
	assert.True(t, md.HasHashPatch())

	// Case 2: Not found
	mockDb.EXPECT().Get([]byte(HasStateHashPatchPrefix)).Return(nil, leveldb.ErrNotFound)
	assert.False(t, md.HasHashPatch())

	// Case 3: Error
	mockDb.EXPECT().Get([]byte(HasStateHashPatchPrefix)).Return(nil, errors.New("mock error"))
	assert.False(t, md.HasHashPatch())
}

func TestAidaDbMetadata_GetUpdatesetSize(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := db.NewMockSubstateDB(ctrl)
	md := NewAidaDbMetadata(mockDb, "ERROR")

	// Case 1: Error
	mockDb.EXPECT().Get([]byte(db.UpdatesetSizeKey)).Return(nil, errors.New("mock error"))
	size := md.GetUpdatesetSize()
	assert.Equal(t, uint64(0), size)

	// Case 2: Not found
	mockDb.EXPECT().Get([]byte(db.UpdatesetSizeKey)).Return(nil, leveldb.ErrNotFound)
	size = md.GetUpdatesetSize()
	assert.Equal(t, uint64(0), size)

	// Case 3: Success
	mockDb.EXPECT().Get([]byte(db.UpdatesetSizeKey)).Return(bigendian.Uint64ToBytes(42), nil)
	size = md.GetUpdatesetSize()
	assert.Equal(t, uint64(42), size)

	// Case 4: Cached value, no DB call
	size = md.GetUpdatesetSize()
	assert.Equal(t, uint64(42), size)
}

func TestAidaDbMetadata_GetUpdatesetInterval(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := db.NewMockSubstateDB(ctrl)
	md := NewAidaDbMetadata(mockDb, "ERROR")

	// Case 1: Error
	mockDb.EXPECT().Get([]byte(db.UpdatesetIntervalKey)).Return(nil, errors.New("mock error"))
	interval := md.GetUpdatesetInterval()
	assert.Equal(t, uint64(0), interval)

	// Case 2: Not found
	mockDb.EXPECT().Get([]byte(db.UpdatesetIntervalKey)).Return(nil, leveldb.ErrNotFound)
	interval = md.GetUpdatesetInterval()
	assert.Equal(t, uint64(0), interval)

	// Case 3: Success
	mockDb.EXPECT().Get([]byte(db.UpdatesetIntervalKey)).Return(bigendian.Uint64ToBytes(99), nil)
	interval = md.GetUpdatesetInterval()
	assert.Equal(t, uint64(99), interval)

	// Case 4: Cached value, no DB call
	interval = md.GetUpdatesetInterval()
	assert.Equal(t, uint64(99), interval)
}

func TestFindEpochNumber_UnknownBlock(t *testing.T) {
	epoch, err := FindEpochNumber(math.MaxInt64, SonicMainnetChainID)
	require.NoError(t, err)
	require.Equal(t, uint64(0), epoch)
}
