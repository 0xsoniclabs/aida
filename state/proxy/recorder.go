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
	"fmt"

	"github.com/0xsoniclabs/aida/state"
	"github.com/0xsoniclabs/aida/tracer/context"
	"github.com/0xsoniclabs/aida/tracer/operation"
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

// RecorderProxy data structure for capturing and recording
// invoked StateDB operations.
type RecorderProxy struct {
	db  state.StateDB   // state db
	ctx *context.Record // record context for recording StateDB operations in a tracefile
}

// NewRecorderProxy creates a new StateDB proxy.
func NewRecorderProxy(db state.StateDB, ctx *context.Record) *RecorderProxy {
	return &RecorderProxy{
		db:  db,
		ctx: ctx,
	}
}

// write new operation to file.
func (r *RecorderProxy) write(op operation.Operation) {
	operation.WriteOp(r.ctx, op)
}

// CreateAccount creates a new account.
func (r *RecorderProxy) CreateAccount(addr common.Address) {
	contract := r.ctx.EncodeContract(addr)
	r.write(operation.NewCreateAccount(contract))
	r.db.CreateAccount(addr)
}

// SubBalance subtracts amount from a contract address.
func (r *RecorderProxy) SubBalance(addr common.Address, amount *uint256.Int, reason tracing.BalanceChangeReason) uint256.Int {
	contract := r.ctx.EncodeContract(addr)
	r.write(operation.NewSubBalance(contract, amount, reason))
	return r.db.SubBalance(addr, amount, reason)
}

// AddBalance adds amount to a contract address.
func (r *RecorderProxy) AddBalance(addr common.Address, amount *uint256.Int, reason tracing.BalanceChangeReason) uint256.Int {
	contract := r.ctx.EncodeContract(addr)
	r.write(operation.NewAddBalance(contract, amount, reason))
	return r.db.AddBalance(addr, amount, reason)
}

// GetBalance retrieves the amount of a contract address.
func (r *RecorderProxy) GetBalance(addr common.Address) *uint256.Int {
	contract := r.ctx.EncodeContract(addr)
	r.write(operation.NewGetBalance(contract))
	balance := r.db.GetBalance(addr)
	return balance
}

// GetNonce retrieves the nonce of a contract address.
func (r *RecorderProxy) GetNonce(addr common.Address) uint64 {
	contract := r.ctx.EncodeContract(addr)
	r.write(operation.NewGetNonce(contract))
	nonce := r.db.GetNonce(addr)
	return nonce
}

// SetNonce sets the nonce of a contract address.
func (r *RecorderProxy) SetNonce(addr common.Address, nonce uint64, reason tracing.NonceChangeReason) {
	contract := r.ctx.EncodeContract(addr)
	r.write(operation.NewSetNonce(contract, nonce, reason))
	r.db.SetNonce(addr, nonce, reason)
}

// GetCodeHash returns the hash of the EVM bytecode.
func (r *RecorderProxy) GetCodeHash(addr common.Address) common.Hash {
	previousContract := r.ctx.PrevContract()
	contract := r.ctx.EncodeContract(addr)
	if previousContract == contract {
		r.write(operation.NewGetCodeHashLc())
	} else {
		r.write(operation.NewGetCodeHash(contract))
	}

	hash := r.db.GetCodeHash(addr)
	return hash
}

// GetCode returns the EVM bytecode of a contract.
func (r *RecorderProxy) GetCode(addr common.Address) []byte {
	contract := r.ctx.EncodeContract(addr)
	r.write(operation.NewGetCode(contract))
	code := r.db.GetCode(addr)
	return code
}

// SetCode sets the EVM bytecode of a contract.
func (r *RecorderProxy) SetCode(addr common.Address, code []byte) []byte {
	contract := r.ctx.EncodeContract(addr)
	r.write(operation.NewSetCode(contract, code))
	return r.db.SetCode(addr, code)
}

