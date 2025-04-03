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

package state

import (
	"fmt"

	"github.com/0xsoniclabs/aida/txcontext"
	"github.com/0xsoniclabs/substate/substate"
	stypes "github.com/0xsoniclabs/substate/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/prque"
	"github.com/ethereum/go-ethereum/core/rawdb"
	geth "github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/stateless"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/ethdb/leveldb"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/trie/utils"
	"github.com/ethereum/go-ethereum/triedb"
	"github.com/holiman/uint256"
)

const (
	triesInMemory    = 1
	memoryUpperLimit = 256 * 1024 * 1024
	imgUpperLimit    = 4 * 1024 * 1024
)

func MakeGethStateDB(directory, variant string, rootHash common.Hash, isArchiveMode bool, chainConduit *ChainConduit) (StateDB, error) {
	if variant != "" {
		return nil, fmt.Errorf("unknown variant: %v", variant)
	}
	const cacheSize = 512
	const fileHandle = 128
	ldb, err := leveldb.New(directory, cacheSize, fileHandle, "", false)
	if err != nil {
		return nil, fmt.Errorf("failed to create a new Level DB, %w", err)
	}
	trieDb := triedb.NewDatabase(rawdb.NewDatabase(ldb), &triedb.Config{})
	evmState := geth.NewDatabase(trieDb, nil)
	if rootHash == (common.Hash{}) {
		rootHash = types.EmptyRootHash
	}
	db, err := geth.New(rootHash, evmState)
	if err != nil {
		return nil, err
	}

	return &GethStateDB{
		Db:                db,
		evmState:          evmState,
		stateRoot:         rootHash,
		triegc:            prque.New[uint64, common.Hash](nil),
		isArchiveMode:     isArchiveMode,
		chainConduit:      chainConduit,
		backend:           ldb,
		accessEvents:      geth.NewAccessEvents(utils.NewPointCache(4096)),
		SubstatePreAlloc:  substate.NewWorldState(),
		SubstatePostAlloc: substate.NewWorldState(),
		AccessedStorage:   make(map[common.Address]map[common.Hash]common.Hash),
	}, nil
}

// openStateDB creates a new statedb from an existing geth database
func (s *GethStateDB) openStateDB() error {
	var err error
	s.Db, err = geth.New(s.stateRoot, s.evmState)
	return err
}

type GethStateDB struct {
	Db                vm.StateDB    // statedb
	evmState          geth.Database // key-value database
	stateRoot         common.Hash   // lastest root hash
	triegc            *prque.Prque[uint64, common.Hash]
	isArchiveMode     bool
	chainConduit      *ChainConduit // chain configuration
	block             uint64
	backend           *leveldb.Database
	accessEvents      *geth.AccessEvents
	SubstatePreAlloc  substate.WorldState
	SubstatePostAlloc substate.WorldState
	AccessedStorage   map[common.Address]map[common.Hash]common.Hash
}

func (c *GethStateDB) substateRecordAccess(addr common.Address) {
	if c.Db.Exist(addr) && !c.Db.HasSelfDestructed(addr) {
		// insert the account in StateDB.SubstatePreAlloc
		if _, exist := c.SubstatePreAlloc[stypes.Address(addr)]; !exist {
			c.SubstatePreAlloc[stypes.Address(addr)] = substate.NewAccount(c.Db.GetNonce(addr), c.Db.GetBalance(addr).ToBig(), c.Db.GetCode(addr))
		}
	}

	// insert empty account in StateDB.SubstatePreAlloc
	// This will prevent insertion of new account created in txs
	if _, exist := c.SubstatePreAlloc[stypes.Address(addr)]; !exist {
		c.SubstatePreAlloc[stypes.Address(addr)] = nil
	}
}

func (c *GethStateDB) substateStorageAccess(addr common.Address, key common.Hash, value common.Hash) {
	if l, found := c.AccessedStorage[addr]; found {
		if _, f2 := l[key]; !f2 {
			c.AccessedStorage[addr][key] = value
		}
	} else {
		c.AccessedStorage[addr] = make(map[common.Hash]common.Hash)
		c.AccessedStorage[addr][key] = value
	}
}

func (s *GethStateDB) CreateAccount(addr common.Address) {
	s.substateRecordAccess(addr)
	s.Db.CreateAccount(addr)
}

func (s *GethStateDB) CreateContract(addr common.Address) {
	s.substateRecordAccess(addr)
	s.Db.CreateContract(addr)
}

func (s *GethStateDB) Exist(addr common.Address) bool {
	s.substateRecordAccess(addr)
	return s.Db.Exist(addr)
}

