// Copyright 2024 Fantom Foundation
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
	"testing"

	"github.com/0xsoniclabs/substate/db"
	"github.com/0xsoniclabs/substate/substate"
	"github.com/Fantom-foundation/lachesis-base/common/bigendian"
	"github.com/stretchr/testify/assert"
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

func TestAidaDbMetadata_SetAll(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockErr := errors.New("mock error")

	// Case 1: No errors
	mockDb := db.NewMockBaseDB(ctrl)
	md := NewAidaDbMetadata(mockDb, "ERROR")
	md.FirstBlock = 100
	md.LastBlock = 200
	md.FirstEpoch = 10
	md.LastEpoch = 20
	md.ChainId = MainnetChainID
	md.DbType = GenType

	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(nil).Times(7) // 7 metadata fields
	err := md.SetAll()
	assert.NoError(t, err)

	// Case 2: Error with SetFirstBlock
	mockDb = db.NewMockBaseDB(ctrl)
	md = NewAidaDbMetadata(mockDb, "ERROR")
	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(mockErr)
	err = md.SetAll()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), mockErr.Error())

	// Case 3: Error with SetLastBlock
	mockDb = db.NewMockBaseDB(ctrl)
	md = NewAidaDbMetadata(mockDb, "ERROR")
	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(nil).Times(1)
	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(mockErr)
	err = md.SetAll()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), mockErr.Error())

	// Case 4: Error with SetFirstEpoch
	mockDb = db.NewMockBaseDB(ctrl)
	md = NewAidaDbMetadata(mockDb, "ERROR")
	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(nil).Times(2)
	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(mockErr)
	err = md.SetAll()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), mockErr.Error())

	// Case 5: Error with SetLastEpoch
	mockDb = db.NewMockBaseDB(ctrl)
	md = NewAidaDbMetadata(mockDb, "ERROR")
	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(nil).Times(3)
	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(mockErr)
	err = md.SetAll()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), mockErr.Error())

	// Case 6: Error with SetChainId
	mockDb = db.NewMockBaseDB(ctrl)
	md = NewAidaDbMetadata(mockDb, "ERROR")
	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(nil).Times(4)
	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(mockErr)
	err = md.SetAll()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), mockErr.Error())

	// Case 7: Error with SetDbType
	mockDb = db.NewMockBaseDB(ctrl)
	md = NewAidaDbMetadata(mockDb, "ERROR")
	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(nil).Times(5)
	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(mockErr)
	err = md.SetAll()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), mockErr.Error())

	// Case 8: Error with SetTimestamp
	mockDb = db.NewMockBaseDB(ctrl)
	md = NewAidaDbMetadata(mockDb, "ERROR")
	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(nil).Times(6)
	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(mockErr)
	err = md.SetAll()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), mockErr.Error())
}

func TestAidaDbMetadata_SetDbHash(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockErr := errors.New("mock error")
	dbHash := []byte("hash123")

	// Case 1: Success
	mockDb := db.NewMockBaseDB(ctrl)
	md := NewAidaDbMetadata(mockDb, "ERROR")
	mockDb.EXPECT().Put([]byte(DbHashPrefix), dbHash).Return(nil)
	err := md.SetDbHash(dbHash)
	assert.NoError(t, err)

	// Case 2: Error
	mockDb = db.NewMockBaseDB(ctrl)
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
	mockDb := db.NewMockBaseDB(ctrl)
	md := NewAidaDbMetadata(mockDb, "ERROR")
	mockDb.EXPECT().Get([]byte(DbHashPrefix)).Return(expectedHash, nil)
	hash := md.GetDbHash()
	assert.Equal(t, expectedHash, hash)

	// Case 2: Not found error
	mockDb = db.NewMockBaseDB(ctrl)
	md = NewAidaDbMetadata(mockDb, "ERROR")
	mockDb.EXPECT().Get([]byte(DbHashPrefix)).Return(nil, leveldb.ErrNotFound)
	hash = md.GetDbHash()
	assert.Nil(t, hash)

	// Case 3: Other error
	mockDb = db.NewMockBaseDB(ctrl)
	md = NewAidaDbMetadata(mockDb, "ERROR")
	mockDb.EXPECT().Get([]byte(DbHashPrefix)).Return(nil, errors.New("other error"))
	hash = md.GetDbHash()
	assert.Nil(t, hash)
}

