package utils

import (
	"fmt"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/stretchr/testify/assert"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/testutil"
	"go.uber.org/mock/gomock"
	"testing"
)

//go:generate mockgen -source db_search_test.go -destination db_search_mock.go -package utils
type ethDatabase interface {
	ethdb.Database
}

func TestSearchableDB_NewSearchableDB(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockethDatabase(ctrl)
	db := NewSearchableDB(mockDb)
	assert.NotNil(t, db)
	assert.Equal(t, mockDb, db.Database)
}

func Test_GetLastKey(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	kv := &testutil.KeyValue{}
	kv.PutString("key", "value")

	// case 1
	mockDB := NewMockethDatabase(ctrl)
	mockDB.EXPECT().NewIterator(gomock.Any(), gomock.Any()).Return(iterator.NewArrayIterator(kv))
	mockDB.EXPECT().NewIterator(gomock.Any(), gomock.Any()).
		DoAndReturn(func(prefix, start []byte) ethdb.Iterator {
			if len(start) > 0 && start[0] <= 49 {
				return iterator.NewArrayIterator(kv)
			}
			return iterator.NewEmptyIterator(nil)
		}).AnyTimes()
	expected, err := StateHashKeyToUint64([]byte("11111111"))
	if err != nil {
		t.Fatalf("Failed to convert key to uint64: %v", err)
	}
	key, err := GetLastKey(mockDB, StateHashPrefix)
	assert.Nil(t, err)
	assert.Equal(t, expected, key)

	// case 2
	mockDB = NewMockethDatabase(ctrl)
	mockDB.EXPECT().NewIterator(gomock.Any(), gomock.Any()).Return(iterator.NewEmptyIterator(nil)).Times(8)
	key, err = GetLastKey(mockDB, StateHashPrefix)
	assert.Error(t, err)
	assert.Equal(t, uint64(0), key)

	// case 3
	mockDB = NewMockethDatabase(ctrl)
	mockDB.EXPECT().NewIterator(gomock.Any(), gomock.Any()).Return(iterator.NewEmptyIterator(nil)).Times(7)
	mockDB.EXPECT().NewIterator(gomock.Any(), gomock.Any()).Return(iterator.NewArrayIterator(kv))
	mockDB.EXPECT().NewIterator(gomock.Any(), gomock.Any()).Return(iterator.NewArrayIterator(kv)).AnyTimes()
	key, err = GetLastKey(mockDB, StateHashPrefix)
	assert.Error(t, err)
	assert.Equal(t, uint64(0), key)

	// case 4
	mockDB = NewMockethDatabase(ctrl)
	mockDB.EXPECT().NewIterator(gomock.Any(), gomock.Any()).Return(iterator.NewArrayIterator(kv))
	mockDB.EXPECT().NewIterator(gomock.Any(), gomock.Any()).
		DoAndReturn(func(prefix, start []byte) ethdb.Iterator {
			return iterator.NewArrayIterator(kv)
		}).AnyTimes()
	key, err = GetLastKey(mockDB, StateHashPrefix)
	assert.Error(t, err)
	assert.Equal(t, uint64(0), key)

}

func TestSearchableDB_binarySearchForLastPrefixKey_AllValues(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Test representative values across the full range
	testValues := []byte{0, 1, 49, 10, 127, 128, 200, 254, 255}

	for _, maxTrueValue := range testValues {
		t.Run(fmt.Sprintf("Returns byte %d", maxTrueValue), func(t *testing.T) {
			mockBackend := NewMockethDatabase(ctrl)
			db := &SearchableDB{mockBackend}

			// Configure mock to control which values return true based on our test case
			mockBackend.EXPECT().NewIterator(gomock.Any(), gomock.Any()).
				DoAndReturn(func(prefix, start []byte) ethdb.Iterator {
					if len(start) > 0 && start[0] <= maxTrueValue {
						kv := &testutil.KeyValue{}
						kv.PutU([]byte{1}, []byte("value"))
						return iterator.NewArrayIterator(kv)
					}
					return iterator.NewEmptyIterator(nil)
				}).AnyTimes()

			result, err := db.binarySearchForLastPrefixKey([]byte{1})
			assert.Nil(t, err)
			assert.Equal(t, maxTrueValue, result)
		})
	}

	// Test error case when no values return true
	t.Run("Returns error when no values match", func(t *testing.T) {
		mockBackend := NewMockethDatabase(ctrl)
		db := &SearchableDB{mockBackend}

		mockBackend.EXPECT().NewIterator(gomock.Any(), gomock.Any()).
			Return(iterator.NewEmptyIterator(nil)).AnyTimes()

		_, err := db.binarySearchForLastPrefixKey([]byte{1})
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "no value found in search")
	})
}

