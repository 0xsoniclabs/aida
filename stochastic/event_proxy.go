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

// TODO: Provide Mocking tests for proxy

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
	// register event
	p.registry.RegisterAddressOp(CreateAccountID, &address)

	// call real StateDB
	p.db.CreateAccount(address)
}

// SubBalance subtracts amount from a contract address.
func (p *EventProxy) SubBalance(address common.Address, amount *uint256.Int, reason tracing.BalanceChangeReason) uint256.Int {
	// register event
	p.registry.RegisterAddressOp(SubBalanceID, &address)

	// call real StateDB
	return p.db.SubBalance(address, amount, reason)
}

// AddBalance adds amount to a contract address.
func (p *EventProxy) AddBalance(address common.Address, amount *uint256.Int, reason tracing.BalanceChangeReason) uint256.Int {
	// register event
	p.registry.RegisterAddressOp(AddBalanceID, &address)

	// call real StateDB
	return p.db.AddBalance(address, amount, reason)
}

// GetBalance retrieves the amount of a contract address.
func (p *EventProxy) GetBalance(address common.Address) *uint256.Int {
	// register event
	p.registry.RegisterAddressOp(GetBalanceID, &address)

	// call real StateDB
	return p.db.GetBalance(address)
}

// GetNonce retrieves the nonce of a contract address.
func (p *EventProxy) GetNonce(address common.Address) uint64 {
	// register event
	p.registry.RegisterAddressOp(GetNonceID, &address)

	// call real StateDB
	return p.db.GetNonce(address)
}

// SetNonce sets the nonce of a contract address.
func (p *EventProxy) SetNonce(address common.Address, nonce uint64, reason tracing.NonceChangeReason) {
	// register event
	p.registry.RegisterAddressOp(SetNonceID, &address)

	// call real StateDB
	p.db.SetNonce(address, nonce, reason)
}

// GetCodeHash returns the hash of the EVM bytecode.
func (p *EventProxy) GetCodeHash(address common.Address) common.Hash {
	// register event
	p.registry.RegisterAddressOp(GetCodeHashID, &address)

	// call real StateDB
	return p.db.GetCodeHash(address)
}

// GetCode returns the EVM bytecode of a contract.
func (p *EventProxy) GetCode(address common.Address) []byte {
	// register event
	p.registry.RegisterAddressOp(GetCodeID, &address)

	// call real StateDB
	return p.db.GetCode(address)
}

// SetCode sets the EVM bytecode of a contract.
func (p *EventProxy) SetCode(address common.Address, code []byte) []byte {
	// register event
	p.registry.RegisterAddressOp(SetCodeID, &address)

	// call real StateDB
	return p.db.SetCode(address, code)
}

// GetCodeSize returns the EVM bytecode's size.
func (p *EventProxy) GetCodeSize(address common.Address) int {
	// register event
	p.registry.RegisterAddressOp(GetCodeSizeID, &address)

	// call real StateDB
	return p.db.GetCodeSize(address)
}

// AddRefund adds gas to the refund counter.
func (p *EventProxy) AddRefund(gas uint64) {
	// call real StateDB
	p.db.AddRefund(gas)
}

// SubRefund subtracts gas to the refund counter.
func (p *EventProxy) SubRefund(gas uint64) {
	// call real StateDB
	p.db.SubRefund(gas)
}

// GetRefund returns the current value of the refund counter.
func (p *EventProxy) GetRefund() uint64 {
	// call real StateDB
	return p.db.GetRefund()
}

// GetCommittedState retrieves a value that is already committed.
func (p *EventProxy) GetCommittedState(address common.Address, key common.Hash) common.Hash {
	// register event
	p.registry.RegisterKeyOp(GetCommittedStateID, &address, &key)

	// call real StateDB
	return p.db.GetCommittedState(address, key)
}