func TestAidaDbMetadata_SetAllMetadata(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockErr := errors.New("mock error")
	dbHash := []byte("hash123")

	// Case 1: Success
	mockDb := db.NewMockBaseDB(ctrl)
	md := NewAidaDbMetadata(mockDb, "ERROR")
	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(nil).Times(8) // 8 metadata fields including hash
	err := md.SetAllMetadata(100, 200, 10, 20, MainnetChainID, dbHash, GenType)
	assert.NoError(t, err)

	// Case 2: Error with SetFirstBlock
	mockDb = db.NewMockBaseDB(ctrl)
	md = NewAidaDbMetadata(mockDb, "ERROR")
	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(mockErr)
	err = md.SetAllMetadata(100, 200, 10, 20, MainnetChainID, dbHash, GenType)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), mockErr.Error())

	// Case 3: Error with SetLastBlock
	mockDb = db.NewMockBaseDB(ctrl)
	md = NewAidaDbMetadata(mockDb, "ERROR")
	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(nil).Times(1)
	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(mockErr)
	err = md.SetAllMetadata(100, 200, 10, 20, MainnetChainID, dbHash, GenType)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), mockErr.Error())

	// Case 4: Error with SetFirstEpoch
	mockDb = db.NewMockBaseDB(ctrl)
	md = NewAidaDbMetadata(mockDb, "ERROR")
	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(nil).Times(2)
	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(mockErr)
	err = md.SetAllMetadata(100, 200, 10, 20, MainnetChainID, dbHash, GenType)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), mockErr.Error())

	// Case 5: Error with SetLastEpoch
	mockDb = db.NewMockBaseDB(ctrl)
	md = NewAidaDbMetadata(mockDb, "ERROR")
	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(nil).Times(3)
	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(mockErr)
	err = md.SetAllMetadata(100, 200, 10, 20, MainnetChainID, dbHash, GenType)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), mockErr.Error())

	// Case 6: Error with SetChainId
	mockDb = db.NewMockBaseDB(ctrl)
	md = NewAidaDbMetadata(mockDb, "ERROR")
	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(nil).Times(4)
	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(mockErr)
	err = md.SetAllMetadata(100, 200, 10, 20, MainnetChainID, dbHash, GenType)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), mockErr.Error())

	// Case 7: Error with SetDbType
	mockDb = db.NewMockBaseDB(ctrl)
	md = NewAidaDbMetadata(mockDb, "ERROR")
	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(nil).Times(5)
	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(mockErr)
	err = md.SetAllMetadata(100, 200, 10, 20, MainnetChainID, dbHash, GenType)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), mockErr.Error())

	// Case 8: Error with SetTimestamp
	mockDb = db.NewMockBaseDB(ctrl)
	md = NewAidaDbMetadata(mockDb, "ERROR")
	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(nil).Times(6)
	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(mockErr)
	err = md.SetAllMetadata(100, 200, 10, 20, MainnetChainID, dbHash, GenType)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), mockErr.Error())

	// Case 9: Error with SetDbHash
	mockDb = db.NewMockBaseDB(ctrl)
	md = NewAidaDbMetadata(mockDb, "ERROR")
	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(nil).Times(7)
	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(mockErr)
	err = md.SetAllMetadata(100, 200, 10, 20, MainnetChainID, dbHash, GenType)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), mockErr.Error())
}

