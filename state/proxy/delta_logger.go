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
	"bufio"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
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

// DeltaLogSink writes textual operations to disk in the format expected by the delta debugger.
type DeltaLogSink struct {
	mu     sync.Mutex
	writer *bufio.Writer
	closer io.Closer
	log    logger.Logger
}

// NewDeltaLogSink creates a sink that logs to the provided writer and logger.
func NewDeltaLogSink(log logger.Logger, writer *bufio.Writer, closer io.Closer) *DeltaLogSink {
	return &DeltaLogSink{
		writer: writer,
		closer: closer,
		log:    log,
	}
}

// Logf writes the formatted message to the sink and flushes immediately so the last operation
// is present even if the process crashes.
func (s *DeltaLogSink) Logf(format string, args ...any) {
	if s == nil {
		return
	}

	line := fmt.Sprintf(format, args...)
	line = strings.TrimSuffix(line, "\n")

	s.mu.Lock()
	if s.writer != nil {
		if _, err := s.writer.WriteString(line + "\n"); err != nil && s.log != nil {
			s.log.Errorf("delta logger: write failed: %v", err)
		}
		if err := s.writer.Flush(); err != nil && s.log != nil {
			s.log.Errorf("delta logger: flush failed: %v", err)
		}
	}
	s.mu.Unlock()

	if s.log != nil {
		s.log.Debug(line)
	}
}

