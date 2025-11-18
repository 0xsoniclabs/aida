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

package recorder

import (
	"math"

	"github.com/0xsoniclabs/aida/state"
	"github.com/0xsoniclabs/aida/stochastic/operations"
	"github.com/0xsoniclabs/aida/txcontext"

	"github.com/ethereum/go-ethereum/common"
	geth_state "github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/stateless"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/trie/utils"
	"github.com/holiman/uint256"
)

// StochasticProxy data structure for counting StateDB operations and their arguments.
type StochasticProxy struct {
	db        state.StateDB // StateDB object
	snapshots []int         // snapshot stack of currently active snapshots
	stats     *Stats        // stats for storing state of Markov process
}

func uint256ToInt64(amount *uint256.Int) int64 {
	if amount == nil {
		return 0
	}
	if amount.BitLen() > 63 {
		return math.MaxInt64
	}
	return int64(amount.Uint64())
}

// NewStochasticProxy creates a new StateDB proxy for recording markov stats
func NewStochasticProxy(db state.StateDB, stats *Stats) *StochasticProxy {
	return &StochasticProxy{
		db:        db,
		snapshots: []int{},
		stats:     stats,
	}
}

func (p *StochasticProxy) CreateAccount(address common.Address) {
	err := p.stats.CountAddressOp(operations.CreateAccountID, &address)
	if err != nil {
		panic(err)
	}
	p.db.CreateAccount(address)
}

func (p *StochasticProxy) CreateContract(addr common.Address) {
	err := p.stats.CountAddressOp(operations.CreateContractID, &addr)
	if err != nil {
		panic(err)
	}
	p.db.CreateContract(addr)
}

func (p *StochasticProxy) SubBalance(address common.Address, amount *uint256.Int, reason tracing.BalanceChangeReason) uint256.Int {
	err := p.stats.CountAddressOp(operations.SubBalanceID, &address)
	if err != nil {
		panic(err)
	}
	p.stats.RecordBalance(uint256ToInt64(amount))
	return p.db.SubBalance(address, amount, reason)
}

func (p *StochasticProxy) AddBalance(address common.Address, amount *uint256.Int, reason tracing.BalanceChangeReason) uint256.Int {
	err := p.stats.CountAddressOp(operations.AddBalanceID, &address)
	if err != nil {
		panic(err)
	}
	p.stats.RecordBalance(uint256ToInt64(amount))
	return p.db.AddBalance(address, amount, reason)
}

func (p *StochasticProxy) GetBalance(address common.Address) *uint256.Int {
	err := p.stats.CountAddressOp(operations.GetBalanceID, &address)
	if err != nil {
		panic(err)
	}
	return p.db.GetBalance(address)
}

func (p *StochasticProxy) GetNonce(address common.Address) uint64 {
	err := p.stats.CountAddressOp(operations.GetNonceID, &address)
	if err != nil {
		panic(err)
	}
	return p.db.GetNonce(address)
}

func (p *StochasticProxy) SetNonce(address common.Address, nonce uint64, reason tracing.NonceChangeReason) {
	err := p.stats.CountAddressOp(operations.SetNonceID, &address)
	if err != nil {
		panic(err)
	}
	p.stats.RecordNonce(nonce)
	p.db.SetNonce(address, nonce, reason)
}

func (p *StochasticProxy) GetCodeHash(address common.Address) common.Hash {
	err := p.stats.CountAddressOp(operations.GetCodeHashID, &address)
	if err != nil {
		panic(err)
	}
	return p.db.GetCodeHash(address)
}

func (p *StochasticProxy) GetCode(address common.Address) []byte {
	err := p.stats.CountAddressOp(operations.GetCodeID, &address)
	if err != nil {
		panic(err)
	}
	return p.db.GetCode(address)
}

func (p *StochasticProxy) SetCode(address common.Address, code []byte, reason tracing.CodeChangeReason) []byte {
	err := p.stats.CountAddressOp(operations.SetCodeID, &address)
	if err != nil {
		panic(err)
	}
	p.stats.RecordCodeSize(len(code))
	return p.db.SetCode(address, code, reason)
}

func (p *StochasticProxy) GetCodeSize(address common.Address) int {
	err := p.stats.CountAddressOp(operations.GetCodeSizeID, &address)
	if err != nil {
		panic(err)
	}
	return p.db.GetCodeSize(address)
}

func (p *StochasticProxy) AddRefund(gas uint64) {
	p.db.AddRefund(gas)
}

func (p *StochasticProxy) SubRefund(gas uint64) {
	p.db.SubRefund(gas)
}

