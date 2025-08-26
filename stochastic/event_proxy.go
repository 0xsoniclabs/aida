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

package stochastic

import (
	"github.com/0xsoniclabs/aida/state"
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

// EventProxy data structure for capturing StateDB events
type EventProxy struct {
	db        state.StateDB  // real StateDB object
	snapshots []int          // snapshot stack of currently active snapshots
	registry  *EventRegistry // event registry for deriving statistical parameters
}

// NewEventProxy creates a new StateDB proxy for recording events.
func NewEventProxy(db state.StateDB, registry *EventRegistry) *EventProxy {
	return &EventProxy{
		db:        db,
		snapshots: []int{},
		registry:  registry,
	}
}

// CreateAccount creates a new account.
func (p *EventProxy) CreateAccount(address common.Address) {
	p.registry.RegisterAddressOp(CreateAccountID, &address)
	p.db.CreateAccount(address)
}

// CreateAccount creates a new contract.
func (p *EventProxy) CreateContract(addr common.Address) {
	p.registry.RegisterAddressOp(CreateContractID, &addr)
	p.db.CreateContract(addr)
}

// SubBalance subtracts amount from a contract address.
func (p *EventProxy) SubBalance(address common.Address, amount *uint256.Int, reason tracing.BalanceChangeReason) uint256.Int {
	p.registry.RegisterAddressOp(SubBalanceID, &address)
	return p.db.SubBalance(address, amount, reason)
}

// AddBalance adds amount to a contract address.
func (p *EventProxy) AddBalance(address common.Address, amount *uint256.Int, reason tracing.BalanceChangeReason) uint256.Int {
	p.registry.RegisterAddressOp(AddBalanceID, &address)
	return p.db.AddBalance(address, amount, reason)
}

// GetBalance retrieves the amount of a contract address.
func (p *EventProxy) GetBalance(address common.Address) *uint256.Int {
	p.registry.RegisterAddressOp(GetBalanceID, &address)
	return p.db.GetBalance(address)
}

// GetNonce retrieves the nonce of a contract address.
func (p *EventProxy) GetNonce(address common.Address) uint64 {
	p.registry.RegisterAddressOp(GetNonceID, &address)
	return p.db.GetNonce(address)
}

// SetNonce sets the nonce of a contract address.
func (p *EventProxy) SetNonce(address common.Address, nonce uint64, reason tracing.NonceChangeReason) {
	p.registry.RegisterAddressOp(SetNonceID, &address)
	p.db.SetNonce(address, nonce, reason)
}

// GetCodeHash returns the hash of the EVM bytecode.
func (p *EventProxy) GetCodeHash(address common.Address) common.Hash {
	p.registry.RegisterAddressOp(GetCodeHashID, &address)
	return p.db.GetCodeHash(address)
}

// GetCode returns the EVM bytecode of a contract.
func (p *EventProxy) GetCode(address common.Address) []byte {
	p.registry.RegisterAddressOp(GetCodeID, &address)
	return p.db.GetCode(address)
}

// SetCode sets the EVM bytecode of a contract.
func (p *EventProxy) SetCode(address common.Address, code []byte) []byte {
	p.registry.RegisterAddressOp(SetCodeID, &address)
	return p.db.SetCode(address, code)
}

// GetCodeSize returns the EVM bytecode's size.
func (p *EventProxy) GetCodeSize(address common.Address) int {
	p.registry.RegisterAddressOp(GetCodeSizeID, &address)
	return p.db.GetCodeSize(address)
}

// AddRefund adds gas to the refund counter.
func (p *EventProxy) AddRefund(gas uint64) {
	p.db.AddRefund(gas)
}

// SubRefund subtracts gas to the refund counter.
func (p *EventProxy) SubRefund(gas uint64) {
	p.db.SubRefund(gas)
}

// GetRefund returns the current value of the refund counter.
func (p *EventProxy) GetRefund() uint64 {
	return p.db.GetRefund()
}

// GetCommittedState retrieves a value that is already committed.
func (p *EventProxy) GetCommittedState(address common.Address, key common.Hash) common.Hash {
	p.registry.RegisterKeyOp(GetCommittedStateID, &address, &key)
	return p.db.GetCommittedState(address, key)
}

// GetState retrieves a value from the StateDB.
func (p *EventProxy) GetState(address common.Address, key common.Hash) common.Hash {
	p.registry.RegisterKeyOp(GetStateID, &address, &key)
	return p.db.GetState(address, key)
}

// GetStorageRoot retrieves the root hash of the storage of an address.
func (p *EventProxy) GetStorageRoot(addr common.Address) common.Hash {
	p.registry.RegisterAddressOp(GetStorageRootID, &addr)
	return p.db.GetStorageRoot(addr)
}

// SetState sets a value in the StateDB.
func (p *EventProxy) SetState(address common.Address, key common.Hash, value common.Hash) common.Hash {
	p.registry.RegisterValueOp(SetStateID, &address, &key, &value)
	return p.db.SetState(address, key, value)
}

// GetTransientState gets a value from the transient storage.
func (p *EventProxy) GetTransientState(addr common.Address, key common.Hash) common.Hash {
	p.registry.RegisterKeyOp(GetTransientStateID, &addr, &key)
	return p.db.GetState(addr, key)
}

// SetTransientState sets a value in the transient storage.
func (p *EventProxy) SetTransientState(addr common.Address, key common.Hash, value common.Hash) {
	p.registry.RegisterValueOp(SetTransientStateID, &addr, &key, &value)
	p.db.SetTransientState(addr, key, value)
}

// SelfDestruct destructs an account.
func (p *EventProxy) SelfDestruct(address common.Address) uint256.Int {
	p.registry.RegisterAddressOp(SelfDestructID, &address)
	return p.db.SelfDestruct(address)
}

// SelDestruct6780 destructs an account.
func (p *EventProxy) SelfDestruct6780(addr common.Address) (uint256.Int, bool) {
	p.registry.RegisterAddressOp(SelfDestruct6780ID, &addr)
	return p.db.SelfDestruct6780(addr)
}

// HasSelfDestructed checks whether a contract has been suicided.
func (p *EventProxy) HasSelfDestructed(address common.Address) bool {
	p.registry.RegisterAddressOp(HasSelfDestructedID, &address)
	return p.db.HasSelfDestructed(address)
}

// Exist checks whether the contract exists in the StateDB.
func (p *EventProxy) Exist(address common.Address) bool {
	p.registry.RegisterAddressOp(ExistID, &address)
	return p.db.Exist(address)
}

// Empty checks whether the contract is either non-existent
// or empty according to the EIP161 specification (balance = nonce = code = 0).
func (p *EventProxy) Empty(address common.Address) bool {
	p.registry.RegisterAddressOp(EmptyID, &address)
	return p.db.Empty(address)
}

// Prepare handles the preparatory steps for executing a state transition.
func (p *EventProxy) Prepare(rules params.Rules, sender, coinbase common.Address, dest *common.Address, precompiles []common.Address, txAccesses types.AccessList) {
	p.db.Prepare(rules, sender, coinbase, dest, precompiles, txAccesses)
}

// AddAddressToAccessList adds an address to the access list.
func (p *EventProxy) AddAddressToAccessList(address common.Address) {
	p.db.AddAddressToAccessList(address)
}

// AddressInAccessList checks whether an address is in the access list.
func (p *EventProxy) AddressInAccessList(address common.Address) bool {
	return p.db.AddressInAccessList(address)
}

// SlotInAccessList checks whether the (address, slot)-tuple is in the access list.
func (p *EventProxy) SlotInAccessList(address common.Address, slot common.Hash) (bool, bool) {
	return p.db.SlotInAccessList(address, slot)
}

// AddSlotToAccessList adds the given (address, slot)-tuple to the access list
func (p *EventProxy) AddSlotToAccessList(address common.Address, slot common.Hash) {
	p.db.AddSlotToAccessList(address, slot)
}

// RevertToSnapshot reverts all state changes from a given revision.
func (p *EventProxy) RevertToSnapshot(snapshot int) {
	p.registry.RegisterOp(RevertToSnapshotID)
	// find snapshot in recordings and record how many snapshots have to be unwound
	for i, recordedSnapshot := range p.snapshots {
		if recordedSnapshot == snapshot {
			delta := len(p.snapshots) - i - 1
			p.registry.RegisterSnapshotDelta(delta)
			// remove snapshot from the active snapshot list
			// i.e., the snapshot given as an argument cannot
			// be reused for another call to RevertToSnapshot
			p.snapshots = p.snapshots[0 : i]
			break
		}
	}
	p.db.RevertToSnapshot(snapshot)
}

// Snapshot returns an identifier for the current revision of the state.
func (p *EventProxy) Snapshot() int {
	p.registry.RegisterOp(SnapshotID)
	snapshot := p.db.Snapshot()
	p.snapshots = append(p.snapshots, snapshot)
	return snapshot
}

// AddLog adds a log entry.
func (p *EventProxy) AddLog(log *types.Log) {
	p.db.AddLog(log)
}

// GetLogs retrieves log entries.
func (p *EventProxy) GetLogs(hash common.Hash, block uint64, blockHash common.Hash, blkTimestamp uint64) []*types.Log {
	return p.db.GetLogs(hash, block, blockHash, blkTimestamp)
}

// PointCache returns the point cache used in computations.
func (p *EventProxy) PointCache() *utils.PointCache {
	return p.db.PointCache()
}

// Witness retrieves the current state witness.
func (p *EventProxy) Witness() *stateless.Witness {
	return p.db.Witness()
}

// AddPreimage adds a SHA3 preimage.
func (p *EventProxy) AddPreimage(address common.Hash, image []byte) {
	// call real StateDB
	p.db.AddPreimage(address, image)
}

// AccessEvents retrieves events.
func (p *EventProxy) AccessEvents() *geth_state.AccessEvents {
	return p.db.AccessEvents()
}

// SetTxContext sets the current transaction hash and index.
func (p *EventProxy) SetTxContext(thash common.Hash, ti int) {
	p.db.SetTxContext(thash, ti)
}

// Finalise the state in StateDB.
func (p *EventProxy) Finalise(deleteEmptyObjects bool) {
	p.db.Finalise(deleteEmptyObjects)
}

// IntermediateRoot computes the current hash of the StateDB.
func (p *EventProxy) IntermediateRoot(deleteEmptyObjects bool) common.Hash {
	return p.db.IntermediateRoot(deleteEmptyObjects)
}

// Commit StateDB
func (p *EventProxy) Commit(block uint64, deleteEmptyObjects bool) (common.Hash, error) {
	return p.db.Commit(block, deleteEmptyObjects)
}

func (p *EventProxy) GetHash() (common.Hash, error) {
	return p.db.GetHash()
}

func (p *EventProxy) Error() error {
	return p.db.Error()
}

// GetSubstatePostAlloc gets substate post allocation.
func (p *EventProxy) GetSubstatePostAlloc() txcontext.WorldState {
	return p.db.GetSubstatePostAlloc()
}

func (p *EventProxy) PrepareSubstate(substate txcontext.WorldState, block uint64) {
	p.db.PrepareSubstate(substate, block)
}

func (p *EventProxy) BeginTransaction(number uint32) error {
	p.registry.RegisterOp(BeginTransactionID)
	if err := p.db.BeginTransaction(number); err != nil {
		return err
	}
	// clear all snapshots
	p.snapshots = []int{}
	return nil
}

func (p *EventProxy) EndTransaction() error {
	p.registry.RegisterOp(EndTransactionID)
	if err := p.db.EndTransaction(); err != nil {
		return err
	}
	// clear all snapshots
	p.snapshots = []int{}
	return nil
}

func (p *EventProxy) BeginBlock(number uint64) error {
	p.registry.RegisterOp(BeginBlockID)
	return p.db.BeginBlock(number)
}

func (p *EventProxy) EndBlock() error {
	p.registry.RegisterOp(EndBlockID)
	return p.db.EndBlock()
}

func (p *EventProxy) BeginSyncPeriod(number uint64) {
	p.registry.RegisterOp(BeginSyncPeriodID)
	p.db.BeginSyncPeriod(number)
}

func (p *EventProxy) EndSyncPeriod() {
	p.registry.RegisterOp(EndSyncPeriodID)
	p.db.EndSyncPeriod()
}

func (p *EventProxy) Close() error {
	return p.db.Close()
}

func (p *EventProxy) StartBulkLoad(uint64) (state.BulkLoad, error) {
	panic("StartBulkLoad not supported by EventProxy")
}

func (p *EventProxy) GetMemoryUsage() *state.MemoryUsage {
	return p.db.GetMemoryUsage()
}

func (p *EventProxy) GetArchiveState(block uint64) (state.NonCommittableStateDB, error) {
	return p.db.GetArchiveState(block)
}

func (p *EventProxy) GetArchiveBlockHeight() (uint64, bool, error) {
	return p.db.GetArchiveBlockHeight()
}

func (p *EventProxy) GetShadowDB() state.StateDB {
	return p.db.GetShadowDB()
}
