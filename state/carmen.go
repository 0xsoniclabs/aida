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
	"bytes"
	"fmt"
	"strings"

	"github.com/0xsoniclabs/aida/txcontext"
	"github.com/0xsoniclabs/carmen/go/carmen"
	_ "github.com/0xsoniclabs/carmen/go/state/cppstate"
	_ "github.com/0xsoniclabs/carmen/go/state/gostate"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/stateless"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/trie/utils"
	"github.com/holiman/uint256"
)

func MakeDefaultCarmenStateDB(dir string, variant string, schema int, archive string) (StateDB, error) {
	return MakeCarmenStateDB(dir, variant, schema, archive, 0, 0, 0, 0)
}

func MakeCarmenStateDB(
	dir string,
	variant string,
	schema int,
	archive string,
	liveDbCacheSize,
	archiveCacheSize,
	checkpointInterval,
	checkpointPeriod int,
) (StateDB, error) {
	var archiveType carmen.Archive

	switch strings.ToLower(archive) {
	case "none":
		archiveType = ""
	case "": // = default option
		fallthrough
	case "ldb":
		fallthrough
	case "leveldb":
		archiveType = "ldb"
	case "sql":
		fallthrough
	case "sqlite":
		archiveType = "sql"
	case "s4":
		archiveType = "s4"
	case "s5":
		archiveType = "s5"
	default:
		return nil, fmt.Errorf("unsupported archive type: %s", archive)
	}

	cfg := carmen.Configuration{
		Variant: carmen.Variant(variant),
		Schema:  carmen.Schema(schema),
		Archive: archiveType,
	}

	properties := make(carmen.Properties)
	if liveDbCacheSize > 0 {
		properties.SetInteger(carmen.LiveDBCache, liveDbCacheSize)
	}

	if archiveCacheSize > 0 {
		properties.SetInteger(carmen.ArchiveCache, archiveCacheSize)
	}

	// How often will Carmen create checkpoints - in blocks
	if checkpointInterval > 0 {
		properties.SetInteger(carmen.CheckpointInterval, checkpointInterval)
	}

	// How often will Carmen create checkpoints - in minutes
	if checkpointPeriod > 0 {
		properties.SetInteger(carmen.CheckpointPeriod, checkpointPeriod)
	}

	db, err := carmen.OpenDatabase(dir, cfg, properties)
	if err != nil {
		return nil, fmt.Errorf("cannot open carmen database; %w", err)
	}

	return &carmenHeadState{
		carmenStateDB: carmenStateDB{
			db:           db,
			accessEvents: state.NewAccessEvents(utils.NewPointCache(4096)),
		},
	}, nil
}

type carmenStateDB struct {
	db           carmen.Database
	txCtx        carmen.TransactionContext
	accessEvents *state.AccessEvents
}

type carmenHeadState struct {
	carmenStateDB
	blkCtx carmen.HeadBlockContext
}

type carmenHistoricState struct {
	carmenStateDB
	blkCtx    carmen.HistoricBlockContext
	blkNumber uint64
}

func (s *carmenStateDB) CreateAccount(addr common.Address) {
	s.txCtx.CreateAccount(carmen.Address(addr))
}

func (s *carmenStateDB) CreateContract(addr common.Address) {
	s.txCtx.CreateContract(carmen.Address(addr))
}

func (s *carmenStateDB) Exist(addr common.Address) bool {
	return s.txCtx.Exist(carmen.Address(addr))
}

func (s *carmenStateDB) Empty(addr common.Address) bool {
	return s.txCtx.Empty(carmen.Address(addr))
}

func (s *carmenStateDB) SelfDestruct(addr common.Address) uint256.Int {
	before := s.txCtx.GetBalance(carmen.Address(addr)).Uint256()
	s.txCtx.SelfDestruct(carmen.Address(addr))
	return before
}

func (s *carmenStateDB) SelfDestruct6780(addr common.Address) (uint256.Int, bool) {
	before := s.txCtx.GetBalance(carmen.Address(addr)).Uint256()
	return before, s.txCtx.SelfDestruct6780(carmen.Address(addr))
}

