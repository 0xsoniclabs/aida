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

type loggingBulkLoad struct {
	nested   state.BulkLoad
	writeLog func(format string, a ...any)
}

func (s *LoggingStateDb) Error() error {
	err := s.state.Error()
	s.writeLog("Error, %v", err)
	return err
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
	s.writeLog("GetHash")
	hash, err := s.state.GetHash()
	return hash, err
}

func (s *LoggingStateDb) Close() error {
	s.writeLog("Close")
	res := s.state.Close()
	// signal and await the close
	close(s.output)
	s.wg.Wait()
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
	s.writeLog("GetArchiveBlockHeight")
	res, empty, err := s.state.GetArchiveBlockHeight()
	return res, empty, err
}

func (s *LoggingStateDb) GetMemoryUsage() *state.MemoryUsage {
	// no logging in this case
	return s.state.GetMemoryUsage()
}

func (s *LoggingStateDb) GetShadowDB() state.StateDB {
	return s.state.GetShadowDB()
}

func (s *LoggingStateDb) Finalise(deleteEmptyObjects bool) {
	s.writeLog("Finalise, %v", deleteEmptyObjects)
	s.state.Finalise(deleteEmptyObjects)
}

func (s *LoggingStateDb) IntermediateRoot(deleteEmptyObjects bool) common.Hash {
	s.writeLog("IntermediateRoot, %v", deleteEmptyObjects)
	res := s.state.IntermediateRoot(deleteEmptyObjects)
	return res
}

func (s *LoggingStateDb) Commit(block uint64, deleteEmptyObjects bool) (common.Hash, error) {
	s.writeLog("Commit, %v, %v", block, deleteEmptyObjects)
	hash, err := s.state.Commit(block, deleteEmptyObjects)
	return hash, err
}

func (s *LoggingStateDb) PrepareSubstate(substate txcontext.WorldState, block uint64) {
	s.writeLog("PrepareSubstate, %v, %v", substate.String(), block)
	s.state.PrepareSubstate(substate, block)
}

func (s *loggingVmStateDb) CreateAccount(addr common.Address) {
	s.writeLog("CreateAccount, %v", addr)
	s.db.CreateAccount(addr)
}

func (s *loggingVmStateDb) Exist(addr common.Address) bool {
	s.writeLog("Exist, %v", addr)
	res := s.db.Exist(addr)
	return res
}

func (s *loggingVmStateDb) Empty(addr common.Address) bool {
	s.writeLog("Empty, %v", addr)
	res := s.db.Empty(addr)
	return res
}

func (s *loggingVmStateDb) SelfDestruct(addr common.Address) uint256.Int {
	s.writeLog("SelfDestruct, %v", addr)
	res := s.db.SelfDestruct(addr)
	return res
}

func (s *loggingVmStateDb) HasSelfDestructed(addr common.Address) bool {
	s.writeLog("HasSelfDestructed, %v", addr)
	res := s.db.HasSelfDestructed(addr)
	return res
}

func (s *loggingVmStateDb) GetBalance(addr common.Address) *uint256.Int {
	s.writeLog("GetBalance, %v", addr)
	res := s.db.GetBalance(addr)
	return res
}

func (s *loggingVmStateDb) AddBalance(addr common.Address, value *uint256.Int, reason tracing.BalanceChangeReason) uint256.Int {
	s.writeLog("AddBalance, %v, %v", addr, value)
	res := s.db.AddBalance(addr, value, reason)
	return res
}

func (s *loggingVmStateDb) SubBalance(addr common.Address, value *uint256.Int, reason tracing.BalanceChangeReason) uint256.Int {
	s.writeLog("SubBalance, %v, %v", addr, value)
	res := s.db.SubBalance(addr, value, reason)
	return res
}

func (s *loggingVmStateDb) GetNonce(addr common.Address) uint64 {
	s.writeLog("GetNonce, %v", addr)
	res := s.db.GetNonce(addr)
	return res
}

func (s *loggingVmStateDb) SetNonce(addr common.Address, value uint64, reason tracing.NonceChangeReason) {
	s.writeLog("SetNonce, %v, %v", addr, value)
	s.db.SetNonce(addr, value, reason)
}