func TestAidaDbMetadata_CheckUpdateMetadata(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Case 1: Success
	cfg := &Config{
		LogLevel: "ERROR",
		ChainID:  MainnetChainID,
	}
	mockAidaDb := db.NewMockBaseDB(ctrl)
	mockPatchDb := db.NewMockBaseDB(ctrl)
	md := NewAidaDbMetadata(mockAidaDb, "ERROR")
	mockAidaDb.EXPECT().Get(gomock.Any()).Return(bigendian.Uint64ToBytes(4564026), nil).AnyTimes()
	mockPatchDb.EXPECT().Get(gomock.Any()).Return(bigendian.Uint64ToBytes(4564026), nil).AnyTimes()
	dbHash, err := md.CheckUpdateMetadata(cfg, mockPatchDb)
	assert.NoError(t, err)
	assert.Equal(t, []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x45, 0xa4, 0x3a}, dbHash)

	// Case 2: Last block is Zero
	cfg = &Config{
		LogLevel: "ERROR",
		ChainID:  ChainID(0),
	}
	mockAidaDb = db.NewMockBaseDB(ctrl)
	mockPatchDb = db.NewMockBaseDB(ctrl)
	md = NewAidaDbMetadata(mockAidaDb, "ERROR")
	mockAidaDb.EXPECT().Get(gomock.Any()).Return(bigendian.Uint64ToBytes(0), nil).AnyTimes()
	mockPatchDb.EXPECT().Get(gomock.Any()).Return(bigendian.Uint64ToBytes(0), nil).AnyTimes()
	dbHash, err = md.CheckUpdateMetadata(cfg, mockPatchDb)
	assert.Error(t, err)
	assert.Nil(t, dbHash)

	// Case 3: Block not aligned
	cfg = &Config{
		LogLevel: "ERROR",
		ChainID:  MainnetChainID,
	}
	mockAidaDb = db.NewMockBaseDB(ctrl)
	mockPatchDb = db.NewMockBaseDB(ctrl)
	md = NewAidaDbMetadata(mockAidaDb, "ERROR")
	mockAidaDb.EXPECT().Get(gomock.Any()).Return(bigendian.Uint64ToBytes(741852), nil).AnyTimes()
	mockPatchDb.EXPECT().Get(gomock.Any()).Return(bigendian.Uint64ToBytes(123456), nil).AnyTimes()
	dbHash, err = md.CheckUpdateMetadata(cfg, mockPatchDb)
	assert.Error(t, err)
	assert.Nil(t, dbHash)

	// Case 4: ChainID mismatch
	cfg = &Config{
		LogLevel: "ERROR",
		ChainID:  MainnetChainID,
	}
	mockAidaDb = db.NewMockBaseDB(ctrl)
	mockPatchDb = db.NewMockBaseDB(ctrl)
	md = NewAidaDbMetadata(mockAidaDb, "ERROR")
	mockAidaDb.EXPECT().Get(gomock.Any()).Return(bigendian.Uint64ToBytes(4564026), nil).AnyTimes()
	mockPatchDb.EXPECT().Get([]byte(ChainIDPrefix)).Return(bigendian.Uint64ToBytes(123456), nil).AnyTimes()
	mockPatchDb.EXPECT().Get(gomock.Any()).Return(bigendian.Uint64ToBytes(4564026), nil).AnyTimes()
	dbHash, err = md.CheckUpdateMetadata(cfg, mockPatchDb)
	assert.Error(t, err)
	assert.Nil(t, dbHash)

	// Case 5
	cfg = &Config{
		LogLevel: "ERROR",
		ChainID:  MainnetChainID,
	}
	mockAidaDb = db.NewMockBaseDB(ctrl)
	mockPatchDb = db.NewMockBaseDB(ctrl)
	md = NewAidaDbMetadata(mockAidaDb, "ERROR")
	mockAidaDb.EXPECT().Get([]byte(ChainIDPrefix)).Return(bigendian.Uint64ToBytes(4564026), nil).AnyTimes()
	mockAidaDb.EXPECT().Get(gomock.Any()).Return(bigendian.Uint64ToBytes(4564025), nil).AnyTimes()
	mockPatchDb.EXPECT().Get([]byte(ChainIDPrefix)).Return(bigendian.Uint64ToBytes(4564026), nil).AnyTimes()
	mockPatchDb.EXPECT().Get(gomock.Any()).Return(bigendian.Uint64ToBytes(4564026), nil).AnyTimes()
	dbHash, err = md.CheckUpdateMetadata(cfg, mockPatchDb)
	assert.NoError(t, err)
	assert.Equal(t, []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x45, 0xa4, 0x3a}, dbHash)

	// Case 6
	cfg = &Config{
		LogLevel: "ERROR",
		ChainID:  MainnetChainID,
	}
	mockAidaDb = db.NewMockBaseDB(ctrl)
	mockPatchDb = db.NewMockBaseDB(ctrl)
	md = NewAidaDbMetadata(mockAidaDb, "ERROR")
	mockAidaDb.EXPECT().Get(gomock.Any()).Return(bigendian.Uint64ToBytes(4564026), nil).AnyTimes()
	mockPatchDb.EXPECT().Get([]byte(FirstBlockPrefix)).Return(bigendian.Uint64ToBytes(0), nil).AnyTimes()
	mockPatchDb.EXPECT().Get(gomock.Any()).Return(bigendian.Uint64ToBytes(4564026), nil).AnyTimes()
	dbHash, err = md.CheckUpdateMetadata(cfg, mockPatchDb)
	assert.NoError(t, err)
	assert.Equal(t, []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x45, 0xa4, 0x3a}, dbHash)
}

func TestAidaDbMetadata_SetFreshMetadata(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// case 1
	mockDb := db.NewMockBaseDB(ctrl)
	md := NewAidaDbMetadata(mockDb, "ERROR")
	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(nil).Times(3)
	err := md.SetFreshMetadata(MainnetChainID)
	assert.NoError(t, err)

	// case 2
	err = md.SetFreshMetadata(0)
	assert.Error(t, err)

	// case 3
	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(errors.New("error"))
	err = md.SetFreshMetadata(MainnetChainID)
	assert.Error(t, err)

	// case 4
	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(nil)
	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(errors.New("error"))
	err = md.SetFreshMetadata(MainnetChainID)
	assert.Error(t, err)

	// case 5
	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(nil).Times(2)
	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(errors.New("error"))
	err = md.SetFreshMetadata(MainnetChainID)
	assert.Error(t, err)
}

func TestAidaDbMetadata_SetBlockRange(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockErr := errors.New("mock error")

	// Case 1: Success
	mockDb := db.NewMockBaseDB(ctrl)
	md := NewAidaDbMetadata(mockDb, "ERROR")

	mockDb.EXPECT().Put([]byte(FirstBlockPrefix), gomock.Any()).Return(nil)
	mockDb.EXPECT().Put([]byte(LastBlockPrefix), gomock.Any()).Return(nil)

	err := md.SetBlockRange(100, 200)
	assert.NoError(t, err)

	// Case 2: Error with SetFirstBlock
	mockDb = db.NewMockBaseDB(ctrl)
	md = NewAidaDbMetadata(mockDb, "ERROR")

	mockDb.EXPECT().Put([]byte(FirstBlockPrefix), gomock.Any()).Return(mockErr)

	err = md.SetBlockRange(100, 200)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), mockErr.Error())

	// Case 3: Error with SetLastBlock
	mockDb = db.NewMockBaseDB(ctrl)
	md = NewAidaDbMetadata(mockDb, "ERROR")

	mockDb.EXPECT().Put([]byte(FirstBlockPrefix), gomock.Any()).Return(nil)
	mockDb.EXPECT().Put([]byte(LastBlockPrefix), gomock.Any()).Return(mockErr)

	err = md.SetBlockRange(100, 200)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), mockErr.Error())
}