func (s *carmenStateDB) HasSelfDestructed(addr common.Address) bool {
	return s.txCtx.HasSelfDestructed(carmen.Address(addr))
}

func (s *carmenStateDB) GetBalance(addr common.Address) *uint256.Int {
	value := s.txCtx.GetBalance(carmen.Address(addr)).Uint256()
	return &value
}

func (s *carmenStateDB) AddBalance(addr common.Address, value *uint256.Int, _ tracing.BalanceChangeReason) uint256.Int {
	before := s.txCtx.GetBalance(carmen.Address(addr)).Uint256()
	s.txCtx.AddBalance(carmen.Address(addr), carmen.NewAmountFromUint256(value))
	return before
}

func (s *carmenStateDB) SubBalance(addr common.Address, value *uint256.Int, _ tracing.BalanceChangeReason) uint256.Int {
	before := s.txCtx.GetBalance(carmen.Address(addr)).Uint256()
	s.txCtx.SubBalance(carmen.Address(addr), carmen.NewAmountFromUint256(value))
	return before
}

func (s *carmenStateDB) GetNonce(addr common.Address) uint64 {
	return s.txCtx.GetNonce(carmen.Address(addr))
}

func (s *carmenStateDB) SetNonce(addr common.Address, value uint64, reason tracing.NonceChangeReason) {
	s.txCtx.SetNonce(carmen.Address(addr), value)
}

func (s *carmenStateDB) GetCommittedState(addr common.Address, key common.Hash) common.Hash {
	return common.Hash(s.txCtx.GetCommittedState(carmen.Address(addr), carmen.Key(key)))
}

func (s *carmenStateDB) GetState(addr common.Address, key common.Hash) common.Hash {
	return common.Hash(s.txCtx.GetState(carmen.Address(addr), carmen.Key(key)))
}

func (s *carmenStateDB) SetState(addr common.Address, key common.Hash, value common.Hash) common.Hash {
	before := s.txCtx.GetState(carmen.Address(addr), carmen.Key(key))
	s.txCtx.SetState(carmen.Address(addr), carmen.Key(key), carmen.Value(value))
	return common.Hash(before)
}

func (s *carmenStateDB) GetStorageRoot(addr common.Address) common.Hash {
	empty := s.txCtx.HasEmptyStorage(carmen.Address(addr))
	var h common.Hash
	if !empty {
		// Carmen does not provide a method to get the storage root for performance reasons
		// as getting a storage root needs computation of hashes in the trie.
		// In practice, the method GetStorageRoot here is used in the EVM only to assess
		// if the storage is empty. For this reason, this method returns a dummy hash here just
		// not to equal to the empty hash when the storage is not empty.
		h[0] = 1
	}
	return h
}

func (s *carmenStateDB) SetTransientState(addr common.Address, key common.Hash, value common.Hash) {
	s.txCtx.SetTransientState(carmen.Address(addr), carmen.Key(key), carmen.Value(value))
}

func (s *carmenStateDB) GetTransientState(address common.Address, key common.Hash) common.Hash {
	return common.Hash(s.txCtx.GetTransientState(carmen.Address(address), carmen.Key(key)))
}

func (s *carmenStateDB) GetCode(addr common.Address) []byte {
	return s.txCtx.GetCode(carmen.Address(addr))
}

func (s *carmenStateDB) GetCodeSize(addr common.Address) int {
	return s.txCtx.GetCodeSize(carmen.Address(addr))
}

func (s *carmenStateDB) GetCodeHash(addr common.Address) common.Hash {
	return common.Hash(s.txCtx.GetCodeHash(carmen.Address(addr)))
}

func (s *carmenStateDB) SetCode(addr common.Address, code []byte) []byte {
	before := bytes.Clone(s.GetCode(addr))
	s.txCtx.SetCode(carmen.Address(addr), code)
	return before
}

func (s *carmenStateDB) Snapshot() int {
	return s.txCtx.Snapshot()
}

func (s *carmenStateDB) RevertToSnapshot(id int) {
	s.txCtx.RevertToSnapshot(id)
}

func (s *carmenHeadState) BeginTransaction(uint32) error {
	var err error
	s.txCtx, err = s.blkCtx.BeginTransaction()
	return err
}