func (s *loggingVmStateDb) GetCommittedState(addr common.Address, key common.Hash) common.Hash {
	s.writeLog("GetCommittedState, %v, %v", addr, key)
	res := s.db.GetCommittedState(addr, key)
	return res
}

func (s *loggingVmStateDb) GetStateAndCommittedState(addr common.Address, key common.Hash) (common.Hash, common.Hash) {
	s.writeLog("GetStateAndCommittedState, %s, %s", addr, key)
	val, origin := s.db.GetStateAndCommittedState(addr, key)
	return val, origin
}

func (s *loggingVmStateDb) GetState(addr common.Address, key common.Hash) common.Hash {
	s.writeLog("GetState, %v, %v", addr, key)
	res := s.db.GetState(addr, key)
	return res
}

func (s *loggingVmStateDb) SetState(addr common.Address, key common.Hash, value common.Hash) common.Hash {
	s.writeLog("SetState, %v, %v, %v", addr, key, value)
	res := s.db.SetState(addr, key, value)
	return res
}

func (s *loggingVmStateDb) SetTransientState(addr common.Address, key common.Hash, value common.Hash) {
	s.writeLog("SetTransientState, %v, %v, %v", addr, key, value)
	s.db.SetTransientState(addr, key, value)
}

func (s *loggingVmStateDb) GetTransientState(addr common.Address, key common.Hash) common.Hash {
	s.writeLog("GetTransientState, %v, %v", addr, key)
	value := s.db.GetTransientState(addr, key)
	return value
}

func (s *loggingVmStateDb) GetCode(addr common.Address) []byte {
	s.writeLog("GetCode, %v", addr)
	res := s.db.GetCode(addr)
	return res
}

func (s *loggingVmStateDb) GetCodeSize(addr common.Address) int {
	s.writeLog("GetCodeSize, %v", addr)
	res := s.db.GetCodeSize(addr)
	return res
}

func (s *loggingVmStateDb) GetCodeHash(addr common.Address) common.Hash {
	s.writeLog("GetCodeHash, %v", addr)
	res := s.db.GetCodeHash(addr)
	return res
}

func (s *loggingVmStateDb) SetCode(addr common.Address, code []byte, reason tracing.CodeChangeReason) []byte {
	s.writeLog("SetCode, %v, %v, %v", addr, hex.EncodeToString(code), reason)
	res := s.db.SetCode(addr, code, reason)
	return res
}

func (s *loggingVmStateDb) Snapshot() int {
	s.writeLog("Snapshot")
	res := s.db.Snapshot()
	return res
}

