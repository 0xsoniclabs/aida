// Package db implements database interfaces for the world state manager.
package db

import (
	"fmt"
	"github.com/Fantom-foundation/Aida-Testing/world-state/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/rlp"
	"io"
	"log"
)

const (
	// CodePrefix represents a prefix added to the code hash to separate code and state data in the KV database.
	// CodePrefix + codeHash (256-bit) -> code
	CodePrefix = "1c"
)

var (
	BlockNumberKey = []byte("blockNumberKey")
)

// ZeroHash represents an empty hash.
var ZeroHash = common.Hash{}

// StateSnapshotDB represents the state snapshot database handle.
type StateSnapshotDB struct {
	hashing crypto.KeccakState
	Backend BackendDatabase
}

// BackendDatabase represents the underlying KV store used for the StateSnapshotDB
type BackendDatabase interface {
	ethdb.KeyValueReader
	ethdb.KeyValueWriter
	ethdb.Batcher
	ethdb.Iteratee
	ethdb.Stater
	ethdb.Compacter
	io.Closer
}

// OpenStateSnapshotDB opens state snapshot database at the given path.
func OpenStateSnapshotDB(path string) (*StateSnapshotDB, error) {
	backend, err := rawdb.NewLevelDBDatabase(path, 1024, 100, "substatedir", false)
	if err != nil {
		return nil, err
	}

	return &StateSnapshotDB{Backend: backend, hashing: crypto.NewKeccakState()}, nil
}

// MustCloseSnapshotDB closes the state snapshot database without raising an error.
func MustCloseSnapshotDB(db *StateSnapshotDB) {
	if db != nil {
		err := db.Backend.Close()
		if err != nil {
			log.Printf("could not close state snapshot; %s\n", err.Error())
		}
	}
}

// PutCode inserts Account code into database
func (db *StateSnapshotDB) PutCode(code []byte) (common.Hash, error) {
	// anything to store?
	if len(code) == 0 {
		return common.Hash{}, nil
	}

	codeHash := crypto.Keccak256Hash(code)
	key := CodeKey(codeHash)
	return codeHash, db.Backend.Put(key, code)
}

// Code loads account code from the database, if available.
func (db *StateSnapshotDB) Code(h common.Hash) ([]byte, error) {
	return db.Backend.Get(CodeKey(h))
}

// PutAccount inserts Account into database
func (db *StateSnapshotDB) PutAccount(acc *types.Account) error {
	// store the code, if any
	if len(acc.Code) > 0 {
		ch, err := db.PutCode(acc.Code)
		if err != nil {
			return err
		}
		acc.CodeHash = ch.Bytes()
	} else {
		//when PutAccount is used to update account, then CodeHash needs to be either updated or explicitly set to nil
		acc.CodeHash = nil
	}

	enc, err := rlp.EncodeToBytes(acc)
	if err != nil {
		return fmt.Errorf("failed encoding account %s to RLP; %s", acc.Hash.String(), err.Error())
	}

	return db.Backend.Put(acc.Hash.Bytes(), enc)
}

// Account tries to read details of the given account address.
func (db *StateSnapshotDB) Account(addr common.Address) (*types.Account, error) {
	h := crypto.HashData(db.hashing, addr.Bytes())
	data, err := db.Backend.Get(h.Bytes())
	if err != nil {
		return nil, err
	}

	acc := types.Account{}
	err = rlp.DecodeBytes(data, &acc)
	if err != nil {
		return nil, err
	}

	// update the account hash
	acc.Hash = h

	// any code to be loaded?
	codeHash := common.Hash{}
	codeHash.SetBytes(acc.CodeHash)
	if codeHash != ZeroHash {
		acc.Code, err = db.Code(codeHash)
		if err != nil {
			return nil, err
		}
	}

	return &acc, nil
}

// CodeKey retrieves storing DB key for supplied codeHash
func CodeKey(codeHash common.Hash) []byte {
	prefix := []byte(CodePrefix)
	return append(prefix, codeHash.Bytes()...)
}

// PutBlockNumber inserts block number into database
func (db *StateSnapshotDB) PutBlockNumber(i uint64) error {
	enc, err := rlp.EncodeToBytes(i)
	if err != nil {
		return fmt.Errorf("failed encoding blockID %d to RLP; %s", i, err.Error())
	}

	return db.Backend.Put(BlockNumberKey, enc)
}

// GetBlockNumber retrieves block number from database
func (db *StateSnapshotDB) GetBlockNumber() (uint64, error) {
	data, err := db.Backend.Get(BlockNumberKey)
	if err != nil {
		return 0, fmt.Errorf("block number not found in database; %s", err.Error())
	}

	var blockNumber uint64
	err = rlp.DecodeBytes(data, &blockNumber)
	if err != nil {
		return 0, fmt.Errorf("failed decoding block number from RLP; %s", err.Error())
	}

	return blockNumber, err
}