func TestAidaDbMetadata_DeleteMetadata(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Case 1: Success - all deletes succeed
	mockDb := db.NewMockBaseDB(ctrl)
	md := NewAidaDbMetadata(mockDb, "ERROR")

	mockDb.EXPECT().Delete([]byte(ChainIDPrefix)).Return(nil)
	mockDb.EXPECT().Delete([]byte(FirstBlockPrefix)).Return(nil)
	mockDb.EXPECT().Delete([]byte(LastBlockPrefix)).Return(nil)
	mockDb.EXPECT().Delete([]byte(FirstEpochPrefix)).Return(nil)
	mockDb.EXPECT().Delete([]byte(LastEpochPrefix)).Return(nil)
	mockDb.EXPECT().Delete([]byte(TypePrefix)).Return(nil)
	mockDb.EXPECT().Delete([]byte(TimestampPrefix)).Return(nil)

	md.DeleteMetadata() // Should not panic or return error

	// Case 2: Some deletes fail - should log errors but not fail
	mockDb = db.NewMockBaseDB(ctrl)
	md = NewAidaDbMetadata(mockDb, "ERROR")

	mockErr := errors.New("delete error")
	mockDb.EXPECT().Delete([]byte(ChainIDPrefix)).Return(mockErr)
	mockDb.EXPECT().Delete([]byte(FirstBlockPrefix)).Return(mockErr)
	mockDb.EXPECT().Delete([]byte(LastBlockPrefix)).Return(mockErr)
	mockDb.EXPECT().Delete([]byte(FirstEpochPrefix)).Return(mockErr)
	mockDb.EXPECT().Delete([]byte(LastEpochPrefix)).Return(mockErr)
	mockDb.EXPECT().Delete([]byte(TypePrefix)).Return(mockErr)
	mockDb.EXPECT().Delete([]byte(TimestampPrefix)).Return(mockErr)

	md.DeleteMetadata() // Should not panic despite errors
}

func TestAidaDbMetadata_UpdateMetadataInOldAidaDb(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockErr := errors.New("mock error")

	// Case 1: No existing metadata, all values should be set
	mockDb := db.NewMockBaseDB(ctrl)
	md := NewAidaDbMetadata(mockDb, "ERROR")

	// Set expectations for checking current values (all return not found)
	mockDb.EXPECT().Get([]byte(ChainIDPrefix)).Return(nil, errors.New("not found"))
	mockDb.EXPECT().Get([]byte(FirstBlockPrefix)).Return(nil, errors.New("not found"))
	mockDb.EXPECT().Get([]byte(LastBlockPrefix)).Return(nil, errors.New("not found"))
	mockDb.EXPECT().Get([]byte(TypePrefix)).Return(nil, errors.New("not found"))

	// Set expectations for setting new values
	mockDb.EXPECT().Put([]byte(ChainIDPrefix), gomock.Any()).Return(nil)
	mockDb.EXPECT().Put([]byte(FirstBlockPrefix), gomock.Any()).Return(nil)
	mockDb.EXPECT().Put([]byte(LastBlockPrefix), gomock.Any()).Return(nil)
	mockDb.EXPECT().Put([]byte(TypePrefix), gomock.Any()).Return(nil)

	err := md.UpdateMetadataInOldAidaDb(MainnetChainID, 100, 200)
	assert.NoError(t, err)

	// Case 2: Error when setting ChainID
	mockDb = db.NewMockBaseDB(ctrl)
	md = NewAidaDbMetadata(mockDb, "ERROR")

	mockDb.EXPECT().Get([]byte(ChainIDPrefix)).Return(nil, errors.New("not found"))
	mockDb.EXPECT().Put([]byte(ChainIDPrefix), gomock.Any()).Return(mockErr)

	err = md.UpdateMetadataInOldAidaDb(MainnetChainID, 100, 200)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), mockErr.Error())

	// Case 3: Error when setting FirstBlock
	mockDb = db.NewMockBaseDB(ctrl)
	md = NewAidaDbMetadata(mockDb, "ERROR")

	mockDb.EXPECT().Get([]byte(ChainIDPrefix)).Return(nil, errors.New("not found"))
	mockDb.EXPECT().Put([]byte(ChainIDPrefix), gomock.Any()).Return(nil)
	mockDb.EXPECT().Get([]byte(FirstBlockPrefix)).Return(nil, errors.New("not found"))
	mockDb.EXPECT().Put([]byte(FirstBlockPrefix), gomock.Any()).Return(mockErr)

	err = md.UpdateMetadataInOldAidaDb(MainnetChainID, 100, 200)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), mockErr.Error())

	// Case 4: Error when setting LastBlock
	mockDb = db.NewMockBaseDB(ctrl)
	md = NewAidaDbMetadata(mockDb, "ERROR")

	mockDb.EXPECT().Get([]byte(ChainIDPrefix)).Return(nil, errors.New("not found"))
	mockDb.EXPECT().Put([]byte(ChainIDPrefix), gomock.Any()).Return(nil)
	mockDb.EXPECT().Get([]byte(FirstBlockPrefix)).Return(nil, errors.New("not found"))
	mockDb.EXPECT().Put([]byte(FirstBlockPrefix), gomock.Any()).Return(nil)
	mockDb.EXPECT().Get([]byte(LastBlockPrefix)).Return(nil, errors.New("not found"))
	mockDb.EXPECT().Put([]byte(LastBlockPrefix), gomock.Any()).Return(mockErr)

	err = md.UpdateMetadataInOldAidaDb(MainnetChainID, 100, 200)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), mockErr.Error())

	// Case 5: Error when setting DbType
	mockDb = db.NewMockBaseDB(ctrl)
	md = NewAidaDbMetadata(mockDb, "ERROR")

	mockDb.EXPECT().Get([]byte(ChainIDPrefix)).Return(nil, errors.New("not found"))
	mockDb.EXPECT().Put([]byte(ChainIDPrefix), gomock.Any()).Return(nil)
	mockDb.EXPECT().Get([]byte(FirstBlockPrefix)).Return(nil, errors.New("not found"))
	mockDb.EXPECT().Put([]byte(FirstBlockPrefix), gomock.Any()).Return(nil)
	mockDb.EXPECT().Get([]byte(LastBlockPrefix)).Return(nil, errors.New("not found"))
	mockDb.EXPECT().Put([]byte(LastBlockPrefix), gomock.Any()).Return(nil)
	mockDb.EXPECT().Get([]byte(TypePrefix)).Return(nil, errors.New("not found"))
	mockDb.EXPECT().Put([]byte(TypePrefix), gomock.Any()).Return(mockErr)

	err = md.UpdateMetadataInOldAidaDb(MainnetChainID, 100, 200)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), mockErr.Error())

	// Case 6: Some metadata already exists
	mockDb = db.NewMockBaseDB(ctrl)
	md = NewAidaDbMetadata(mockDb, "ERROR")

	// Chain ID already exists
	mockDb.EXPECT().Get([]byte(ChainIDPrefix)).Return(bigendian.Uint64ToBytes(uint64(MainnetChainID)), nil)
	// First block needs to be set
	mockDb.EXPECT().Get([]byte(FirstBlockPrefix)).Return(nil, errors.New("not found"))
	mockDb.EXPECT().Put([]byte(FirstBlockPrefix), gomock.Any()).Return(nil)
	// Last block already exists
	mockDb.EXPECT().Get([]byte(LastBlockPrefix)).Return(bigendian.Uint64ToBytes(999), nil)
	// Type still needs to be set
	mockDb.EXPECT().Get([]byte(TypePrefix)).Return(nil, errors.New("not found"))
	mockDb.EXPECT().Put([]byte(TypePrefix), gomock.Any()).Return(nil)

	err = md.UpdateMetadataInOldAidaDb(MainnetChainID, 100, 200)
	assert.NoError(t, err)
}