// GetCodeSize returns the EVM bytecode's size.
func (r *RecorderProxy) GetCodeSize(addr common.Address) int {
	contract := r.ctx.EncodeContract(addr)
	r.write(operation.NewGetCodeSize(contract))
	size := r.db.GetCodeSize(addr)
	return size
}

// AddRefund adds gas to the refund counter.
func (r *RecorderProxy) AddRefund(gas uint64) {
	r.db.AddRefund(gas)
}

// SubRefund subtracts gas to the refund counter.
func (r *RecorderProxy) SubRefund(gas uint64) {
	r.db.SubRefund(gas)
}

// GetRefund returns the current value of the refund counter.
func (r *RecorderProxy) GetRefund() uint64 {
	gas := r.db.GetRefund()
	return gas
}

// GetCommittedState retrieves a value that is already committed.
func (r *RecorderProxy) GetCommittedState(addr common.Address, key common.Hash) common.Hash {
	previousContract := r.ctx.PrevContract()
	contract := r.ctx.EncodeContract(addr)
	key, kPos := r.ctx.EncodeKey(key)
	if previousContract == contract && kPos == 0 {
		r.write(operation.NewGetCommittedStateLcls())
	} else {
		r.write(operation.NewGetCommittedState(contract, key))
	}
	value := r.db.GetCommittedState(addr, key)
	return value
}

// GetState retrieves a value from the StateDB.
func (r *RecorderProxy) GetState(addr common.Address, key common.Hash) common.Hash {
	previousContract := r.ctx.PrevContract()
	contract := r.ctx.EncodeContract(addr)
	key, kPos := r.ctx.EncodeKey(key)
	var op operation.Operation
	if contract == previousContract {
		if kPos == 0 {
			op = operation.NewGetStateLcls()
		} else if kPos != -1 {
			op = operation.NewGetStateLccs(kPos)
		} else {
			op = operation.NewGetStateLc(key)
		}
	} else {
		op = operation.NewGetState(contract, key)
	}
	r.write(op)
	value := r.db.GetState(addr, key)
	return value
}

// SetState sets a value in the StateDB.
func (r *RecorderProxy) SetState(addr common.Address, key common.Hash, value common.Hash) common.Hash {
	previousContract := r.ctx.PrevContract()
	contract := r.ctx.EncodeContract(addr)
	key, kPos := r.ctx.EncodeKey(key)
	if contract == previousContract && kPos == 0 {
		r.write(operation.NewSetStateLcls(value))
	} else {
		r.write(operation.NewSetState(contract, key, value))
	}
	return r.db.SetState(addr, key, value)
}

func (r *RecorderProxy) SetTransientState(addr common.Address, key common.Hash, value common.Hash) {
	previousContract := r.ctx.PrevContract()
	contract := r.ctx.EncodeContract(addr)
	key, kPos := r.ctx.EncodeKey(key)

	if contract == previousContract && kPos == 0 {
		r.write(operation.NewSetTransientStateLcls(value))
	} else {
		r.write(operation.NewSetTransientState(contract, key, value))
	}
	r.db.SetTransientState(addr, key, value)
}

func (r *RecorderProxy) GetTransientState(addr common.Address, key common.Hash) common.Hash {
	previousContract := r.ctx.PrevContract()
	contract := r.ctx.EncodeContract(addr)
	key, kPos := r.ctx.EncodeKey(key)
	var op operation.Operation
	if contract == previousContract {
		if kPos == 0 {
			op = operation.NewGetTransientStateLcls()
		} else if kPos != -1 {
			op = operation.NewGetTransientStateLccs(kPos)
		} else {
			op = operation.NewGetTransientStateLc(key)
		}
	} else {
		op = operation.NewGetTransientState(contract, key)
	}
	r.write(op)
	value := r.db.GetTransientState(addr, key)
	return value
}

// SelfDestruct marks the given account as suicided. This clears the account balance.
// The account is still available until the state is committed;
// return a non-nil account after SelfDestruct.
func (r *RecorderProxy) SelfDestruct(addr common.Address) uint256.Int {
	contract := r.ctx.EncodeContract(addr)
	r.write(operation.NewSelfDestruct(contract))
	return r.db.SelfDestruct(addr)
}