func (s *carmenStateDB) EndTransaction() error {
	return s.txCtx.Commit()
}

func (s *carmenHeadState) BeginBlock(block uint64) error {
	var err error
	s.blkCtx, err = s.db.BeginBlock(block)
	return err
}

func (s *carmenHeadState) EndBlock() error {
	return s.blkCtx.Commit()
}

func (s *carmenHeadState) BeginSyncPeriod(number uint64) {
	// ignored for Carmen
}

func (s *carmenHeadState) EndSyncPeriod() {
	// ignored for Carmen
}

func (s *carmenStateDB) GetHash() (common.Hash, error) {
	var hash common.Hash
	err := s.db.QueryHeadState(func(ctxt carmen.QueryContext) {
		hash = common.Hash(ctxt.GetStateHash())
	})
	return hash, err
}

func (s *carmenStateDB) Close() error {
	return s.db.Close()
}

func (s *carmenStateDB) AddRefund(amount uint64) {
	s.txCtx.AddRefund(amount)
}

func (s *carmenStateDB) SubRefund(amount uint64) {
	s.txCtx.SubRefund(amount)
}

func (s *carmenStateDB) GetRefund() uint64 {
	return s.txCtx.GetRefund()
}

func (s *carmenStateDB) Prepare(rules params.Rules, sender, coinbase common.Address, dest *common.Address, precompiles []common.Address, txAccesses types.AccessList) {
	s.txCtx.ClearAccessList()
	s.txCtx.AddAddressToAccessList(carmen.Address(sender))
	if dest != nil {
		s.txCtx.AddAddressToAccessList(carmen.Address(*dest))
	}
	for _, addr := range precompiles {
		s.txCtx.AddAddressToAccessList(carmen.Address(addr))
	}
	for _, el := range txAccesses {
		s.txCtx.AddAddressToAccessList(carmen.Address(el.Address))
		for _, key := range el.StorageKeys {
			s.txCtx.AddSlotToAccessList(carmen.Address(el.Address), carmen.Key(key))
		}
	}
	if rules.IsShanghai {
		s.txCtx.AddAddressToAccessList(carmen.Address(coinbase))
	}
}

func (s *carmenStateDB) AddressInAccessList(addr common.Address) bool {
	return s.txCtx.IsAddressInAccessList(carmen.Address(addr))
}

func (s *carmenStateDB) SlotInAccessList(addr common.Address, slot common.Hash) (addressOk bool, slotOk bool) {
	return s.txCtx.IsSlotInAccessList(carmen.Address(addr), carmen.Key(slot))
}

func (s *carmenStateDB) AddAddressToAccessList(addr common.Address) {
	s.txCtx.AddAddressToAccessList(carmen.Address(addr))
}

func (s *carmenStateDB) AddSlotToAccessList(addr common.Address, slot common.Hash) {
	s.txCtx.AddSlotToAccessList(carmen.Address(addr), carmen.Key(slot))
}

func (s *carmenStateDB) AddLog(log *types.Log) {
	topics := make([]carmen.Hash, 0, len(log.Topics))
	for _, topic := range log.Topics {
		topics = append(topics, carmen.Hash(topic))
	}
	s.txCtx.AddLog(&carmen.Log{
		Address: carmen.Address(log.Address),
		Topics:  topics,
		Data:    log.Data,
	})
}

func (s *carmenStateDB) GetLogs(common.Hash, uint64, common.Hash, uint64) []*types.Log {
	list := s.txCtx.GetLogs()

	res := make([]*types.Log, 0, len(list))
	for _, log := range list {
		topics := make([]common.Hash, 0, len(log.Topics))
		for _, topic := range log.Topics {
			topics = append(topics, common.Hash(topic))
		}
		res = append(res, &types.Log{
			Address: common.Address(log.Address),
			Topics:  topics,
			Data:    log.Data,
		})

	}
	return res
}

func (s *carmenStateDB) PointCache() *utils.PointCache {
	// this should not be relevant for revisions up to Cancun
	panic("PointCache not implemented")
}