func TestNewAidaDbMetadata(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := db.NewMockBaseDB(ctrl)
	metadata := NewAidaDbMetadata(mockDb, "ERROR")
	assert.Equal(t, mockDb, metadata.Db)
}

func TestProcessPatchLikeMetadata(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockErr := errors.New("mock error")

	// no error
	mockDb := db.NewMockBaseDB(ctrl)
	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	err := ProcessPatchLikeMetadata(mockDb, "ERROR", 0, 0, 0, 0, 0, true, nil)
	assert.NoError(t, err)

	// no error
	mockDb = db.NewMockBaseDB(ctrl)
	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	err = ProcessPatchLikeMetadata(mockDb, "ERROR", 0, 0, 0, 0, 0, false, nil)
	assert.NoError(t, err)

	// SetFirstBlock error
	mockDb = db.NewMockBaseDB(ctrl)
	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(mockErr)
	err = ProcessPatchLikeMetadata(mockDb, "ERROR", 0, 0, 0, 0, 0, true, nil)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), mockErr.Error())

	// SetLastBlock error
	mockDb = db.NewMockBaseDB(ctrl)
	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(nil).Times(1)
	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(mockErr)
	err = ProcessPatchLikeMetadata(mockDb, "ERROR", 0, 0, 0, 0, 0, true, nil)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), mockErr.Error())

	// SetFirstEpoch error
	mockDb = db.NewMockBaseDB(ctrl)
	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(nil).Times(2)
	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(mockErr)
	err = ProcessPatchLikeMetadata(mockDb, "ERROR", 0, 0, 0, 0, 0, true, nil)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), mockErr.Error())

	// SetLastEpoch error
	mockDb = db.NewMockBaseDB(ctrl)
	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(nil).Times(3)
	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(mockErr)
	err = ProcessPatchLikeMetadata(mockDb, "ERROR", 0, 0, 0, 0, 0, true, nil)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), mockErr.Error())

	// SetChainID error
	mockDb = db.NewMockBaseDB(ctrl)
	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(nil).Times(4)
	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(mockErr)
	err = ProcessPatchLikeMetadata(mockDb, "ERROR", 0, 0, 0, 0, 0, true, nil)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), mockErr.Error())

	// SetDbType error
	mockDb = db.NewMockBaseDB(ctrl)
	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(nil).Times(5)
	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(mockErr)
	err = ProcessPatchLikeMetadata(mockDb, "ERROR", 0, 0, 0, 0, 0, true, nil)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), mockErr.Error())

	// SetTimestamp error
	mockDb = db.NewMockBaseDB(ctrl)
	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(nil).Times(6)
	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(mockErr)
	err = ProcessPatchLikeMetadata(mockDb, "ERROR", 0, 0, 0, 0, 0, true, nil)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), mockErr.Error())

	// SetDbHash error
	mockDb = db.NewMockBaseDB(ctrl)
	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(nil).Times(7)
	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(mockErr)
	err = ProcessPatchLikeMetadata(mockDb, "ERROR", 0, 0, 0, 0, 0, true, nil)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), mockErr.Error())
}

