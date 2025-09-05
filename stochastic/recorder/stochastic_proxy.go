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

package recorder

// TODO: Provide Mocking tests for proxy

import (
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
	db        state.StateDB  // real StateDB object
	snapshots []int          // snapshot stack of currently active snapshots
	registry  *EventRegistry // event registry for deriving statistical parameters
}

// NewStochasticProxy creates a new StateDB proxy for recording events.
func NewStochasticProxy(db state.StateDB, registry *EventRegistry) *StochasticProxy {
	return &StochasticProxy{
		db:        db,
		snapshots: []int{},
		registry:  registry,
	}
}

// CreateAccount creates a new account.
func (p *StochasticProxy) CreateAccount(address common.Address) {
	p.registry.RegisterAddressOp(operations.CreateAccountID, &address)
	p.db.CreateAccount(address)
}

// CreateAccount creates a new contract.
func (p *StochasticProxy) CreateContract(addr common.Address) {
	p.registry.RegisterAddressOp(operations.CreateContractID, &addr)
	p.db.CreateContract(addr)
}

// SubBalance subtracts amount from a contract address.
func (p *StochasticProxy) SubBalance(address common.Address, amount *uint256.Int, reason tracing.BalanceChangeReason) uint256.Int {
	p.registry.RegisterAddressOp(operations.SubBalanceID, &address)
	return p.db.SubBalance(address, amount, reason)
}

// AddBalance adds amount to a contract address.
func (p *StochasticProxy) AddBalance(address common.Address, amount *uint256.Int, reason tracing.BalanceChangeReason) uint256.Int {
	p.registry.RegisterAddressOp(operations.AddBalanceID, &address)
	return p.db.AddBalance(address, amount, reason)
}

// GetBalance retrieves the amount of a contract address.
func (p *StochasticProxy) GetBalance(address common.Address) *uint256.Int {
	p.registry.RegisterAddressOp(operations.GetBalanceID, &address)
	return p.db.GetBalance(address)
}

// GetNonce retrieves the nonce of a contract address.
func (p *StochasticProxy) GetNonce(address common.Address) uint64 {
	p.registry.RegisterAddressOp(operations.GetNonceID, &address)
	return p.db.GetNonce(address)
}

// SetNonce sets the nonce of a contract address.
func (p *StochasticProxy) SetNonce(address common.Address, nonce uint64, reason tracing.NonceChangeReason) {
	p.registry.RegisterAddressOp(operations.SetNonceID, &address)
	p.db.SetNonce(address, nonce, reason)
}

// GetCodeHash returns the hash of the EVM bytecode.
func (p *StochasticProxy) GetCodeHash(address common.Address) common.Hash {
	p.registry.RegisterAddressOp(operations.GetCodeHashID, &address)
	return p.db.GetCodeHash(address)
}

// GetCode returns the EVM bytecode of a contract.
func (p *StochasticProxy) GetCode(address common.Address) []byte {
	p.registry.RegisterAddressOp(operations.GetCodeID, &address)
	return p.db.GetCode(address)
}

// SetCode sets the EVM bytecode of a contract.
func (p *StochasticProxy) SetCode(address common.Address, code []byte) []byte {
	p.registry.RegisterAddressOp(operations.SetCodeID, &address)
	return p.db.SetCode(address, code)
}

// GetCodeSize returns the EVM bytecode's size.
func (p *StochasticProxy) GetCodeSize(address common.Address) int {
	p.registry.RegisterAddressOp(operations.GetCodeSizeID, &address)
	return p.db.GetCodeSize(address)
}

// AddRefund adds gas to the refund counter.
func (p *StochasticProxy) AddRefund(gas uint64) {
	p.db.AddRefund(gas)
}

// SubRefund subtracts gas to the refund counter.
func (p *StochasticProxy) SubRefund(gas uint64) {
	p.db.SubRefund(gas)
}

// GetRefund returns the current value of the refund counter.
func (p *StochasticProxy) GetRefund() uint64 {
	return p.db.GetRefund()
}

// GetCommittedState retrieves a value that is already committed.
func (p *StochasticProxy) GetCommittedState(address common.Address, key common.Hash) common.Hash {
	p.registry.RegisterKeyOp(operations.GetCommittedStateID, &address, &key)
	return p.db.GetCommittedState(address, key)
}