func (s *GethStateDB) Empty(addr common.Address) bool {
	s.substateRecordAccess(addr)
	return s.Db.Empty(addr)
}

func (s *GethStateDB) SelfDestruct(addr common.Address) uint256.Int {
	s.substateRecordAccess(addr)
	return s.Db.SelfDestruct(addr)
}

func (s *GethStateDB) SelfDestruct6780(addr common.Address) (uint256.Int, bool) {
	s.substateRecordAccess(addr)
	return s.Db.SelfDestruct6780(addr)
}

func (s *GethStateDB) HasSelfDestructed(addr common.Address) bool {
	s.substateRecordAccess(addr)
	return s.Db.HasSelfDestructed(addr)
}

func (s *GethStateDB) GetBalance(addr common.Address) *uint256.Int {
	s.substateRecordAccess(addr)
	return s.Db.GetBalance(addr)
}

func (s *GethStateDB) AddBalance(addr common.Address, value *uint256.Int, reason tracing.BalanceChangeReason) uint256.Int {
	s.substateRecordAccess(addr)
	return s.Db.AddBalance(addr, value, reason)
}

func (s *GethStateDB) SubBalance(addr common.Address, value *uint256.Int, reason tracing.BalanceChangeReason) uint256.Int {
	s.substateRecordAccess(addr)
	return s.Db.SubBalance(addr, value, reason)
}

func (s *GethStateDB) GetNonce(addr common.Address) uint64 {
	s.substateRecordAccess(addr)
	return s.Db.GetNonce(addr)
}

func (s *GethStateDB) SetNonce(addr common.Address, value uint64, reason tracing.NonceChangeReason) {
	s.substateRecordAccess(addr)
	s.Db.SetNonce(addr, value, reason)
}

func (s *GethStateDB) GetCommittedState(addr common.Address, key common.Hash) common.Hash {
	s.substateRecordAccess(addr)
	return s.Db.GetCommittedState(addr, key)
}

func (s *GethStateDB) GetState(addr common.Address, key common.Hash) common.Hash {
	value := s.Db.GetState(addr, key)

	s.substateRecordAccess(addr)
	s.substateStorageAccess(addr, key, value)
	return s.Db.GetState(addr, key)
}

func (s *GethStateDB) SetState(addr common.Address, key common.Hash, value common.Hash) common.Hash {
	s.substateRecordAccess(addr)
	s.substateStorageAccess(addr, key, value)
	return s.Db.SetState(addr, key, value)
}

func (s *GethStateDB) GetStorageRoot(addr common.Address) common.Hash {
	s.substateRecordAccess(addr)
	return s.Db.GetStorageRoot(addr)
}

func (s *GethStateDB) GetTransientState(addr common.Address, key common.Hash) common.Hash {
	s.substateRecordAccess(addr)
	return s.Db.GetTransientState(addr, key)
}

func (s *GethStateDB) SetTransientState(addr common.Address, key common.Hash, value common.Hash) {
	s.substateRecordAccess(addr)
	s.Db.SetTransientState(addr, key, value)
}

func (s *GethStateDB) GetCode(addr common.Address) []byte {
	s.substateRecordAccess(addr)
	return s.Db.GetCode(addr)
}

func (s *GethStateDB) GetCodeHash(addr common.Address) common.Hash {
	s.substateRecordAccess(addr)
	return s.Db.GetCodeHash(addr)
}

func (s *GethStateDB) GetCodeSize(addr common.Address) int {
	s.substateRecordAccess(addr)
	return s.Db.GetCodeSize(addr)
}

func (s *GethStateDB) SetCode(addr common.Address, code []byte) []byte {
	s.substateRecordAccess(addr)
	return s.Db.SetCode(addr, code)
}

func (s *GethStateDB) Snapshot() int {
	return s.Db.Snapshot()
}

func (s *GethStateDB) RevertToSnapshot(id int) {
	s.Db.RevertToSnapshot(id)
}

func (s *GethStateDB) Error() error {
	// TODO return geth's dberror
	return nil
}

func (s *GethStateDB) BeginTransaction(number uint32) error {
	// ignored
	return nil
}

func (s *GethStateDB) EndTransaction() error {
	if s.chainConduit == nil || s.chainConduit.IsFinalise(s.block) {
		// Opera or Ethereum after Byzantium
		s.Finalise(true)
	} else {
		// Ethereum before Byzantium
		s.IntermediateRoot(s.chainConduit.DeleteEmptyObjects(s.block))
	}
	return nil
}

func (s *GethStateDB) BeginBlock(number uint64) error {
	if err := s.openStateDB(); err != nil {
		return fmt.Errorf("cannot open geth state-db; %w", err)
	}
	s.block = number
	return nil
}