func TestProcessCloneLikeMetadata(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockErr := errors.New("mock error")

	// no error
	mockDb := db.NewMockBaseDB(ctrl)
	mockDb.EXPECT().Get(gomock.Any()).Return(bigendian.Uint64ToBytes(99), nil).AnyTimes()
	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	err := ProcessCloneLikeMetadata(mockDb, NoType, "ERROR", 0, 0, MainnetChainID)
	assert.NoError(t, err)

	// SetFirstBlock error
	mockDb = db.NewMockBaseDB(ctrl)
	mockDb.EXPECT().Get(gomock.Any()).Return(bigendian.Uint64ToBytes(99), nil).AnyTimes()
	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(mockErr)
	err = ProcessCloneLikeMetadata(mockDb, NoType, "ERROR", 0, 0, MainnetChainID)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), mockErr.Error())

	// SetLastBlock error
	mockDb = db.NewMockBaseDB(ctrl)
	mockDb.EXPECT().Get(gomock.Any()).Return(bigendian.Uint64ToBytes(99), nil).AnyTimes()
	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(nil).Times(1)
	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(mockErr)
	err = ProcessCloneLikeMetadata(mockDb, NoType, "ERROR", 0, 0, MainnetChainID)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), mockErr.Error())

	// SetFirstEpoch error
	mockDb = db.NewMockBaseDB(ctrl)
	mockDb.EXPECT().Get(gomock.Any()).Return(bigendian.Uint64ToBytes(99), nil).AnyTimes()
	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(nil).Times(2)
	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(mockErr)
	err = ProcessCloneLikeMetadata(mockDb, NoType, "ERROR", 0, 0, MainnetChainID)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), mockErr.Error())

	// SetLastEpoch error
	mockDb = db.NewMockBaseDB(ctrl)
	mockDb.EXPECT().Get(gomock.Any()).Return(bigendian.Uint64ToBytes(99), nil).AnyTimes()
	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(nil).Times(3)
	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(mockErr)
	err = ProcessCloneLikeMetadata(mockDb, NoType, "ERROR", 0, 0, MainnetChainID)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), mockErr.Error())

	// SetChainID error
	mockDb = db.NewMockBaseDB(ctrl)
	mockDb.EXPECT().Get(gomock.Any()).Return(bigendian.Uint64ToBytes(99), nil).AnyTimes()
	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(nil).Times(4)
	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(mockErr)
	err = ProcessCloneLikeMetadata(mockDb, NoType, "ERROR", 0, 0, MainnetChainID)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), mockErr.Error())

	// SetDbType error
	mockDb = db.NewMockBaseDB(ctrl)
	mockDb.EXPECT().Get(gomock.Any()).Return(bigendian.Uint64ToBytes(99), nil).AnyTimes()
	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(nil).Times(5)
	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(mockErr)
	err = ProcessCloneLikeMetadata(mockDb, NoType, "ERROR", 0, 0, MainnetChainID)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), mockErr.Error())

	// SetTimestamp error
	mockDb = db.NewMockBaseDB(ctrl)
	mockDb.EXPECT().Get(gomock.Any()).Return(bigendian.Uint64ToBytes(99), nil).AnyTimes()
	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(nil).Times(6)
	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(mockErr)
	err = ProcessCloneLikeMetadata(mockDb, NoType, "ERROR", 0, 0, MainnetChainID)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), mockErr.Error())
}

func TestProcessGenLikeMetadata(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := db.NewMockBaseDB(ctrl)
	mockDb.EXPECT().Get(gomock.Any()).Return(bigendian.Uint64ToBytes(99), nil).AnyTimes()
	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	err := ProcessGenLikeMetadata(mockDb, 0, 0, 0, 0, MainnetChainID, "ERROR", nil)
	assert.NoError(t, err)
}