// HasSelfDestructed checks whether a contract has been suicided.
func (r *RecorderProxy) HasSelfDestructed(addr common.Address) bool {
	hasSelfDestructed := r.db.HasSelfDestructed(addr)
	r.write(operation.NewHasSelfDestructed(addr))
	return hasSelfDestructed
}

// Exist checks whether the contract exists in the StateDB.
// Notably this also returns true for suicided accounts.
func (r *RecorderProxy) Exist(addr common.Address) bool {
	contract := r.ctx.EncodeContract(addr)
	r.write(operation.NewExist(contract))
	return r.db.Exist(addr)
}

// Empty checks whether the contract is either non-existent
// or empty according to the EIP161 specification (balance = nonce = code = 0).
func (r *RecorderProxy) Empty(addr common.Address) bool {
	empty := r.db.Empty(addr)
	return empty
}

// PrepareAccessList handles the preparatory steps for executing a state transition with
// regards to both EIP-2929 and EIP-2930:
//
// - Add writeer to access list (2929)
// - Add destination to access list (2929)
// - Add precompiles to access list (2929)
// - Add the contents of the optional tx access list (2930)
//
// This method should only be called if Berlin/2929+2930 is applicable at the current number.
func (r *RecorderProxy) Prepare(rules params.Rules, sender, coinbase common.Address, dest *common.Address, precompiles []common.Address, txAccesses types.AccessList) {
	r.db.Prepare(rules, sender, coinbase, dest, precompiles, txAccesses)
}

// AddAddressToAccessList adds an address to the access list.
func (r *RecorderProxy) AddAddressToAccessList(addr common.Address) {
	r.db.AddAddressToAccessList(addr)
}

// AddressInAccessList checks whether an address is in the access list.
func (r *RecorderProxy) AddressInAccessList(addr common.Address) bool {
	ok := r.db.AddressInAccessList(addr)
	return ok
}

// SlotInAccessList checks whether the (address, slot)-tuple is in the access list.
func (r *RecorderProxy) SlotInAccessList(addr common.Address, slot common.Hash) (bool, bool) {
	addressOk, slotOk := r.db.SlotInAccessList(addr, slot)
	return addressOk, slotOk
}

// AddSlotToAccessList adds the given (address, slot)-tuple to the access list
func (r *RecorderProxy) AddSlotToAccessList(addr common.Address, slot common.Hash) {
	r.db.AddSlotToAccessList(addr, slot)
}

// RevertToSnapshot reverts all state changes from a given revision.
func (r *RecorderProxy) RevertToSnapshot(snapshot int) {
	r.write(operation.NewRevertToSnapshot(snapshot))
	r.db.RevertToSnapshot(snapshot)
}

// Snapshot returns an identifier for the current revision of the state.
func (r *RecorderProxy) Snapshot() int {
	snapshot := r.db.Snapshot()
	// TODO: check overrun
	r.write(operation.NewSnapshot(int32(snapshot)))
	return snapshot
}

// AddLog adds a log entry.
func (r *RecorderProxy) AddLog(log *types.Log) {
	r.db.AddLog(log)
}

// GetLogs retrieves log entries.
func (r *RecorderProxy) GetLogs(hash common.Hash, block uint64, blockHash common.Hash, blkTimestamp uint64) []*types.Log {
	return r.db.GetLogs(hash, block, blockHash, blkTimestamp)
}

// PointCache returns the point cache used in computations.
func (r *RecorderProxy) PointCache() *utils.PointCache {
	//TODO record operation id
	return r.db.PointCache()
}

// Witness retrieves the current state witness.
func (r *RecorderProxy) Witness() *stateless.Witness {
	//TODO record operation id
	return r.db.Witness()
}

// AddPreimage adds a SHA3 preimage.
func (r *RecorderProxy) AddPreimage(addr common.Hash, image []byte) {
	r.db.AddPreimage(addr, image)
}