func (s *GethStateDB) EndBlock() error {
	var err error
	//commit at the end of a block
	s.stateRoot, err = s.Commit(s.block, true)
	if err != nil {
		panic("StateDB commit failed")
	}
	// if archival node, flush trie to disk after each block
	if s.evmState != nil {
		if err = s.trieCommit(); err != nil {
			return fmt.Errorf("cannot commit trie; %w", err)
		}
		s.trieCap()
	}
	return nil
}

func (s *GethStateDB) BeginSyncPeriod(number uint64) {
	// ignored
}

func (s *GethStateDB) EndSyncPeriod() {
	// if not archival node, flush trie to disk after each sync-period
	if s.evmState != nil && !s.isArchiveMode {
		s.trieCleanCommit()
		s.trieCap()
	}
}

func (s *GethStateDB) GetHash() (common.Hash, error) {
	return s.IntermediateRoot(true), nil
}

func (s *GethStateDB) Finalise(deleteEmptyObjects bool) {
	if db, ok := s.Db.(*geth.StateDB); ok {
		db.Finalise(deleteEmptyObjects)
	}
}

func (s *GethStateDB) IntermediateRoot(deleteEmptyObjects bool) common.Hash {
	if db, ok := s.Db.(*geth.StateDB); ok {
		return db.IntermediateRoot(deleteEmptyObjects)
	}
	return common.Hash{}
}

func (s *GethStateDB) Commit(block uint64, deleteEmptyObjects bool) (common.Hash, error) {
	if db, ok := s.Db.(*geth.StateDB); ok {
		return db.Commit(block, deleteEmptyObjects, false)
	}
	return common.Hash{}, nil
}

func (s *GethStateDB) SetTxContext(thash common.Hash, ti int) {
	if db, ok := s.Db.(*geth.StateDB); ok {
		db.SetTxContext(thash, ti)
	}
	return
}

func (s *GethStateDB) PrepareSubstate(substate txcontext.WorldState, block uint64) {
	// ignored
}

func (s *GethStateDB) GetSubstatePostAlloc() txcontext.WorldState {
	//TODO reenable equal check
	//if db, ok := s.db.(*geth.StateDB); ok {
	//	return substatecontext.NewWorldState(db.GetPostWorldState())
	//}

	return nil
}

func (s *GethStateDB) Close() error {
	// Skip closing if implementation is not Geth based.
	state, ok := s.Db.(*geth.StateDB)
	if !ok {
		return nil
	}
	// Commit data to trie.
	hash, err := state.Commit(s.block, true, false)
	if err != nil {
		return err
	}

	// Close underlying trie caching intermediate results.
	tdb := state.Database().TrieDB()
	if err := tdb.Commit(hash, true); err != nil {
		return err
	}
	// Close underlying LevelDB instance.
	if err := tdb.Close(); err != nil {
		return err
	}
	return s.backend.Close()
}

func (s *GethStateDB) AddRefund(gas uint64) {
	s.Db.AddRefund(gas)
}

func (s *GethStateDB) SubRefund(gas uint64) {
	s.Db.SubRefund(gas)
}
func (s *GethStateDB) GetRefund() uint64 {
	return s.Db.GetRefund()
}
func (s *GethStateDB) Prepare(rules params.Rules, sender, coinbase common.Address, dest *common.Address, precompiles []common.Address, txAccesses types.AccessList) {
	s.Db.Prepare(rules, sender, coinbase, dest, precompiles, txAccesses)
}

func (s *GethStateDB) AddressInAccessList(addr common.Address) bool {
	return s.Db.AddressInAccessList(addr)
}
func (s *GethStateDB) SlotInAccessList(addr common.Address, slot common.Hash) (addressOk bool, slotOk bool) {
	return s.Db.SlotInAccessList(addr, slot)
}
func (s *GethStateDB) AddAddressToAccessList(addr common.Address) {
	s.Db.AddAddressToAccessList(addr)
}
func (s *GethStateDB) AddSlotToAccessList(addr common.Address, slot common.Hash) {
	s.Db.AddSlotToAccessList(addr, slot)
}

func (s *GethStateDB) AddLog(log *types.Log) {
	s.Db.AddLog(log)
}
func (s *GethStateDB) AddPreimage(hash common.Hash, preimage []byte) {
	panic("Add Preimage")
	s.Db.AddPreimage(hash, preimage)
}

func (s *GethStateDB) AccessEvents() *geth.AccessEvents {
	return s.accessEvents
}

func (s *GethStateDB) GetLogs(hash common.Hash, block uint64, blockHash common.Hash) []*types.Log {
	if db, ok := s.Db.(*geth.StateDB); ok {
		return db.GetLogs(hash, block, blockHash)
	}
	return []*types.Log{}
}