// GetState retrieves a value from the StateDB.
func (p *StochasticProxy) GetState(address common.Address, key common.Hash) common.Hash {
	p.registry.RegisterKeyOp(operations.GetStateID, &address, &key)
	return p.db.GetState(address, key)
}

// GetStorageRoot retrieves the root hash of the storage of an address.
func (p *StochasticProxy) GetStorageRoot(addr common.Address) common.Hash {
	p.registry.RegisterAddressOp(operations.GetStorageRootID, &addr)
	return p.db.GetStorageRoot(addr)
}

// SetState sets a value in the StateDB.
func (p *StochasticProxy) SetState(address common.Address, key common.Hash, value common.Hash) common.Hash {
	p.registry.RegisterValueOp(operations.SetStateID, &address, &key, &value)
	return p.db.SetState(address, key, value)
}

// GetTransientState gets a value from the transient storage.
func (p *StochasticProxy) GetTransientState(addr common.Address, key common.Hash) common.Hash {
	p.registry.RegisterKeyOp(operations.GetTransientStateID, &addr, &key)
	return p.db.GetState(addr, key)
}

// SetTransientState sets a value in the transient storage.
func (p *StochasticProxy) SetTransientState(addr common.Address, key common.Hash, value common.Hash) {
	p.registry.RegisterValueOp(operations.SetTransientStateID, &addr, &key, &value)
	p.db.SetTransientState(addr, key, value)
}

// SelfDestruct destructs an account.
func (p *StochasticProxy) SelfDestruct(address common.Address) uint256.Int {
	p.registry.RegisterAddressOp(operations.SelfDestructID, &address)
	return p.db.SelfDestruct(address)
}

// SelDestruct6780 destructs an account.
func (p *StochasticProxy) SelfDestruct6780(addr common.Address) (uint256.Int, bool) {
	p.registry.RegisterAddressOp(operations.SelfDestruct6780ID, &addr)
	return p.db.SelfDestruct6780(addr)
}

// HasSelfDestructed checks whether a contract has been suicided.
func (p *StochasticProxy) HasSelfDestructed(address common.Address) bool {
	p.registry.RegisterAddressOp(operations.HasSelfDestructedID, &address)
	return p.db.HasSelfDestructed(address)
}

// Exist checks whether the contract exists in the StateDB.
func (p *StochasticProxy) Exist(address common.Address) bool {
	p.registry.RegisterAddressOp(operations.ExistID, &address)
	return p.db.Exist(address)
}

// Empty checks whether the contract is either non-existent
// or empty according to the EIP161 specification (balance = nonce = code = 0).
func (p *StochasticProxy) Empty(address common.Address) bool {
	p.registry.RegisterAddressOp(operations.EmptyID, &address)
	return p.db.Empty(address)
}

// Prepare handles the preparatory steps for executing a state transition.
func (p *StochasticProxy) Prepare(rules params.Rules, sender, coinbase common.Address, dest *common.Address, precompiles []common.Address, txAccesses types.AccessList) {
	p.db.Prepare(rules, sender, coinbase, dest, precompiles, txAccesses)
}

// AddAddressToAccessList adds an address to the access list.
func (p *StochasticProxy) AddAddressToAccessList(address common.Address) {
	p.db.AddAddressToAccessList(address)
}

// AddressInAccessList checks whether an address is in the access list.
func (p *StochasticProxy) AddressInAccessList(address common.Address) bool {
	return p.db.AddressInAccessList(address)
}

// SlotInAccessList checks whether the (address, slot)-tuple is in the access list.
func (p *StochasticProxy) SlotInAccessList(address common.Address, slot common.Hash) (bool, bool) {
	return p.db.SlotInAccessList(address, slot)
}

// AddSlotToAccessList adds the given (address, slot)-tuple to the access list
func (p *StochasticProxy) AddSlotToAccessList(address common.Address, slot common.Hash) {
	p.db.AddSlotToAccessList(address, slot)
}

// RevertToSnapshot reverts all state changes from a given revision.
func (p *StochasticProxy) RevertToSnapshot(snapshot int) {
	p.registry.RegisterOp(operations.RevertToSnapshotID)
	// find snapshot in recordings and record how many snapshots have to be unwound
	for i, recordedSnapshot := range p.snapshots {
		if recordedSnapshot == snapshot {
			p.registry.RegisterSnapshotDelta(len(p.snapshots) - i - 1)
			p.snapshots = p.snapshots[0 : i+1]
			break
		}
	}
	p.db.RevertToSnapshot(snapshot)
}