func (p *StochasticProxy) GetRefund() uint64 {
	return p.db.GetRefund()
}

func (p *StochasticProxy) GetCommittedState(address common.Address, key common.Hash) common.Hash {
	err := p.stats.CountKeyOp(operations.GetCommittedStateID, &address, &key)
	if err != nil {
		panic(err)
	}
	return p.db.GetCommittedState(address, key)
}

func (p *StochasticProxy) GetStateAndCommittedState(address common.Address, hash common.Hash) (common.Hash, common.Hash) {
	err := p.stats.CountKeyOp(operations.GetStateAndCommittedStateID, &address, &hash)
	if err != nil {
		panic(err)
	}
	return p.db.GetStateAndCommittedState(address, hash)
}

func (p *StochasticProxy) GetState(address common.Address, key common.Hash) common.Hash {
	err := p.stats.CountKeyOp(operations.GetStateID, &address, &key)
	if err != nil {
		panic(err)
	}
	return p.db.GetState(address, key)
}

func (p *StochasticProxy) GetStorageRoot(addr common.Address) common.Hash {
	err := p.stats.CountAddressOp(operations.GetStorageRootID, &addr)
	if err != nil {
		panic(err)
	}
	return p.db.GetStorageRoot(addr)
}

func (p *StochasticProxy) SetState(address common.Address, key common.Hash, value common.Hash) common.Hash {
	err := p.stats.CountValueOp(operations.SetStateID, &address, &key, &value)
	if err != nil {
		panic(err)
	}
	return p.db.SetState(address, key, value)
}

func (p *StochasticProxy) GetTransientState(addr common.Address, key common.Hash) common.Hash {
	err := p.stats.CountKeyOp(operations.GetTransientStateID, &addr, &key)
	if err != nil {
		panic(err)
	}
	return p.db.GetTransientState(addr, key)
}

func (p *StochasticProxy) SetTransientState(addr common.Address, key common.Hash, value common.Hash) {
	err := p.stats.CountValueOp(operations.SetTransientStateID, &addr, &key, &value)
	if err != nil {
		panic(err)
	}
	p.db.SetTransientState(addr, key, value)
}

func (p *StochasticProxy) SelfDestruct(address common.Address) uint256.Int {
	err := p.stats.CountAddressOp(operations.SelfDestructID, &address)
	if err != nil {
		panic(err)
	}
	return p.db.SelfDestruct(address)
}

func (p *StochasticProxy) SelfDestruct6780(addr common.Address) (uint256.Int, bool) {
	err := p.stats.CountAddressOp(operations.SelfDestruct6780ID, &addr)
	if err != nil {
		panic(err)
	}
	return p.db.SelfDestruct6780(addr)
}

func (p *StochasticProxy) HasSelfDestructed(address common.Address) bool {
	err := p.stats.CountAddressOp(operations.HasSelfDestructedID, &address)
	if err != nil {
		panic(err)
	}
	return p.db.HasSelfDestructed(address)
}

func (p *StochasticProxy) Exist(address common.Address) bool {
	err := p.stats.CountAddressOp(operations.ExistID, &address)
	if err != nil {
		panic(err)
	}
	return p.db.Exist(address)
}

func (p *StochasticProxy) Empty(address common.Address) bool {
	err := p.stats.CountAddressOp(operations.EmptyID, &address)
	if err != nil {
		panic(err)
	}
	return p.db.Empty(address)
}

func (p *StochasticProxy) Prepare(rules params.Rules, sender, coinbase common.Address, dest *common.Address, precompiles []common.Address, txAccesses types.AccessList) {
	p.db.Prepare(rules, sender, coinbase, dest, precompiles, txAccesses)
}

func (p *StochasticProxy) AddAddressToAccessList(address common.Address) {
	p.db.AddAddressToAccessList(address)
}

func (p *StochasticProxy) AddressInAccessList(address common.Address) bool {
	return p.db.AddressInAccessList(address)
}

func (p *StochasticProxy) SlotInAccessList(address common.Address, slot common.Hash) (bool, bool) {
	return p.db.SlotInAccessList(address, slot)
}

func (p *StochasticProxy) AddSlotToAccessList(address common.Address, slot common.Hash) {
	p.db.AddSlotToAccessList(address, slot)
}

func (p *StochasticProxy) RevertToSnapshot(snapshot int) {
	for i, recordedSnapshot := range p.snapshots {
		if recordedSnapshot == snapshot {
			err := p.stats.CountSnapshot(len(p.snapshots) - i - 1)
			if err != nil {
				panic(err)
			}
			p.snapshots = p.snapshots[0:i]
			break
		}
	}
	p.db.RevertToSnapshot(snapshot)
}