func TestSearchableDB_getLongestEncodedKeyZeroPrefixLength(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := NewMockethDatabase(ctrl)
	db := &SearchableDB{
		mockDb,
	}

	// Case: Found at index 2
	kv := &testutil.KeyValue{}
	kv.PutU([]byte{1}, []byte("value"))
	mockDb.EXPECT().NewIterator(gomock.Any(), gomock.Any()).Return(iterator.NewEmptyIterator(nil)).Times(2)
	mockDb.EXPECT().NewIterator(gomock.Any(), gomock.Any()).Return(iterator.NewArrayIterator(kv))

	result, err := db.getLongestEncodedKeyZeroPrefixLength(StateHashPrefix)

	assert.Nil(t, err)
	assert.Equal(t, byte(2), result)

	// Case: Not found
	mockDb.EXPECT().NewIterator(gomock.Any(), gomock.Any()).Return(iterator.NewEmptyIterator(nil)).Times(8)

	result, err = db.getLongestEncodedKeyZeroPrefixLength(StateHashPrefix)

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "unable to find prefix")
	assert.Equal(t, byte(0), result)
}

func TestSearchableDB_hasKeyValuesFor(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBackend := NewMockethDatabase(ctrl)
	db := &SearchableDB{mockBackend}

	// Case 1: Success - found at max value
	kv := &testutil.KeyValue{}
	kv.PutU([]byte{1}, []byte("value"))
	mockBackend.EXPECT().NewIterator(gomock.Any(), gomock.Any()).Return(iterator.NewArrayIterator(kv))

	result := db.hasKeyValuesFor([]byte{1}, []byte{1})

	assert.True(t, result)
}

func TestSearchableDB_binarySearchForLastPrefixKey(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBackend := NewMockethDatabase(ctrl)
	db := &SearchableDB{mockBackend}

	// Case 1: found at max value
	kv := &testutil.KeyValue{}
	kv.PutU([]byte{1}, []byte("value"))
	mockBackend.EXPECT().NewIterator(gomock.Any(), gomock.Any()).Return(iterator.NewArrayIterator(kv)).Times(8)
	mockBackend.EXPECT().NewIterator(gomock.Any(), gomock.Any()).Return(iterator.NewArrayIterator(kv))
	mockBackend.EXPECT().NewIterator(gomock.Any(), gomock.Any()).Return(iterator.NewArrayIterator(kv))

	result, err := db.binarySearchForLastPrefixKey([]byte{1})
	assert.Nil(t, err)
	assert.Equal(t, byte(0x80), result)

	// Case 2: found at min value
	kv = &testutil.KeyValue{}
	kv.PutU([]byte{1}, []byte("value"))
	mockBackend.EXPECT().NewIterator(gomock.Any(), gomock.Any()).Return(iterator.NewArrayIterator(kv)).Times(8)
	mockBackend.EXPECT().NewIterator(gomock.Any(), gomock.Any()).Return(iterator.NewArrayIterator(kv)).Times(2)

	result, err = db.binarySearchForLastPrefixKey([]byte{1})
	assert.Nil(t, err)
	assert.Equal(t, byte(0x7f), result)

	// case 3: undefined behavior
	kv = &testutil.KeyValue{}
	kv.PutU([]byte{1}, []byte("value"))
	mockBackend.EXPECT().NewIterator(gomock.Any(), gomock.Any()).Return(iterator.NewArrayIterator(kv)).Times(9)

	result, err = db.binarySearchForLastPrefixKey([]byte{1})
	assert.NotNil(t, err)
	assert.Equal(t, byte(0), result)
}