func (s *GethStateDB) PointCache() *utils.PointCache {
	return s.Db.PointCache()
}

func (s *GethStateDB) Witness() *stateless.Witness {
	return s.Db.Witness()
}

func (s *GethStateDB) StartBulkLoad(block uint64) (BulkLoad, error) {
	if err := s.BeginBlock(block); err != nil {
		return nil, err
	}
	if err := s.BeginTransaction(0); err != nil {
		return nil, err
	}
	return &gethBulkLoad{db: s}, nil
}

func (s *GethStateDB) GetArchiveState(block uint64) (NonCommittableStateDB, error) {
	return nil, fmt.Errorf("archive states are not (yet) supported by this DB implementation")
}

func (s *GethStateDB) GetArchiveBlockHeight() (uint64, bool, error) {
	return 0, false, fmt.Errorf("archive states are not (yet) supported by this DB implementation")
}

func (s *GethStateDB) GetMemoryUsage() *MemoryUsage {
	// not supported yet
	return &MemoryUsage{uint64(0), nil}
}

type gethBulkLoad struct {
	db    *GethStateDB
	block uint64
}

func (l *gethBulkLoad) CreateAccount(addr common.Address) {
	l.db.CreateAccount(addr)
}

func (l *gethBulkLoad) SetBalance(addr common.Address, value *uint256.Int) {
	old := l.db.GetBalance(addr)
	value = value.Sub(value, old)
	l.db.AddBalance(addr, value, tracing.BalanceChangeUnspecified)
}

func (l *gethBulkLoad) SetNonce(addr common.Address, nonce uint64) {
	l.db.SetNonce(addr, nonce, tracing.NonceChangeGenesis)
}

func (l *gethBulkLoad) SetState(addr common.Address, key common.Hash, value common.Hash) {
	l.db.SetState(addr, key, value)
}

func (l *gethBulkLoad) SetCode(addr common.Address, code []byte) {
	l.db.SetCode(addr, code)
}

func (l *gethBulkLoad) Close() error {
	l.db.EndTransaction()
	l.db.EndBlock()
	_, err := l.db.Commit(l.block, false)
	l.block++
	return err
}

// trieCommit commits changes to disk if archive node; otherwise, performs garbage collection.
func (s *GethStateDB) trieCommit() error {
	triedb := s.evmState.TrieDB()
	// If we're applying genesis or running an archive node, always flush
	if s.isArchiveMode {
		if err := triedb.Commit(s.stateRoot, false); err != nil {
			return fmt.Errorf("Failed to flush trie DB into main DB. %v", err)
		}
	} else {
		// Full but not archive node, do proper garbage collection
		triedb.Reference(s.stateRoot, common.Hash{}) // metadata reference to keep trie alive
		s.triegc.Push(s.stateRoot, s.block)

		if current := s.block; current > triesInMemory {
			// If we exceeded our memory allowance, flush matured singleton nodes to disk
			s.trieCap()

			// Find the next state trie we need to commit
			chosen := current - triesInMemory

			// Garbage collect all below the chosen block
			for !s.triegc.Empty() {
				root, number := s.triegc.Pop()
				if uint64(-number) > chosen {
					s.triegc.Push(root, number)
					break
				}
				triedb.Dereference(root)
			}
		}
	}
	return nil
}

// trieCleanCommit cleans old state trie and commit changes.
func (s *GethStateDB) trieCleanCommit() error {
	// Don't need to reference the current state root
	// due to it already be referenced on `Commit()` function
	triedb := s.evmState.TrieDB()
	if current := s.block; current > triesInMemory {
		// Find the next state trie we need to commit
		chosen := current - triesInMemory
		// Garbage collect all below the chosen block
		for !s.triegc.Empty() {
			root, number := s.triegc.Pop()
			if uint64(-number) > chosen {
				s.triegc.Push(root, number)
				break
			}
			triedb.Dereference(root)
		}
	}
	// commit the state trie after clean up
	err := triedb.Commit(s.stateRoot, false)
	return err
}

// trieCap flushes matured singleton nodes to disk.
func (s *GethStateDB) trieCap() {
	triedb := s.evmState.TrieDB()
	_, nodes, imgs := triedb.Size()
	if nodes > memoryUpperLimit+ethdb.IdealBatchSize || imgs > imgUpperLimit {
		//If we exceeded our memory allowance, flush matured singleton nodes to disk
		triedb.Cap(memoryUpperLimit)
	}
}

func (s *GethStateDB) GetShadowDB() StateDB {
	return nil
}