// GetState retrieves a value from the StateDB.
func (p *EventProxy) GetState(address common.Address, key common.Hash) common.Hash {
	// register event
	p.registry.RegisterKeyOp(GetStateID, &address, &key)

	// call real StateDB
	return p.db.GetState(address, key)
}

// SetState sets a value in the StateDB.
func (p *EventProxy) SetState(address common.Address, key common.Hash, value common.Hash) common.Hash {
	// register event
	p.registry.RegisterValueOp(SetStateID, &address, &key, &value)

	// call real StateDB
	return p.db.SetState(address, key, value)
}
func (p *EventProxy) SetTransientState(addr common.Address, key common.Hash, value common.Hash) {
	p.registry.RegisterValueOp(SetTransientStateID, &addr, &key, &value)
	p.db.SetTransientState(addr, key, value)
}

func (p *EventProxy) GetTransientState(addr common.Address, key common.Hash) common.Hash {
	p.registry.RegisterKeyOp(GetTransientStateID, &addr, &key)
	return p.db.GetState(addr, key)
}

// SelfDestruct an account.
func (p *EventProxy) SelfDestruct(address common.Address) uint256.Int {
	// register event
	p.registry.RegisterAddressOp(SelfDestructID, &address)

	// call real StateDB
	return p.db.SelfDestruct(address)
}

// HasSelfDestructed checks whether a contract has been suicided.
func (p *EventProxy) HasSelfDestructed(address common.Address) bool {
	// register event
	p.registry.RegisterAddressOp(HasSelfDestructedID, &address)

	// call real StateDB
	return p.db.HasSelfDestructed(address)
}

// Exist checks whether the contract exists in the StateDB.
func (p *EventProxy) Exist(address common.Address) bool {
	// register event
	p.registry.RegisterAddressOp(ExistID, &address)

	// call real StateDB
	return p.db.Exist(address)
}

// Empty checks whether the contract is either non-existent
// or empty according to the EIP161 specification (balance = nonce = code = 0).
func (p *EventProxy) Empty(address common.Address) bool {
	// register event
	p.registry.RegisterAddressOp(EmptyID, &address)

	// call real StateDB
	return p.db.Empty(address)
}

// Prepare handles the preparatory steps for executing a state transition.
func (p *EventProxy) Prepare(rules params.Rules, sender, coinbase common.Address, dest *common.Address, precompiles []common.Address, txAccesses types.AccessList) {
	// call real StateDB
	p.db.Prepare(rules, sender, coinbase, dest, precompiles, txAccesses)
}

// AddAddressToAccessList adds an address to the access list.
func (p *EventProxy) AddAddressToAccessList(address common.Address) {
	// call real StateDB
	p.db.AddAddressToAccessList(address)
}

// AddressInAccessList checks whether an address is in the access list.
func (p *EventProxy) AddressInAccessList(address common.Address) bool {
	// call real StateDB
	return p.db.AddressInAccessList(address)
}

// SlotInAccessList checks whether the (address, slot)-tuple is in the access list.
func (p *EventProxy) SlotInAccessList(address common.Address, slot common.Hash) (bool, bool) {
	// call real StateDB
	return p.db.SlotInAccessList(address, slot)
}

// AddSlotToAccessList adds the given (address, slot)-tuple to the access list
func (p *EventProxy) AddSlotToAccessList(address common.Address, slot common.Hash) {
	// call real StateDB
	p.db.AddSlotToAccessList(address, slot)
}

// RevertToSnapshot reverts all state changes from a given revision.
func (p *EventProxy) RevertToSnapshot(snapshot int) {
	// register event
	p.registry.RegisterOp(RevertToSnapshotID)

	// find snapshot
	for i, recordedSnapshot := range p.snapshots {
		if recordedSnapshot == snapshot {
			// register snapshot delta
			p.registry.RegisterSnapshotDelta(len(p.snapshots) - i - 1)

			// cull all elements between found snapshot and top-of-stack
			p.snapshots = p.snapshots[0:i]
			break
		}
	}

	// call real StateDB
	p.db.RevertToSnapshot(snapshot)
}