func (p *StochasticProxy) Snapshot() int {
	err := p.stats.CountOp(operations.SnapshotID)
	if err != nil {
		panic(err)
	}
	snapshot := p.db.Snapshot()
	p.snapshots = append(p.snapshots, snapshot)
	return snapshot
}

func (p *StochasticProxy) AddLog(log *types.Log) {
	p.db.AddLog(log)
}

func (p *StochasticProxy) GetLogs(hash common.Hash, block uint64, blockHash common.Hash, blkTimestamp uint64) []*types.Log {
	return p.db.GetLogs(hash, block, blockHash, blkTimestamp)
}

func (p *StochasticProxy) PointCache() *utils.PointCache {
	return p.db.PointCache()
}

func (p *StochasticProxy) Witness() *stateless.Witness {
	return p.db.Witness()
}

func (p *StochasticProxy) AddPreimage(address common.Hash, image []byte) {
	// call real StateDB
	p.db.AddPreimage(address, image)
}

func (p *StochasticProxy) AccessEvents() *geth_state.AccessEvents {
	return p.db.AccessEvents()
}

func (p *StochasticProxy) SetTxContext(thash common.Hash, ti int) {
	p.db.SetTxContext(thash, ti)
}

func (p *StochasticProxy) Finalise(deleteEmptyObjects bool) {
	p.db.Finalise(deleteEmptyObjects)
}

func (p *StochasticProxy) IntermediateRoot(deleteEmptyObjects bool) common.Hash {
	return p.db.IntermediateRoot(deleteEmptyObjects)
}

func (p *StochasticProxy) Commit(block uint64, deleteEmptyObjects bool) (common.Hash, error) {
	return p.db.Commit(block, deleteEmptyObjects)
}

func (p *StochasticProxy) GetHash() (common.Hash, error) {
	return p.db.GetHash()
}

func (p *StochasticProxy) Error() error {
	return p.db.Error()
}

func (p *StochasticProxy) GetSubstatePostAlloc() txcontext.WorldState {
	return p.db.GetSubstatePostAlloc()
}

func (p *StochasticProxy) PrepareSubstate(substate txcontext.WorldState, block uint64) {
	p.db.PrepareSubstate(substate, block)
}

func (p *StochasticProxy) BeginTransaction(number uint32) error {
	if err := p.stats.CountOp(operations.BeginTransactionID); err != nil {
		return err
	}
	if err := p.db.BeginTransaction(number); err != nil {
		return err
	}
	p.snapshots = []int{}
	return nil
}

func (p *StochasticProxy) EndTransaction() error {
	if err := p.stats.CountOp(operations.EndTransactionID); err != nil {
		return err
	}
	if err := p.db.EndTransaction(); err != nil {
		return err
	}
	p.snapshots = []int{}
	return nil
}

func (p *StochasticProxy) BeginBlock(number uint64) error {
	if err := p.stats.CountOp(operations.BeginBlockID); err != nil {
		return err
	}
	return p.db.BeginBlock(number)
}

func (p *StochasticProxy) EndBlock() error {
	if err := p.stats.CountOp(operations.EndBlockID); err != nil {
		return err
	}
	return p.db.EndBlock()
}

func (p *StochasticProxy) BeginSyncPeriod(number uint64) {
	if err := p.stats.CountOp(operations.BeginSyncPeriodID); err != nil {
		panic(err)
	}
	p.db.BeginSyncPeriod(number)
}

func (p *StochasticProxy) EndSyncPeriod() {
	if err := p.stats.CountOp(operations.EndSyncPeriodID); err != nil {
		panic(err)
	}
	p.db.EndSyncPeriod()
}

func (p *StochasticProxy) Close() error {
	return p.db.Close()
}

func (p *StochasticProxy) StartBulkLoad(uint64) (state.BulkLoad, error) {
	// This code should never be reachable under any circumstances
	panic("StartBulkLoad not supported by EventProxy")
}

func (p *StochasticProxy) GetMemoryUsage() *state.MemoryUsage {
	return p.db.GetMemoryUsage()
}

func (p *StochasticProxy) GetArchiveState(block uint64) (state.NonCommittableStateDB, error) {
	return p.db.GetArchiveState(block)
}

func (p *StochasticProxy) GetArchiveBlockHeight() (uint64, bool, error) {
	return p.db.GetArchiveBlockHeight()
}

func (p *StochasticProxy) GetShadowDB() state.StateDB {
	return p.db.GetShadowDB()
}