// Snapshot returns an identifier for the current revision of the state.
func (p *StochasticProxy) Snapshot() int {
	p.registry.RegisterOp(operations.SnapshotID)
	snapshot := p.db.Snapshot()
	p.snapshots = append(p.snapshots, snapshot)
	return snapshot
}

// AddLog adds a log entry.
func (p *StochasticProxy) AddLog(log *types.Log) {
	p.db.AddLog(log)
}

// GetLogs retrieves log entries.
func (p *StochasticProxy) GetLogs(hash common.Hash, block uint64, blockHash common.Hash, blkTimestamp uint64) []*types.Log {
	return p.db.GetLogs(hash, block, blockHash, blkTimestamp)
}

// PointCache returns the point cache used in computations.
func (p *StochasticProxy) PointCache() *utils.PointCache {
	return p.db.PointCache()
}

// Witness retrieves the current state witness.
func (p *StochasticProxy) Witness() *stateless.Witness {
	return p.db.Witness()
}

// AddPreimage adds a SHA3 preimage.
func (p *StochasticProxy) AddPreimage(address common.Hash, image []byte) {
	// call real StateDB
	p.db.AddPreimage(address, image)
}

// AccessEvents retrieves events.
func (p *StochasticProxy) AccessEvents() *geth_state.AccessEvents {
	return p.db.AccessEvents()
}

// SetTxContext sets the current transaction hash and index.
func (p *StochasticProxy) SetTxContext(thash common.Hash, ti int) {
	p.db.SetTxContext(thash, ti)
}

// Finalise the state in StateDB.
func (p *StochasticProxy) Finalise(deleteEmptyObjects bool) {
	p.db.Finalise(deleteEmptyObjects)
}

// IntermediateRoot computes the current hash of the StateDB.
func (p *StochasticProxy) IntermediateRoot(deleteEmptyObjects bool) common.Hash {
	return p.db.IntermediateRoot(deleteEmptyObjects)
}

// Commit StateDB
func (p *StochasticProxy) Commit(block uint64, deleteEmptyObjects bool) (common.Hash, error) {
	return p.db.Commit(block, deleteEmptyObjects)
}

func (p *StochasticProxy) GetHash() (common.Hash, error) {
	return p.db.GetHash()
}

func (p *StochasticProxy) Error() error {
	return p.db.Error()
}

// GetSubstatePostAlloc gets substate post allocation.
func (p *StochasticProxy) GetSubstatePostAlloc() txcontext.WorldState {
	return p.db.GetSubstatePostAlloc()
}

func (p *StochasticProxy) PrepareSubstate(substate txcontext.WorldState, block uint64) {
	p.db.PrepareSubstate(substate, block)
}

func (p *StochasticProxy) BeginTransaction(number uint32) error {
	p.registry.RegisterOp(operations.BeginTransactionID)
	if err := p.db.BeginTransaction(number); err != nil {
		return err
	}
	// clear all snapshots
	p.snapshots = []int{}
	return nil
}

func (p *StochasticProxy) EndTransaction() error {
	p.registry.RegisterOp(operations.EndTransactionID)
	if err := p.db.EndTransaction(); err != nil {
		return err
	}
	// clear all snapshots
	p.snapshots = []int{}
	return nil
}

func (p *StochasticProxy) BeginBlock(number uint64) error {
	p.registry.RegisterOp(operations.BeginBlockID)
	return p.db.BeginBlock(number)
}

func (p *StochasticProxy) EndBlock() error {
	p.registry.RegisterOp(operations.EndBlockID)
	return p.db.EndBlock()
}

func (p *StochasticProxy) BeginSyncPeriod(number uint64) {
	p.registry.RegisterOp(operations.BeginSyncPeriodID)
	p.db.BeginSyncPeriod(number)
}

func (p *StochasticProxy) EndSyncPeriod() {
	p.registry.RegisterOp(operations.EndSyncPeriodID)
	p.db.EndSyncPeriod()
}

func (p *StochasticProxy) Close() error {
	return p.db.Close()
}

func (p *StochasticProxy) StartBulkLoad(uint64) (state.BulkLoad, error) {
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