// Snapshot returns an identifier for the current revision of the state.
func (p *EventProxy) Snapshot() int {
	// register event
	p.registry.RegisterOp(SnapshotID)

	// call real StateDB
	snapshot := p.db.Snapshot()

	// add snapshot
	p.snapshots = append(p.snapshots, snapshot)

	return snapshot
}

// AddLog adds a log entry.
func (p *EventProxy) AddLog(log *types.Log) {
	// call real StateDB
	p.db.AddLog(log)
}

// GetLogs retrieves log entries.
func (p *EventProxy) GetLogs(hash common.Hash, block uint64, blockHash common.Hash, blkTimestamp uint64) []*types.Log {
	// call real StateDB
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

func (p *EventProxy) AccessEvents() *geth_state.AccessEvents {
	return p.db.AccessEvents()
}

// SetTxContext sets the current transaction hash and index.
func (p *EventProxy) SetTxContext(thash common.Hash, ti int) {
	// call real StateDB
	p.db.SetTxContext(thash, ti)
}

// Finalise the state in StateDB.
func (p *EventProxy) Finalise(deleteEmptyObjects bool) {
	// call real StateDB
	p.db.Finalise(deleteEmptyObjects)
}

// IntermediateRoot computes the current hash of the StateDB.
func (p *EventProxy) IntermediateRoot(deleteEmptyObjects bool) common.Hash {
	// call real StateDB
	return p.db.IntermediateRoot(deleteEmptyObjects)
}

// Commit StateDB
func (p *EventProxy) Commit(block uint64, deleteEmptyObjects bool) (common.Hash, error) {
	// call real StateDB
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
	// call real StateDB
	return p.db.GetSubstatePostAlloc()
}

func (p *EventProxy) PrepareSubstate(substate txcontext.WorldState, block uint64) {
	p.db.PrepareSubstate(substate, block)
}

func (p *EventProxy) BeginTransaction(number uint32) error {
	// register event
	p.registry.RegisterOp(BeginTransactionID)

	// call real StateDB
	if err := p.db.BeginTransaction(number); err != nil {
		return err
	}

	// clear all snapshots
	p.snapshots = []int{}
	return nil
}

func (p *EventProxy) EndTransaction() error {
	// register event
	p.registry.RegisterOp(EndTransactionID)

	// call real StateDB
	if err := p.db.EndTransaction(); err != nil {
		return err
	}

	// clear all snapshots
	p.snapshots = []int{}
	return nil
}

func (p *EventProxy) BeginBlock(number uint64) error {
	// register event
	p.registry.RegisterOp(BeginBlockID)

	// call real StateDB
	return p.db.BeginBlock(number)
}

func (p *EventProxy) EndBlock() error {
	// register event
	p.registry.RegisterOp(EndBlockID)

	// call real StateDB
	return p.db.EndBlock()
}

func (p *EventProxy) BeginSyncPeriod(number uint64) {
	// register event
	p.registry.RegisterOp(BeginSyncPeriodID)

	// call real StateDB
	p.db.BeginSyncPeriod(number)
}

func (p *EventProxy) EndSyncPeriod() {
	// register event
	p.registry.RegisterOp(EndSyncPeriodID)

	// call real StateDB
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

func (p *EventProxy) CreateContract(addr common.Address) {
	p.registry.RegisterOp(CreateContractID)
	p.db.CreateContract(addr)
}

func (p *EventProxy) SelfDestruct6780(addr common.Address) (uint256.Int, bool) {
	p.registry.RegisterOp(SelfDestruct6780ID)
	return p.db.SelfDestruct6780(addr)
}

func (p *EventProxy) GetStorageRoot(addr common.Address) common.Hash {
	p.registry.RegisterOp(CreateContractID)
	return p.db.GetStorageRoot(addr)
}