// Flush flushes buffered data and fsyncs when supported by the closer.
func (s *DeltaLogSink) Flush() error {
	if s == nil {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.writer != nil {
		if err := s.writer.Flush(); err != nil {
			return err
		}
	}

	if syncer, ok := s.closer.(interface{ Sync() error }); ok {
		return syncer.Sync()
	}
	return nil
}

// Close flushes and closes the underlying writer/closer.
func (s *DeltaLogSink) Close() error {
	if s == nil {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	var err error
	if s.writer != nil {
		err = errors.Join(err, s.writer.Flush())
	}
	if syncer, ok := s.closer.(interface{ Sync() error }); ok {
		err = errors.Join(err, syncer.Sync())
	}
	if s.closer != nil {
		err = errors.Join(err, s.closer.Close())
	}

	s.writer = nil
	s.closer = nil
	return err
}

// DeltaLoggingStateDB logs operations in the delta debugger format before executing them.
type DeltaLoggingStateDB struct {
	deltaLoggingVmStateDb
	state state.StateDB
	sink  *DeltaLogSink
}

// NewDeltaLoggerProxy wraps the given StateDB with the delta logger.
func NewDeltaLoggerProxy(db state.StateDB, sink *DeltaLogSink) state.StateDB {
	if sink == nil {
		return db
	}

	return &DeltaLoggingStateDB{
		deltaLoggingVmStateDb: deltaLoggingVmStateDb{
			db:   db,
			sink: sink,
		},
		state: db,
		sink:  sink,
	}
}

func (s *DeltaLoggingStateDB) Error() error {
	s.logf("Error")
	return s.state.Error()
}

func (s *DeltaLoggingStateDB) BeginBlock(blk uint64) error {
	s.logf("BeginBlock, %d", blk)
	return s.state.BeginBlock(blk)
}

func (s *DeltaLoggingStateDB) EndBlock() error {
	s.logf("EndBlock")
	return s.state.EndBlock()
}

func (s *DeltaLoggingStateDB) BeginSyncPeriod(number uint64) {
	s.logf("BeginSyncPeriod, %d", number)
	s.state.BeginSyncPeriod(number)
}

func (s *DeltaLoggingStateDB) EndSyncPeriod() {
	s.logf("EndSyncPeriod")
	s.state.EndSyncPeriod()
}

func (s *DeltaLoggingStateDB) GetHash() (common.Hash, error) {
	s.logf("GetHash")
	return s.state.GetHash()
}

func (s *DeltaLoggingStateDB) Close() error {
	s.logf("Close")
	if err := s.state.Close(); err != nil {
		return err
	}
	return s.sink.Flush()
}

// We intentionally do not wrap bulk-load operations because the delta replayer does not
// understand them. This keeps generated traces compatible with the delta debugger.
func (s *DeltaLoggingStateDB) StartBulkLoad(block uint64) (state.BulkLoad, error) {
	return s.state.StartBulkLoad(block)
}

func (s *DeltaLoggingStateDB) GetArchiveState(block uint64) (state.NonCommittableStateDB, error) {
	archive, err := s.state.GetArchiveState(block)
	if err != nil {
		return nil, err
	}
	return &deltaLoggingNonCommittableStateDB{
		deltaLoggingVmStateDb: deltaLoggingVmStateDb{
			db:   archive,
			sink: s.sink,
		},
		nonCommittableStateDB: archive,
	}, nil
}

func (s *DeltaLoggingStateDB) GetArchiveBlockHeight() (uint64, bool, error) {
	s.logf("GetArchiveBlockHeight")
	return s.state.GetArchiveBlockHeight()
}

func (s *DeltaLoggingStateDB) GetMemoryUsage() *state.MemoryUsage {
	return s.state.GetMemoryUsage()
}

func (s *DeltaLoggingStateDB) GetShadowDB() state.StateDB {
	return s.state.GetShadowDB()
}

func (s *DeltaLoggingStateDB) Finalise(deleteEmptyObjects bool) {
	s.logf("Finalise, %s", formatBool(deleteEmptyObjects))
	s.state.Finalise(deleteEmptyObjects)
}

func (s *DeltaLoggingStateDB) IntermediateRoot(deleteEmptyObjects bool) common.Hash {
	s.logf("IntermediateRoot, %s", formatBool(deleteEmptyObjects))
	return s.state.IntermediateRoot(deleteEmptyObjects)
}

func (s *DeltaLoggingStateDB) Commit(block uint64, deleteEmptyObjects bool) (common.Hash, error) {
	s.logf("Commit, %s", formatBool(deleteEmptyObjects))
	return s.state.Commit(block, deleteEmptyObjects)
}

func (s *DeltaLoggingStateDB) PrepareSubstate(substate txcontext.WorldState, block uint64) {
	s.logf("PrepareSubstate")
	s.state.PrepareSubstate(substate, block)
}

type deltaLoggingVmStateDb struct {
	db   state.VmStateDB
	sink *DeltaLogSink
}

func (s *deltaLoggingVmStateDb) logf(format string, args ...any) {
	if s.sink != nil {
		s.sink.Logf(format, args...)
	}
}

func (s *deltaLoggingVmStateDb) CreateAccount(addr common.Address) {
	s.logf("CreateAccount, %s", addr.Hex())
	s.db.CreateAccount(addr)
}

func (s *deltaLoggingVmStateDb) CreateContract(addr common.Address) {
	s.logf("CreateContract, %s", addr.Hex())
	s.db.CreateContract(addr)
}

func (s *deltaLoggingVmStateDb) Exist(addr common.Address) bool {
	s.logf("Exist, %s", addr.Hex())
	return s.db.Exist(addr)
}

func (s *deltaLoggingVmStateDb) Empty(addr common.Address) bool {
	s.logf("Empty, %s", addr.Hex())
	return s.db.Empty(addr)
}

func (s *deltaLoggingVmStateDb) SelfDestruct(addr common.Address) uint256.Int {
	s.logf("SelfDestruct, %s", addr.Hex())
	return s.db.SelfDestruct(addr)
}

func (s *deltaLoggingVmStateDb) SelfDestruct6780(addr common.Address) (uint256.Int, bool) {
	s.logf("SelfDestruct6780, %s", addr.Hex())
	return s.db.SelfDestruct6780(addr)
}

func (s *deltaLoggingVmStateDb) HasSelfDestructed(addr common.Address) bool {
	s.logf("HasSelfDestructed, %s", addr.Hex())
	return s.db.HasSelfDestructed(addr)
}

func (s *deltaLoggingVmStateDb) GetBalance(addr common.Address) *uint256.Int {
	s.logf("GetBalance, %s", addr.Hex())
	return s.db.GetBalance(addr)
}

func (s *deltaLoggingVmStateDb) AddBalance(addr common.Address, value *uint256.Int, reason tracing.BalanceChangeReason) uint256.Int {
	amount := formatUint256(value)
	s.logf("AddBalance, %s, %s, 0, %s, %s", addr.Hex(), amount, reason.String(), amount)
	return s.db.AddBalance(addr, value, reason)
}

func (s *deltaLoggingVmStateDb) SubBalance(addr common.Address, value *uint256.Int, reason tracing.BalanceChangeReason) uint256.Int {
	amount := formatUint256(value)
	s.logf("SubBalance, %s, %s, 0, %s, %s", addr.Hex(), amount, reason.String(), amount)
	return s.db.SubBalance(addr, value, reason)
}

func (s *deltaLoggingVmStateDb) GetNonce(addr common.Address) uint64 {
	s.logf("GetNonce, %s", addr.Hex())
	return s.db.GetNonce(addr)
}

func (s *deltaLoggingVmStateDb) SetNonce(addr common.Address, nonce uint64, reason tracing.NonceChangeReason) {
	s.logf("SetNonce, %s, %s, %s", addr.Hex(), formatUint64(nonce), reason.String())
	s.db.SetNonce(addr, nonce, reason)
}

func (s *deltaLoggingVmStateDb) GetCommittedState(addr common.Address, key common.Hash) common.Hash {
	s.logf("GetCommittedState, %s, %s", addr.Hex(), key.Hex())
	return s.db.GetCommittedState(addr, key)
}

func (s *deltaLoggingVmStateDb) GetStateAndCommittedState(addr common.Address, key common.Hash) (common.Hash, common.Hash) {
	s.logf("GetStateAndCommittedState, %s, %s", addr.Hex(), key.Hex())
	return s.db.GetStateAndCommittedState(addr, key)
}

func (s *deltaLoggingVmStateDb) GetState(addr common.Address, key common.Hash) common.Hash {
	s.logf("GetState, %s, %s", addr.Hex(), key.Hex())
	return s.db.GetState(addr, key)
}

func (s *deltaLoggingVmStateDb) SetState(addr common.Address, key common.Hash, value common.Hash) common.Hash {
	s.logf("SetState, %s, %s, %s", addr.Hex(), key.Hex(), value.Hex())
	return s.db.SetState(addr, key, value)
}

func (s *deltaLoggingVmStateDb) SetTransientState(addr common.Address, key common.Hash, value common.Hash) {
	s.logf("SetTransientState, %s, %s, %s", addr.Hex(), key.Hex(), value.Hex())
	s.db.SetTransientState(addr, key, value)
}

func (s *deltaLoggingVmStateDb) GetTransientState(addr common.Address, key common.Hash) common.Hash {
	s.logf("GetTransientState, %s, %s", addr.Hex(), key.Hex())
	return s.db.GetTransientState(addr, key)
}

func (s *deltaLoggingVmStateDb) GetCodeHash(addr common.Address) common.Hash {
	s.logf("GetCodeHash, %s", addr.Hex())
	return s.db.GetCodeHash(addr)
}

func (s *deltaLoggingVmStateDb) GetCode(addr common.Address) []byte {
	s.logf("GetCode, %s", addr.Hex())
	return s.db.GetCode(addr)
}

func (s *deltaLoggingVmStateDb) SetCode(addr common.Address, code []byte, reason tracing.CodeChangeReason) []byte {
	s.logf("SetCode, %s, %s", addr.Hex(), formatBytes(code))
	return s.db.SetCode(addr, code, reason)
}

func (s *deltaLoggingVmStateDb) GetCodeSize(addr common.Address) int {
	s.logf("GetCodeSize, %s", addr.Hex())
	return s.db.GetCodeSize(addr)
}

func (s *deltaLoggingVmStateDb) Snapshot() int {
	s.logf("Snapshot")
	return s.db.Snapshot()
}

func (s *deltaLoggingVmStateDb) RevertToSnapshot(id int) {
	s.logf("RevertToSnapshot, %d", id)
	s.db.RevertToSnapshot(id)
}

func (s *deltaLoggingVmStateDb) BeginTransaction(tx uint32) error {
	s.logf("BeginTransaction, %s", strconv.FormatUint(uint64(tx), 10))
	return s.db.BeginTransaction(tx)
}

func (s *deltaLoggingVmStateDb) EndTransaction() error {
	s.logf("EndTransaction")
	return s.db.EndTransaction()
}

func (s *deltaLoggingVmStateDb) Finalise(deleteEmptyObjects bool) {
	s.logf("Finalise, %s", formatBool(deleteEmptyObjects))
	s.db.Finalise(deleteEmptyObjects)
}

func (s *deltaLoggingVmStateDb) AddRefund(amount uint64) {
	s.logf("AddRefund, %s", formatUint64(amount))
	s.db.AddRefund(amount)
}

func (s *deltaLoggingVmStateDb) SubRefund(amount uint64) {
	s.logf("SubRefund, %s", formatUint64(amount))
	s.db.SubRefund(amount)
}

func (s *deltaLoggingVmStateDb) GetRefund() uint64 {
	s.logf("GetRefund")
	return s.db.GetRefund()
}

func (s *deltaLoggingVmStateDb) Prepare(rules params.Rules, sender, coinbase common.Address, dest *common.Address, precompiles []common.Address, txAccesses types.AccessList) {
	s.logf("Prepare")
	s.db.Prepare(rules, sender, coinbase, dest, precompiles, txAccesses)
}

func (s *deltaLoggingVmStateDb) AddressInAccessList(addr common.Address) bool {
	s.logf("AddressInAccessList, %s", addr.Hex())
	return s.db.AddressInAccessList(addr)
}

func (s *deltaLoggingVmStateDb) SlotInAccessList(addr common.Address, slot common.Hash) (addressOk bool, slotOk bool) {
	s.logf("SlotInAccessList, %s, %s", addr.Hex(), slot.Hex())
	return s.db.SlotInAccessList(addr, slot)
}

func (s *deltaLoggingVmStateDb) AddAddressToAccessList(addr common.Address) {
	s.logf("AddAddressToAccessList, %s", addr.Hex())
	s.db.AddAddressToAccessList(addr)
}

func (s *deltaLoggingVmStateDb) AddSlotToAccessList(addr common.Address, slot common.Hash) {
	s.logf("AddSlotToAccessList, %s, %s", addr.Hex(), slot.Hex())
	s.db.AddSlotToAccessList(addr, slot)
}

func (s *deltaLoggingVmStateDb) AddLog(entry *types.Log) {
	s.logf("AddLog")
	s.db.AddLog(entry)
}

func (s *deltaLoggingVmStateDb) GetLogs(hash common.Hash, block uint64, blockHash common.Hash, blkTimestamp uint64) []*types.Log {
	s.logf("GetLogs, %s, %s, %s, %s", hash.Hex(), formatUint64(block), blockHash.Hex(), formatUint64(blkTimestamp))
	return s.db.GetLogs(hash, block, blockHash, blkTimestamp)
}

func (s *deltaLoggingVmStateDb) PointCache() *utils.PointCache {
	s.logf("PointCache")
	return s.db.PointCache()
}

func (s *deltaLoggingVmStateDb) Witness() *stateless.Witness {
	s.logf("Witness")
	return s.db.Witness()
}

func (s *deltaLoggingVmStateDb) SetTxContext(thash common.Hash, ti int) {
	s.logf("SetTxContext, %s, %d", thash.Hex(), ti)
	s.db.SetTxContext(thash, ti)
}

func (s *deltaLoggingVmStateDb) GetSubstatePostAlloc() txcontext.WorldState {
	s.logf("GetSubstatePostAlloc")
	return s.db.GetSubstatePostAlloc()
}

func (s *deltaLoggingVmStateDb) AddPreimage(hash common.Hash, data []byte) {
	s.logf("AddPreimage, %s, %s", hash.Hex(), formatBytes(data))
	s.db.AddPreimage(hash, data)
}

func (s *deltaLoggingVmStateDb) AccessEvents() *geth.AccessEvents {
	s.logf("AccessEvents")
	return s.db.AccessEvents()
}

func (s *deltaLoggingVmStateDb) GetStorageRoot(addr common.Address) common.Hash {
	s.logf("GetStorageRoot, %s", addr.Hex())
	return s.db.GetStorageRoot(addr)
}

type deltaLoggingNonCommittableStateDB struct {
	deltaLoggingVmStateDb
	nonCommittableStateDB state.NonCommittableStateDB
}

func (s *deltaLoggingNonCommittableStateDB) GetHash() (common.Hash, error) {
	s.logf("GetHash")
	return s.nonCommittableStateDB.GetHash()
}

func (s *deltaLoggingNonCommittableStateDB) Release() error {
	s.logf("Release")
	return s.nonCommittableStateDB.Release()
}

func formatUint256(value *uint256.Int) string {
	if value == nil {
		return "0"
	}
	return value.String()
}

func formatBool(value bool) string {
	return strconv.FormatBool(value)
}

func formatUint64(value uint64) string {
	return strconv.FormatUint(value, 10)
}

func formatBytes(data []byte) string {
	return "0x" + hex.EncodeToString(data)
}