func (s *carmenStateDB) Witness() *stateless.Witness {
	// this should not be relevant for revisions up to Cancun
	return nil
}

func (s *carmenStateDB) Finalise(bool) {
	// ignored
}

func (s *carmenStateDB) IntermediateRoot(bool) common.Hash {
	// ignored
	return common.Hash{}
}

func (s *carmenStateDB) Commit(uint64, bool) (common.Hash, error) {
	// ignored
	return common.Hash{}, nil
}

func (s *carmenStateDB) SetTxContext(common.Hash, int) {
	// ignored
}

func (s *carmenStateDB) PrepareSubstate(txcontext.WorldState, uint64) {
	// ignored
}

func (s *carmenStateDB) GetSubstatePostAlloc() txcontext.WorldState {
	// ignored
	return nil
}

func (s *carmenStateDB) AddPreimage(common.Hash, []byte) {
	// ignored
	panic("AddPreimage not implemented")
}

func (s *carmenStateDB) AccessEvents() *state.AccessEvents {
	return s.accessEvents
}

func (s *carmenStateDB) Error() error {
	// ignored
	return nil
}

func (s *carmenHistoricState) BeginTransaction(uint32) error {
	var err error
	s.txCtx, err = s.blkCtx.BeginTransaction()
	return err
}

func (s *carmenHistoricState) GetHash() (common.Hash, error) {
	h, err := s.db.GetHistoricStateHash(s.blkNumber)
	return common.Hash(h), err
}

func (s *carmenStateDB) StartBulkLoad(block uint64) (BulkLoad, error) {
	bl, err := s.db.StartBulkLoad(block)
	if err != nil {
		return nil, fmt.Errorf("cannot start bulkload; %w", err)
	}
	return &carmenBulkLoad{bl}, nil
}

func (s *carmenHeadState) GetArchiveState(block uint64) (NonCommittableStateDB, error) {
	historicBlkCtx, err := s.db.GetHistoricContext(block)
	if err != nil {
		return nil, err
	}

	return &carmenHistoricState{
		carmenStateDB: carmenStateDB{
			db: s.db,
		},
		blkCtx:    historicBlkCtx,
		blkNumber: block,
	}, nil
}

func (s *carmenHeadState) GetArchiveBlockHeight() (uint64, bool, error) {
	blk, err := s.db.GetArchiveBlockHeight()
	if err != nil {
		return 0, false, err
	}
	if blk == -1 {
		return 0, true, nil
	}
	return uint64(blk), false, nil
}

func (s *carmenStateDB) GetMemoryUsage() *MemoryUsage {
	if s.db == nil {
		return &MemoryUsage{uint64(0), nil}
	}
	usage := s.db.GetMemoryFootprint()
	if usage == nil {
		return &MemoryUsage{uint64(0), nil}
	}
	return &MemoryUsage{uint64(usage.Total()), usage}
}

func (s *carmenStateDB) GetShadowDB() StateDB {
	return nil
}

func (s *carmenHistoricState) Release() error {
	return s.blkCtx.Close()
}

// ----------------------------------------------------------------------------
//                                  BulkLoad
// ----------------------------------------------------------------------------

type carmenBulkLoad struct {
	load carmen.BulkLoad
}

func (l *carmenBulkLoad) CreateAccount(addr common.Address) {
	l.load.CreateAccount(carmen.Address(addr))
}

func (l *carmenBulkLoad) SetBalance(addr common.Address, value *uint256.Int) {
	l.load.SetBalance(carmen.Address(addr), carmen.NewAmountFromUint256(value))
}

func (l *carmenBulkLoad) SetNonce(addr common.Address, nonce uint64) {
	l.load.SetNonce(carmen.Address(addr), nonce)
}

func (l *carmenBulkLoad) SetState(addr common.Address, key common.Hash, value common.Hash) {
	l.load.SetState(carmen.Address(addr), carmen.Key(key), carmen.Value(value))
}

func (l *carmenBulkLoad) SetCode(addr common.Address, code []byte) {
	l.load.SetCode(carmen.Address(addr), code)
}

func (l *carmenBulkLoad) Close() error {
	return l.load.Finalize()
}