func TestAidaDbMetadata_genMetadata(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockErr := errors.New("mock error")

	// no error
	mockDb := db.NewMockBaseDB(ctrl)
	md := NewAidaDbMetadata(mockDb, "ERROR")
	mockDb.EXPECT().Get(gomock.Any()).Return(bigendian.Uint64ToBytes(99), nil).AnyTimes()
	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	err := md.genMetadata(0, 0, 0, 0, MainnetChainID, nil)
	assert.NoError(t, err)

	// SetFirstBlock error
	mockDb = db.NewMockBaseDB(ctrl)
	md = NewAidaDbMetadata(mockDb, "ERROR")
	mockDb.EXPECT().Get(gomock.Any()).Return(bigendian.Uint64ToBytes(99), nil).AnyTimes()
	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(mockErr)
	err = md.genMetadata(0, 0, 0, 0, MainnetChainID, nil)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), mockErr.Error())

	// SetLastBlock error
	mockDb = db.NewMockBaseDB(ctrl)
	md = NewAidaDbMetadata(mockDb, "ERROR")
	mockDb.EXPECT().Get(gomock.Any()).Return(bigendian.Uint64ToBytes(99), nil).AnyTimes()
	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(nil).Times(1)
	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(mockErr)
	err = md.genMetadata(0, 0, 0, 0, MainnetChainID, nil)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), mockErr.Error())

	// SetFirstEpoch error
	mockDb = db.NewMockBaseDB(ctrl)
	md = NewAidaDbMetadata(mockDb, "ERROR")
	mockDb.EXPECT().Get(gomock.Any()).Return(bigendian.Uint64ToBytes(99), nil).AnyTimes()
	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(nil).Times(2)
	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(mockErr)
	err = md.genMetadata(0, 0, 0, 0, MainnetChainID, nil)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), mockErr.Error())

	// SetLastEpoch error
	mockDb = db.NewMockBaseDB(ctrl)
	md = NewAidaDbMetadata(mockDb, "ERROR")
	mockDb.EXPECT().Get(gomock.Any()).Return(bigendian.Uint64ToBytes(99), nil).AnyTimes()
	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(nil).Times(3)
	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(mockErr)
	err = md.genMetadata(0, 0, 0, 0, MainnetChainID, nil)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), mockErr.Error())

	// SetChainID error
	mockDb = db.NewMockBaseDB(ctrl)
	md = NewAidaDbMetadata(mockDb, "ERROR")
	mockDb.EXPECT().Get(gomock.Any()).Return(bigendian.Uint64ToBytes(99), nil).AnyTimes()
	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(nil).Times(4)
	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(mockErr)
	err = md.genMetadata(0, 0, 0, 0, MainnetChainID, nil)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), mockErr.Error())

	// SetDbType error
	mockDb = db.NewMockBaseDB(ctrl)
	md = NewAidaDbMetadata(mockDb, "ERROR")
	mockDb.EXPECT().Get(gomock.Any()).Return(bigendian.Uint64ToBytes(99), nil).AnyTimes()
	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(nil).Times(5)
	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(mockErr)
	err = md.genMetadata(0, 0, 0, 0, MainnetChainID, nil)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), mockErr.Error())

	// SetTimestamp error
	mockDb = db.NewMockBaseDB(ctrl)
	md = NewAidaDbMetadata(mockDb, "ERROR")
	mockDb.EXPECT().Get(gomock.Any()).Return(bigendian.Uint64ToBytes(99), nil).AnyTimes()
	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(nil).Times(6)
	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(mockErr)
	err = md.genMetadata(0, 0, 0, 0, MainnetChainID, nil)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), mockErr.Error())

	// SetDbHash error
	mockDb = db.NewMockBaseDB(ctrl)
	md = NewAidaDbMetadata(mockDb, "ERROR")
	mockDb.EXPECT().Get(gomock.Any()).Return(bigendian.Uint64ToBytes(99), nil).AnyTimes()
	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(nil).Times(7)
	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(mockErr)
	err = md.genMetadata(0, 0, 0, 0, MainnetChainID, nil)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), mockErr.Error())

}

func TestAidaDbMetadata_getVerboseDbType(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := db.NewMockBaseDB(ctrl)
	md := NewAidaDbMetadata(mockDb, "ERROR")
	md.DbType = GenType
	dbType := md.getVerboseDbType()
	assert.Equal(t, "Generate", dbType)

	md.DbType = CloneType
	dbType = md.getVerboseDbType()
	assert.Equal(t, "Clone", dbType)

	md.DbType = PatchType
	dbType = md.getVerboseDbType()
	assert.Equal(t, "Patch", dbType)

	md.DbType = NoType
	dbType = md.getVerboseDbType()
	assert.Equal(t, "NoType", dbType)

	md.DbType = 99
	dbType = md.getVerboseDbType()
	assert.Equal(t, "unknown db type", dbType)
}

func TestAidaDbMetadata_getBlockRange(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := db.NewMockBaseDB(ctrl)
	md := NewAidaDbMetadata(mockDb, "ERROR")

	mockDb.EXPECT().Get(gomock.Any()).Return(bigendian.Uint64ToBytes(99), nil).Times(2)
	err := md.getBlockRange()
	assert.NoError(t, err)
	assert.Equal(t, uint64(99), md.FirstBlock)
	assert.Equal(t, uint64(99), md.LastBlock)

	mockDb.EXPECT().Get(gomock.Any()).Return(bigendian.Uint64ToBytes(0), nil).Times(2)
	err = md.getBlockRange()
	assert.Error(t, err)
}

func TestAidaDbMetadata_SetHasHashPatch(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := db.NewMockBaseDB(ctrl)
	md := NewAidaDbMetadata(mockDb, "ERROR")

	mockDb.EXPECT().Put(gomock.Any(), gomock.Any()).Return(nil)
	err := md.SetHasHashPatch()
	assert.NoError(t, err)
}

