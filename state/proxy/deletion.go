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

type ContractLiveliness struct {
	Addr      common.Address
	IsDeleted bool //if false, the account was created
}

// DeletionProxy data structure for capturing and recording
// invoked StateDB operations.
type DeletionProxy struct {
	db  state.StateDB // state db
	ch  chan ContractLiveliness
	log logger.Logger
}

// NewDeletionProxy creates a new StateDB proxy.
func NewDeletionProxy(db state.StateDB, ch chan ContractLiveliness, logLevel string) *DeletionProxy {
	r := new(DeletionProxy)
	r.db = db
	r.ch = ch
	r.log = logger.NewLogger(logLevel, "Proxy Deletion")
	return r
}

// CreateAccount creates a new account.
func (r *DeletionProxy) CreateAccount(addr common.Address) {
	r.db.CreateAccount(addr)
	r.ch <- ContractLiveliness{Addr: addr, IsDeleted: false}
}

// SubBalance subtracts amount from a contract address.
func (r *DeletionProxy) SubBalance(addr common.Address, amount *uint256.Int, reason tracing.BalanceChangeReason) uint256.Int {
	return r.db.SubBalance(addr, amount, reason)
}

// AddBalance adds amount to a contract address.
func (r *DeletionProxy) AddBalance(addr common.Address, amount *uint256.Int, reason tracing.BalanceChangeReason) uint256.Int {
	return r.db.AddBalance(addr, amount, reason)
}

// GetBalance retrieves the amount of a contract address.
func (r *DeletionProxy) GetBalance(addr common.Address) *uint256.Int {
	balance := r.db.GetBalance(addr)
	return balance
}

// GetNonce retrieves the nonce of a contract address.
func (r *DeletionProxy) GetNonce(addr common.Address) uint64 {
	nonce := r.db.GetNonce(addr)
	return nonce
}

// SetNonce sets the nonce of a contract address.
func (r *DeletionProxy) SetNonce(addr common.Address, nonce uint64, reason tracing.NonceChangeReason) {
	r.db.SetNonce(addr, nonce, reason)
}

// GetCodeHash returns the hash of the EVM bytecode.
func (r *DeletionProxy) GetCodeHash(addr common.Address) common.Hash {
	hash := r.db.GetCodeHash(addr)
	return hash
}

// GetCode returns the EVM bytecode of a contract.
func (r *DeletionProxy) GetCode(addr common.Address) []byte {
	code := r.db.GetCode(addr)
	return code
}

// SetCode sets the EVM bytecode of a contract.
func (r *DeletionProxy) SetCode(addr common.Address, code []byte) []byte {
	return r.db.SetCode(addr, code)
}

// GetCodeSize returns the EVM bytecode's size.
func (r *DeletionProxy) GetCodeSize(addr common.Address) int {
	size := r.db.GetCodeSize(addr)
	return size
}

// AddRefund adds gas to the refund counter.
func (r *DeletionProxy) AddRefund(gas uint64) {
	r.db.AddRefund(gas)
}

// SubRefund subtracts gas to the refund counter.
func (r *DeletionProxy) SubRefund(gas uint64) {
	r.db.SubRefund(gas)
}

// GetRefund returns the current value of the refund counter.
func (r *DeletionProxy) GetRefund() uint64 {
	gas := r.db.GetRefund()
	return gas
}

// GetCommittedState retrieves a value that is already committed.
func (r *DeletionProxy) GetCommittedState(addr common.Address, key common.Hash) common.Hash {
	value := r.db.GetCommittedState(addr, key)
	return value
}

// GetState retrieves a value from the StateDB.
func (r *DeletionProxy) GetState(addr common.Address, key common.Hash) common.Hash {
	value := r.db.GetState(addr, key)
	return value
}

// SetState sets a value in the StateDB.
func (r *DeletionProxy) SetState(addr common.Address, key common.Hash, value common.Hash) common.Hash {
	return r.db.SetState(addr, key, value)
}

func (r *DeletionProxy) SetTransientState(addr common.Address, key common.Hash, value common.Hash) {
	r.db.SetTransientState(addr, key, value)
}

func (r *DeletionProxy) GetTransientState(addr common.Address, key common.Hash) common.Hash {
	return r.db.GetTransientState(addr, key)
}

// SelfDestruct marks the given account as suicided. This clears the account balance.
// The account is still available until the state is committed;
// return a non-nil account after SelfDestruct.
func (r *DeletionProxy) SelfDestruct(addr common.Address) uint256.Int {
	res := r.db.SelfDestruct(addr)
	r.ch <- ContractLiveliness{Addr: addr, IsDeleted: true}
	return res
}

// HasSelfDestructed checks whether a contract has been suicided.
func (r *DeletionProxy) HasSelfDestructed(addr common.Address) bool {
	hasSelfDestructed := r.db.HasSelfDestructed(addr)
	return hasSelfDestructed
}

// Exist checks whether the contract exists in the StateDB.
// Notably this also returns true for suicided accounts.
func (r *DeletionProxy) Exist(addr common.Address) bool {
	return r.db.Exist(addr)
}

// Empty checks whether the contract is either non-existent
// or empty according to the EIP161 specification (balance = nonce = code = 0).
func (r *DeletionProxy) Empty(addr common.Address) bool {
	empty := r.db.Empty(addr)
	return empty
}

// Prepare handles the preparatory steps for executing a state transition with
// regards to both EIP-2929 and EIP-2930:
//
// - Add sender to access list (2929)
// - Add destination to access list (2929)
// - Add precompiles to access list (2929)
// - Add the contents of the optional tx access list (2930)
//
// This method should only be called if Berlin/2929+2930 is applicable at the current number.
func (r *DeletionProxy) Prepare(rules params.Rules, sender, coinbase common.Address, dest *common.Address, precompiles []common.Address, txAccesses types.AccessList) {
	r.db.Prepare(rules, sender, coinbase, dest, precompiles, txAccesses)
}