func (s *loggingVmStateDb) RevertToSnapshot(id int) {
	s.writeLog("RevertToSnapshot, %v", id)
	s.db.RevertToSnapshot(id)
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

func (s *loggingVmStateDb) AddRefund(amount uint64) {
	s.writeLog("AddRefund, %v", amount)
	s.db.AddRefund(amount)
}

func (s *loggingVmStateDb) SubRefund(amount uint64) {
	s.writeLog("SubRefund, %v", amount)
	s.db.SubRefund(amount)
}

func (s *loggingVmStateDb) GetRefund() uint64 {
	s.writeLog("GetRefund")
	res := s.db.GetRefund()
	return res
}

func (s *loggingVmStateDb) Prepare(rules params.Rules, sender, coinbase common.Address, dest *common.Address, precompiles []common.Address, txAccesses types.AccessList) {
	s.writeLog("Prepare, %v, %v, %v, %v", sender, dest, precompiles, txAccesses)
	s.db.Prepare(rules, sender, coinbase, dest, precompiles, txAccesses)
}

func (s *loggingVmStateDb) AddressInAccessList(addr common.Address) bool {
	s.writeLog("AddressInAccessList, %v", addr)
	res := s.db.AddressInAccessList(addr)
	return res
}

func (s *loggingVmStateDb) SlotInAccessList(addr common.Address, slot common.Hash) (addressOk bool, slotOk bool) {
	s.writeLog("SlotInAccessList, %v, %v", addr, slot)
	a, b := s.db.SlotInAccessList(addr, slot)
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
	s.writeLog("GetLogs, %v, %v, %v, %v", hash, block, blockHash, blkTimestamp)
	res := s.db.GetLogs(hash, block, blockHash, blkTimestamp)
	return res
}

// PointCache returns the point cache used in computations.
func (s *loggingVmStateDb) PointCache() *utils.PointCache {
	s.writeLog("PointCache")
	res := s.db.PointCache()
	return res
}

// Witness retrieves the current state witness.
func (s *loggingVmStateDb) Witness() *stateless.Witness {
	s.writeLog("Witness")
	res := s.db.Witness()
	return res
}

func (s *loggingVmStateDb) SetTxContext(thash common.Hash, ti int) {
	s.writeLog("SetTxContext, %v, %v", thash, ti)
	s.db.SetTxContext(thash, ti)
}

func (s *loggingVmStateDb) GetSubstatePostAlloc() txcontext.WorldState {
	s.writeLog("GetSubstatePostAlloc")
	res := s.db.GetSubstatePostAlloc()
	return res
}

func (s *loggingVmStateDb) AddPreimage(hash common.Hash, data []byte) {
	s.writeLog("AddPreimage, %v, %v", hash, hex.EncodeToString(data))
	s.db.AddPreimage(hash, data)
}

func (s *loggingVmStateDb) AccessEvents() *geth.AccessEvents {
	s.writeLog("AccessEvents")
	res := s.db.AccessEvents()
	return res
}

func (s *loggingVmStateDb) CreateContract(addr common.Address) {
	s.writeLog("CreateContract, %v", addr)
	s.db.CreateContract(addr)
}

func (s *loggingVmStateDb) SelfDestruct6780(addr common.Address) (uint256.Int, bool) {
	s.writeLog("SelfDestruct6780, %v", addr)
	balance, success := s.db.SelfDestruct6780(addr)
	return balance, success
}

func (s *loggingVmStateDb) GetStorageRoot(addr common.Address) common.Hash {
	s.writeLog("GetStorageRoot, %v", addr)
	res := s.db.GetStorageRoot(addr)
	return res
}

func (s *loggingVmStateDb) writeLog(format string, a ...any) {
	str := fmt.Sprintf(format, a...)
	s.output <- str
	s.log.Debug(str)
}

func (s *loggingNonCommittableStateDb) GetHash() (common.Hash, error) {
	s.writeLog("GetHash")
	hash, err := s.nonCommittableStateDB.GetHash()
	if err != nil {
		return common.Hash{}, err
	}
	return hash, nil
}

func (s *loggingNonCommittableStateDb) Release() error {
	s.writeLog("Release")
	return s.nonCommittableStateDB.Release()
}

func (l *loggingBulkLoad) CreateAccount(addr common.Address) {
	l.writeLog("Bulk, CreateAccount, %v", addr)
	l.nested.CreateAccount(addr)
}
func (l *loggingBulkLoad) SetBalance(addr common.Address, balance *uint256.Int) {
	l.writeLog("Bulk, SetBalance, %v, %v", addr, balance)
	l.nested.SetBalance(addr, balance)
}

func (l *loggingBulkLoad) SetNonce(addr common.Address, nonce uint64) {
	l.writeLog("Bulk, SetNonce, %v, %v", addr, nonce)
	l.nested.SetNonce(addr, nonce)
}

func (l *loggingBulkLoad) SetState(addr common.Address, key common.Hash, value common.Hash) {
	l.writeLog("Bulk, SetState, %v, %v, %v", addr, key, value)
	l.nested.SetState(addr, key, value)
}

func (l *loggingBulkLoad) SetCode(addr common.Address, code []byte) {
	l.writeLog("Bulk, SetCode, %v, %v", addr, hex.EncodeToString(code))
	l.nested.SetCode(addr, code)
}

func (l *loggingBulkLoad) Close() error {
	l.writeLog("Bulk, Close")
	res := l.nested.Close()
	return res
}