func TestAidaDbMetadata_SetUpdatesetInterval(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := db.NewMockBaseDB(ctrl)
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

	mockDb := db.NewMockBaseDB(ctrl)
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

	mockDb := db.NewMockBaseDB(ctrl)
	md := NewAidaDbMetadata(mockDb, "ERROR")
	mockDb.EXPECT().Get([]byte(FirstBlockPrefix)).Return(bigendian.Uint64ToBytes(100), nil)
	firstBlock := md.GetFirstBlock()
	assert.Equal(t, uint64(100), firstBlock)

	mockDb.EXPECT().Get([]byte(FirstBlockPrefix)).Return(nil, errors.New("mock error"))
	firstBlock = md.GetFirstBlock()
	assert.Equal(t, uint64(0), firstBlock)

	mockDb.EXPECT().Get([]byte(FirstBlockPrefix)).Return(nil, leveldb.ErrNotFound)
	firstBlock = md.GetFirstBlock()
	assert.Equal(t, uint64(0), firstBlock)
}

func TestAidaDbMetadata_GetLastBlock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := db.NewMockBaseDB(ctrl)
	md := NewAidaDbMetadata(mockDb, "ERROR")
	mockDb.EXPECT().Get([]byte(LastBlockPrefix)).Return(bigendian.Uint64ToBytes(100), nil)
	lastBlock := md.GetLastBlock()
	assert.Equal(t, uint64(100), lastBlock)

	mockDb.EXPECT().Get([]byte(LastBlockPrefix)).Return(nil, errors.New("mock error"))
	lastBlock = md.GetLastBlock()
	assert.Equal(t, uint64(0), lastBlock)

	mockDb.EXPECT().Get([]byte(LastBlockPrefix)).Return(nil, leveldb.ErrNotFound)
	lastBlock = md.GetLastBlock()
	assert.Equal(t, uint64(0), lastBlock)
}

func TestAidaDbMetadata_GetFirstEpoch(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := db.NewMockBaseDB(ctrl)
	md := NewAidaDbMetadata(mockDb, "ERROR")
	mockDb.EXPECT().Get([]byte(FirstEpochPrefix)).Return(bigendian.Uint64ToBytes(100), nil)
	data := md.GetFirstEpoch()
	assert.Equal(t, uint64(100), data)

	mockDb.EXPECT().Get([]byte(FirstEpochPrefix)).Return(nil, errors.New("mock error"))
	data = md.GetFirstEpoch()
	assert.Equal(t, uint64(0), data)

	mockDb.EXPECT().Get([]byte(FirstEpochPrefix)).Return(nil, leveldb.ErrNotFound)
	data = md.GetFirstEpoch()
	assert.Equal(t, uint64(0), data)
}

func TestAidaDbMetadata_GetLastEpoch(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := db.NewMockBaseDB(ctrl)
	md := NewAidaDbMetadata(mockDb, "ERROR")
	mockDb.EXPECT().Get([]byte(LastEpochPrefix)).Return(bigendian.Uint64ToBytes(100), nil)
	data := md.GetLastEpoch()
	assert.Equal(t, uint64(100), data)

	mockDb.EXPECT().Get([]byte(LastEpochPrefix)).Return(nil, errors.New("mock error"))
	data = md.GetLastEpoch()
	assert.Equal(t, uint64(0), data)

	mockDb.EXPECT().Get([]byte(LastEpochPrefix)).Return(nil, leveldb.ErrNotFound)
	data = md.GetLastEpoch()
	assert.Equal(t, uint64(0), data)
}

func TestAidaDbMetadata_GetChainID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := db.NewMockBaseDB(ctrl)
	md := NewAidaDbMetadata(mockDb, "ERROR")
	mockDb.EXPECT().Get([]byte(ChainIDPrefix)).Return(bigendian.Uint64ToBytes(100), nil)
	data := md.GetChainID()
	assert.Equal(t, ChainID(100), data)

	mockDb = db.NewMockBaseDB(ctrl)
	md = NewAidaDbMetadata(mockDb, "ERROR")
	mockDb.EXPECT().Get([]byte(ChainIDPrefix)).Return([]byte{0, 1}, nil)
	data = md.GetChainID()
	assert.Equal(t, ChainID(1), data)

	mockDb.EXPECT().Get([]byte(ChainIDPrefix)).Return(nil, errors.New("mock error"))
	data = md.GetChainID()
	assert.Equal(t, ChainID(0), data)

	mockDb.EXPECT().Get([]byte(ChainIDPrefix)).Return(nil, leveldb.ErrNotFound)
	data = md.GetChainID()
	assert.Equal(t, ChainID(0), data)
}

func TestAidaDbMetadata_GetTimestamp(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := db.NewMockBaseDB(ctrl)
	md := NewAidaDbMetadata(mockDb, "ERROR")
	mockDb.EXPECT().Get([]byte(TimestampPrefix)).Return(bigendian.Uint64ToBytes(100), nil)
	data := md.GetTimestamp()
	assert.Equal(t, uint64(100), data)

	mockDb.EXPECT().Get([]byte(TimestampPrefix)).Return(nil, errors.New("mock error"))
	data = md.GetTimestamp()
	assert.Equal(t, uint64(0), data)

	mockDb.EXPECT().Get([]byte(TimestampPrefix)).Return(nil, leveldb.ErrNotFound)
	data = md.GetTimestamp()
	assert.Equal(t, uint64(0), data)
}

func TestAidaDbMetadata_GetDbType(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := db.NewMockBaseDB(ctrl)
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
