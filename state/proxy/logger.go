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

package proxy

import (
	"encoding/hex"
	"fmt"
	"sync"

	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/state"
	"github.com/0xsoniclabs/aida/txcontext"
	"github.com/ethereum/go-ethereum/common"
	geth "github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/stateless"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/trie/utils"
	"github.com/holiman/uint256"
)

// NewLoggerProxy wraps the given StateDB instance into a logging wrapper causing
// every StateDB operation (except BulkLoading) to be logged for debugging.
func NewLoggerProxy(db state.StateDB, log logger.Logger, output chan string, wg *sync.WaitGroup) state.StateDB {
	return &LoggingStateDb{
		loggingVmStateDb: loggingVmStateDb{
			db:     db,
			log:    log,
			output: output,
			wg:     wg,
		},

		state: db,
	}
}

type loggingVmStateDb struct {
	db     state.VmStateDB
	log    logger.Logger
	output chan string
	wg     *sync.WaitGroup
}

type loggingNonCommittableStateDb struct {
	loggingVmStateDb
	nonCommittableStateDB state.NonCommittableStateDB
}

type LoggingStateDb struct {
	loggingVmStateDb
	state state.StateDB
}

func (s *loggingVmStateDb) CreateAccount(addr common.Address) {
	s.db.CreateAccount(addr)
	s.writeLog("CreateAccount, %v", addr)
}

func (s *loggingVmStateDb) Exist(addr common.Address) bool {
	res := s.db.Exist(addr)
	s.writeLog("Exist, %v, %v", addr, res)
	return res
}

func (s *loggingVmStateDb) Empty(addr common.Address) bool {
	res := s.db.Empty(addr)
	s.writeLog("Empty, %v, %v", addr, res)
	return res
}

func (s *loggingVmStateDb) SelfDestruct(addr common.Address) uint256.Int {
	res := s.db.SelfDestruct(addr)
	s.writeLog("SelfDestruct, %v, %v", addr, res)
	return res
}

func (s *loggingVmStateDb) HasSelfDestructed(addr common.Address) bool {
	res := s.db.HasSelfDestructed(addr)
	s.writeLog("HasSelfDestructed, %v, %v", addr, res)
	return res
}

func (s *loggingVmStateDb) GetBalance(addr common.Address) *uint256.Int {
	res := s.db.GetBalance(addr)
	s.writeLog("GetBalance, %v, %v", addr, res)
	return res
}

func (s *loggingVmStateDb) AddBalance(addr common.Address, value *uint256.Int, reason tracing.BalanceChangeReason) uint256.Int {
	res := s.db.AddBalance(addr, value, reason)
	s.writeLog("AddBalance, %v, %v, %v, %v, %v", addr, value, s.db.GetBalance(addr), reason, res)
	return res
}

func (s *loggingVmStateDb) SubBalance(addr common.Address, value *uint256.Int, reason tracing.BalanceChangeReason) uint256.Int {
	res := s.db.SubBalance(addr, value, reason)
	s.writeLog("SubBalance, %v, %v, %v, %v, %v", addr, value, s.db.GetBalance(addr), reason, res)
	return res
}

func (s *loggingVmStateDb) GetNonce(addr common.Address) uint64 {
	res := s.db.GetNonce(addr)
	s.writeLog("GetNonce, %v, %v", addr, res)
	return res
}

func (s *loggingVmStateDb) SetNonce(addr common.Address, value uint64, reason tracing.NonceChangeReason) {
	s.db.SetNonce(addr, value, reason)
	s.writeLog("SetNonce, %v, %v, %v", addr, value, reason)
}

func (s *loggingVmStateDb) GetCommittedState(addr common.Address, key common.Hash) common.Hash {
	res := s.db.GetCommittedState(addr, key)
	s.writeLog("GetCommittedState, %v, %v, %v", addr, key, res)
	return res
}

func (s *loggingVmStateDb) GetState(addr common.Address, key common.Hash) common.Hash {
	res := s.db.GetState(addr, key)
	s.writeLog("GetState, %v, %v, %v", addr, key, res)
	return res
}