func (r *RecorderProxy) AccessEvents() *geth.AccessEvents {
	return r.db.AccessEvents()
}

// SetTxContext sets the current transaction hash and index.
func (r *RecorderProxy) SetTxContext(thash common.Hash, ti int) {
	r.db.SetTxContext(thash, ti)
}

// Finalise the state in StateDB.
func (r *RecorderProxy) Finalise(deleteEmptyObjects bool) {
	r.write(operation.NewFinalise(deleteEmptyObjects))
	r.db.Finalise(deleteEmptyObjects)
}

// IntermediateRoot computes the current hash of the StateDB.
// It is called in between transactions to get the root hash that
// goes into transaction receipts.
func (r *RecorderProxy) IntermediateRoot(deleteEmptyObjects bool) common.Hash {
	return r.db.IntermediateRoot(deleteEmptyObjects)
}

func (r *RecorderProxy) Commit(block uint64, deleteEmptyObjects bool) (common.Hash, error) {
	return r.db.Commit(block, deleteEmptyObjects)
}

func (r *RecorderProxy) Error() error {
	return r.db.Error()
}

// GetSubstatePostAlloc gets substate post allocation.
func (r *RecorderProxy) GetSubstatePostAlloc() txcontext.WorldState {
	return r.db.GetSubstatePostAlloc()
}

func (r *RecorderProxy) PrepareSubstate(substate txcontext.WorldState, block uint64) {
	r.db.PrepareSubstate(substate, block)
}

func (r *RecorderProxy) BeginTransaction(number uint32) error {
	r.write(operation.NewBeginTransaction(number))
	r.db.BeginTransaction(number)
	return nil
}

func (r *RecorderProxy) EndTransaction() error {
	r.write(operation.NewEndTransaction())
	r.db.EndTransaction()
	return nil
}

func (r *RecorderProxy) BeginBlock(number uint64) error {
	r.write(operation.NewBeginBlock(number))
	r.db.BeginBlock(number)
	return nil
}

func (r *RecorderProxy) EndBlock() error {
	r.write(operation.NewEndBlock())
	r.db.EndBlock()
	return nil
}

func (r *RecorderProxy) BeginSyncPeriod(number uint64) {
	r.write(operation.NewBeginSyncPeriod(number))
	r.db.BeginSyncPeriod(number)
}

func (r *RecorderProxy) EndSyncPeriod() {
	r.write(operation.NewEndSyncPeriod())
	r.db.EndSyncPeriod()
}

func (r *RecorderProxy) GetHash() (common.Hash, error) {
	// TODO: record this event
	return r.db.GetHash()
}

func (r *RecorderProxy) GetArchiveState(block uint64) (state.NonCommittableStateDB, error) {
	return nil, fmt.Errorf("archive states are not (yet) supported by this DB implementation")
}

func (r *RecorderProxy) GetArchiveBlockHeight() (uint64, bool, error) {
	return 0, false, fmt.Errorf("archive states are not (yet) supported by this DB implementation")
}

func (r *RecorderProxy) Close() error {
	return r.db.Close()
}

func (r *RecorderProxy) StartBulkLoad(uint64) (state.BulkLoad, error) {
	panic("StartBulkLoad not supported by RecorderProxy")
}

func (r *RecorderProxy) GetMemoryUsage() *state.MemoryUsage {
	return r.db.GetMemoryUsage()
}

func (r *RecorderProxy) GetShadowDB() state.StateDB {
	return r.db.GetShadowDB()
}
func (r *RecorderProxy) CreateContract(addr common.Address) {
	r.db.CreateContract(addr)
}

func (r *RecorderProxy) SelfDestruct6780(addr common.Address) (uint256.Int, bool) {
	return r.db.SelfDestruct6780(addr)
}

func (r *RecorderProxy) GetStorageRoot(addr common.Address) common.Hash {
	return r.db.GetStorageRoot(addr)
}