// AddAddressToAccessList adds an address to the access list.
func (r *DeletionProxy) AddAddressToAccessList(addr common.Address) {
	r.db.AddAddressToAccessList(addr)
}

// AddressInAccessList checks whether an address is in the access list.
func (r *DeletionProxy) AddressInAccessList(addr common.Address) bool {
	ok := r.db.AddressInAccessList(addr)
	return ok
}

// SlotInAccessList checks whether the (address, slot)-tuple is in the access list.
func (r *DeletionProxy) SlotInAccessList(addr common.Address, slot common.Hash) (bool, bool) {
	addressOk, slotOk := r.db.SlotInAccessList(addr, slot)
	return addressOk, slotOk
}

// AddSlotToAccessList adds the given (address, slot)-tuple to the access list
func (r *DeletionProxy) AddSlotToAccessList(addr common.Address, slot common.Hash) {
	r.db.AddSlotToAccessList(addr, slot)
}

// RevertToSnapshot reverts all state changes from a given revision.
func (r *DeletionProxy) RevertToSnapshot(snapshot int) {
	r.db.RevertToSnapshot(snapshot)
}

// Snapshot returns an identifier for the current revision of the state.
func (r *DeletionProxy) Snapshot() int {
	snapshot := r.db.Snapshot()
	return snapshot
}

// AddLog adds a log entry.
func (r *DeletionProxy) AddLog(log *types.Log) {
	r.db.AddLog(log)
}

// GetLogs retrieves log entries.
func (r *DeletionProxy) GetLogs(hash common.Hash, block uint64, blockHash common.Hash, blkTimestamp uint64) []*types.Log {
	return r.db.GetLogs(hash, block, blockHash, blkTimestamp)
}

// PointCache returns the point cache used in computations.
func (r *DeletionProxy) PointCache() *utils.PointCache {
	return r.db.PointCache()
}

// Witness retrieves the current state witness.
func (r *DeletionProxy) Witness() *stateless.Witness {
	return r.db.Witness()
}

// AddPreimage adds a SHA3 preimage.
func (r *DeletionProxy) AddPreimage(addr common.Hash, image []byte) {
	r.db.AddPreimage(addr, image)
}

func (r *DeletionProxy) AccessEvents() *geth.AccessEvents {
	return r.db.AccessEvents()
}

// Prepare sets the current transaction hash and index.
func (r *DeletionProxy) SetTxContext(thash common.Hash, ti int) {
	r.db.SetTxContext(thash, ti)
}

// Finalise the state in StateDB.
func (r *DeletionProxy) Finalise(deleteEmptyObjects bool) {
	r.db.Finalise(deleteEmptyObjects)
}

// IntermediateRoot computes the current hash of the StateDB.
// It is called in between transactions to get the root hash that
// goes into transaction receipts.
func (r *DeletionProxy) IntermediateRoot(deleteEmptyObjects bool) common.Hash {
	return r.db.IntermediateRoot(deleteEmptyObjects)
}

func (r *DeletionProxy) Commit(block uint64, deleteEmptyObjects bool) (common.Hash, error) {
	return r.db.Commit(block, deleteEmptyObjects)
}

func (r *DeletionProxy) GetHash() (common.Hash, error) {
	return r.db.GetHash()
}

func (r *DeletionProxy) Error() error {
	return r.db.Error()
}

// GetSubstatePostAlloc gets substate post allocation.
func (r *DeletionProxy) GetSubstatePostAlloc() txcontext.WorldState {
	return r.db.GetSubstatePostAlloc()
}

func (r *DeletionProxy) PrepareSubstate(substate txcontext.WorldState, block uint64) {
	r.db.PrepareSubstate(substate, block)
}

func (r *DeletionProxy) BeginTransaction(number uint32) error {
	return r.db.BeginTransaction(number)
}

func (r *DeletionProxy) EndTransaction() error {
	return r.db.EndTransaction()
}

func (r *DeletionProxy) BeginBlock(number uint64) error {
	return r.db.BeginBlock(number)
}

func (r *DeletionProxy) EndBlock() error {
	return r.db.EndBlock()
}

func (r *DeletionProxy) BeginSyncPeriod(number uint64) {
	r.db.BeginSyncPeriod(number)
}

func (r *DeletionProxy) EndSyncPeriod() {
	r.db.EndSyncPeriod()
}

func (r *DeletionProxy) GetArchiveState(block uint64) (state.NonCommittableStateDB, error) {
	return r.db.GetArchiveState(block)
}

func (r *DeletionProxy) GetArchiveBlockHeight() (uint64, bool, error) {
	return r.db.GetArchiveBlockHeight()
}

func (r *DeletionProxy) Close() error {
	return r.db.Close()
}

func (r *DeletionProxy) StartBulkLoad(uint64) (state.BulkLoad, error) {
	r.log.Fatal("StartBulkLoad not supported by DeletionProxy")
	return nil, nil
}

func (r *DeletionProxy) GetMemoryUsage() *state.MemoryUsage {
	return r.db.GetMemoryUsage()
}

func (r *DeletionProxy) GetShadowDB() state.StateDB {
	return r.db.GetShadowDB()
}

func (r *DeletionProxy) CreateContract(addr common.Address) {
	r.db.CreateContract(addr)
}

func (r *DeletionProxy) SelfDestruct6780(addr common.Address) (uint256.Int, bool) {
	return r.db.SelfDestruct6780(addr)
}

func (r *DeletionProxy) GetStorageRoot(addr common.Address) common.Hash {
	return r.db.GetStorageRoot(addr)
}