func (s *loggingVmStateDb) SetState(addr common.Address, key common.Hash, value common.Hash) common.Hash {
	res := s.db.SetState(addr, key, value)
	s.writeLog("SetState, %v, %v, %v, %v", addr, key, value, res)
	return res
}

func (s *loggingVmStateDb) SetTransientState(addr common.Address, key common.Hash, value common.Hash) {
	s.writeLog("SetTransientState, %v, %v, %v", addr, key, value)
	s.db.SetTransientState(addr, key, value)
}

func (s *loggingVmStateDb) GetTransientState(addr common.Address, key common.Hash) common.Hash {
	value := s.db.GetTransientState(addr, key)
	s.writeLog("GetTransientState, %v, %v, %v", addr, key, value)
	return value
}

func (s *loggingVmStateDb) GetCode(addr common.Address) []byte {
	res := s.db.GetCode(addr)
	s.writeLog("GetCode, %v, %v", addr, hex.EncodeToString(res))
	return res
}

func (s *loggingVmStateDb) GetCodeSize(addr common.Address) int {
	res := s.db.GetCodeSize(addr)
	s.writeLog("GetCodeSize, %v, %v", addr, res)
	return res
}

func (s *loggingVmStateDb) GetCodeHash(addr common.Address) common.Hash {
	res := s.db.GetCodeHash(addr)
	s.writeLog("GetCodeHash, %v, %v", addr, res)
	return res
}

func (s *loggingVmStateDb) SetCode(addr common.Address, code []byte) []byte {
	res := s.db.SetCode(addr, code)
	s.writeLog("SetCode, %v, %v, %v", addr, code, res)
	return res
}

func (s *loggingVmStateDb) Snapshot() int {
	res := s.db.Snapshot()
	s.writeLog("Snapshot, %v", res)
	return res
}

func (s *loggingVmStateDb) RevertToSnapshot(id int) {
	s.db.RevertToSnapshot(id)
	s.writeLog("RevertToSnapshot, %v", id)
}

func (s *LoggingStateDb) Error() error {
	err := s.state.Error()
	s.writeLog("Error, %v", err)
	return err
}

func (s *loggingVmStateDb) BeginTransaction(tx uint32) error {
	s.writeLog("BeginTransaction, %v", tx)
	return s.db.BeginTransaction(tx)
}

func (s *loggingVmStateDb) EndTransaction() error {
	s.writeLog("EndTransaction")
	return s.db.EndTransaction()
}

func (s *loggingVmStateDb) Finalise(deleteEmptyObjects bool) {
	s.writeLog("Finalise, %v", deleteEmptyObjects)
	s.db.Finalise(deleteEmptyObjects)
}

func (s *LoggingStateDb) BeginBlock(blk uint64) error {
	s.writeLog("BeginBlock, %v", blk)
	return s.state.BeginBlock(blk)
}

func (s *LoggingStateDb) EndBlock() error {
	s.writeLog("EndBlock")
	return s.state.EndBlock()
}

func (s *LoggingStateDb) BeginSyncPeriod(number uint64) {
	s.writeLog("BeginSyncPeriod, %v", number)
	s.state.BeginSyncPeriod(number)
}

func (s *LoggingStateDb) EndSyncPeriod() {
	s.writeLog("EndSyncPeriod")
	s.state.EndSyncPeriod()
}

func (s *LoggingStateDb) GetHash() (common.Hash, error) {
	hash, err := s.state.GetHash()
	s.writeLog("GetHash, %v", hash)
	return hash, err
}

func (s *loggingNonCommittableStateDb) GetHash() (common.Hash, error) {
	hash, err := s.nonCommittableStateDB.GetHash()
	if err != nil {
		s.writeLog("GetHash, %v", err)
		return common.Hash{}, err
	} else {
		s.writeLog("GetHash, %v", hash)
	}
	return hash, nil
}

func (s *LoggingStateDb) Close() error {
	res := s.state.Close()
	s.writeLog("Close")
	// signal and await the close
	close(s.output)
	s.wg.Wait()
	return res
}

func (s *loggingVmStateDb) AddRefund(amount uint64) {
	s.db.AddRefund(amount)
	s.writeLog("AddRefund, %v, %v", amount, s.db.GetRefund())
}

func (s *loggingVmStateDb) SubRefund(amount uint64) {
	s.db.SubRefund(amount)
	s.writeLog("SubRefund, %v, %v", amount, s.db.GetRefund())
}

func (s *loggingVmStateDb) GetRefund() uint64 {
	res := s.db.GetRefund()
	s.writeLog("GetRefund, %v", res)
	return res
}

func (s *loggingVmStateDb) Prepare(rules params.Rules, sender, coinbase common.Address, dest *common.Address, precompiles []common.Address, txAccesses types.AccessList) {
	s.writeLog("Prepare, %v, %v, %v, %v", sender, dest, precompiles, txAccesses)
	s.db.Prepare(rules, sender, coinbase, dest, precompiles, txAccesses)
}

func (s *loggingVmStateDb) AddressInAccessList(addr common.Address) bool {
	res := s.db.AddressInAccessList(addr)
	s.writeLog("AddressInAccessList, %v, %v", addr, res)
	return res
}

func (s *loggingVmStateDb) SlotInAccessList(addr common.Address, slot common.Hash) (addressOk bool, slotOk bool) {
	a, b := s.db.SlotInAccessList(addr, slot)
	s.writeLog("SlotInAccessList, %v, %v, %v, %v", addr, slot, a, b)
	return a, b
}

func (s *loggingVmStateDb) AddAddressToAccessList(addr common.Address) {
	s.writeLog("AddAddressToAccessList, %v", addr)
	s.db.AddAddressToAccessList(addr)
}

func (s *loggingVmStateDb) AddSlotToAccessList(addr common.Address, slot common.Hash) {
	s.writeLog("AddSlotToAccessList, %v, %v", addr, slot)
	s.db.AddSlotToAccessList(addr, slot)
}

func (s *loggingVmStateDb) AddLog(entry *types.Log) {
	s.writeLog("AddLog, %v", entry)
	s.db.AddLog(entry)
}

func (s *loggingVmStateDb) GetLogs(hash common.Hash, block uint64, blockHash common.Hash, blkTimestamp uint64) []*types.Log {
	res := s.db.GetLogs(hash, block, blockHash, blkTimestamp)
	s.writeLog("GetLogs, %v, %v, %v, %v, %v", hash, block, blockHash, blkTimestamp, res)
	return res
}

// PointCache returns the point cache used in computations.
func (s *loggingVmStateDb) PointCache() *utils.PointCache {
	res := s.db.PointCache()
	s.writeLog("PointCache, %v", res)
	return res
}

// Witness retrieves the current state witness.
func (s *loggingVmStateDb) Witness() *stateless.Witness {
	res := s.db.Witness()
	s.writeLog("Witness, %v", res)
	return res
}

func (s *LoggingStateDb) Finalise(deleteEmptyObjects bool) {
	s.writeLog("Finalise, %v", deleteEmptyObjects)
	s.state.Finalise(deleteEmptyObjects)
}

func (s *LoggingStateDb) IntermediateRoot(deleteEmptyObjects bool) common.Hash {
	res := s.state.IntermediateRoot(deleteEmptyObjects)
	s.writeLog("IntermediateRoot, %v, %v", deleteEmptyObjects, res)
	return res
}

func (s *LoggingStateDb) Commit(block uint64, deleteEmptyObjects bool) (common.Hash, error) {
	hash, err := s.state.Commit(block, deleteEmptyObjects)
	s.writeLog("Commit, %v, %v, %v", deleteEmptyObjects, hash, err)
	return hash, err
}

func (s *loggingVmStateDb) SetTxContext(thash common.Hash, ti int) {
	s.db.SetTxContext(thash, ti)
	s.writeLog("SetTxContext, %v, %v", thash, ti)
}

func (s *LoggingStateDb) PrepareSubstate(substate txcontext.WorldState, block uint64) {
	s.state.PrepareSubstate(substate, block)
	s.writeLog("PrepareSubstate, %v", substate.String())
}

func (s *loggingVmStateDb) GetSubstatePostAlloc() txcontext.WorldState {
	res := s.db.GetSubstatePostAlloc()
	s.writeLog("GetSubstatePostAlloc, %v", res.String())
	return res
}

func (s *loggingVmStateDb) AddPreimage(hash common.Hash, data []byte) {
	s.db.AddPreimage(hash, data)
	s.writeLog("AddPreimage, %v, %v", hash, data)
}

func (s *loggingVmStateDb) AccessEvents() *geth.AccessEvents {
	res := s.db.AccessEvents()
	s.writeLog("AccessEvents, %v", res)
	return res
}

func (s *LoggingStateDb) StartBulkLoad(block uint64) (state.BulkLoad, error) {
	bl, err := s.state.StartBulkLoad(block)
	if err != nil {
		return nil, fmt.Errorf("cannot start bulkload; %w", err)
	}
	return &loggingBulkLoad{
		nested:   bl,
		writeLog: s.writeLog,
	}, nil
}

func (s *LoggingStateDb) GetArchiveState(block uint64) (state.NonCommittableStateDB, error) {
	archive, err := s.state.GetArchiveState(block)
	if err != nil {
		return nil, err
	}
	return &loggingNonCommittableStateDb{
		loggingVmStateDb: loggingVmStateDb{
			db:     archive,
			log:    s.log,
			output: s.output,
		},
		nonCommittableStateDB: archive,
	}, nil
}

func (s *LoggingStateDb) GetArchiveBlockHeight() (uint64, bool, error) {
	res, empty, err := s.state.GetArchiveBlockHeight()
	s.writeLog("GetArchiveBlockHeight, %v, %t, %v", res, empty, err)
	return res, empty, err
}

func (s *LoggingStateDb) GetMemoryUsage() *state.MemoryUsage {
	// no logging in this case
	return s.state.GetMemoryUsage()
}

func (s *LoggingStateDb) GetShadowDB() state.StateDB {
	return s.state.GetShadowDB()
}

func (s *loggingVmStateDb) CreateContract(addr common.Address) {
	s.writeLog("CreateContract, %v", addr)
	s.db.CreateContract(addr)
}

func (s *loggingVmStateDb) SelfDestruct6780(addr common.Address) (uint256.Int, bool) {
	balance, success := s.db.SelfDestruct6780(addr)
	s.writeLog("SelfDestruct6780, %v, %v, %v", addr, balance, success)
	return balance, success
}

func (s *loggingVmStateDb) GetStorageRoot(addr common.Address) common.Hash {
	res := s.db.GetStorageRoot(addr)
	s.writeLog("GetStorageRoot, %v, %v", res, addr)
	return res
}

func (s *loggingNonCommittableStateDb) Release() error {
	s.writeLog("Release")
	s.nonCommittableStateDB.Release()
	return nil
}

type loggingBulkLoad struct {
	nested   state.BulkLoad
	writeLog func(format string, a ...any)
}

func (l *loggingBulkLoad) CreateAccount(addr common.Address) {
	l.nested.CreateAccount(addr)
	l.writeLog("Bulk, CreateAccount, %v", addr)
}
func (l *loggingBulkLoad) SetBalance(addr common.Address, balance *uint256.Int) {
	l.nested.SetBalance(addr, balance)
	l.writeLog("Bulk, SetBalance, %v, %v", addr, balance)
}

func (l *loggingBulkLoad) SetNonce(addr common.Address, nonce uint64) {
	l.nested.SetNonce(addr, nonce)
	l.writeLog("Bulk, SetNonce, %v, %v", addr, nonce)
}

func (l *loggingBulkLoad) SetState(addr common.Address, key common.Hash, value common.Hash) {
	l.nested.SetState(addr, key, value)
	l.writeLog("Bulk, SetState, %v, %v, %v", addr, key, value)
}

func (l *loggingBulkLoad) SetCode(addr common.Address, code []byte) {
	l.nested.SetCode(addr, code)
	l.writeLog("Bulk, SetCode, %v, %v", addr, code)
}

func (l *loggingBulkLoad) Close() error {
	res := l.nested.Close()
	l.writeLog("Bulk, Close, %v", res)
	return res
}

func (s *loggingVmStateDb) writeLog(format string, a ...any) {
	str := fmt.Sprintf(format, a...)
	s.output <- str
	s.log.Debug(str)
}
