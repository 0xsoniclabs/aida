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
	"github.com/0xsoniclabs/aida/txctx"
	"github.com/ethereum/go-ethereum/common"
	geth_state "github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/stateless"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/trie/utils"
	"github.com/holiman/uint256"
)

// TracerProxy data structure for writing StateDB operations
type TracerProxy struct {
	db   state.StateDB      // StateDB object
	ctx  *ArgumentContext   // context that keeps track of the argument history
	                        // of previous StateDB operations
}

// NewTracerProxy creates a new StateDB proxy for recording and writing events.
func NewTracerProxy(db state.StateDB, ctx *ArgumentContext) *TracerProxy {
	return &TracerProxy{
		db:   db,
		ctx:  ctx,
	}
}

func (p *TracerProxy) CreateAccount(address common.Address) {
	p.ctx.WriteAddressOp(CreateAccountID, &address, []byte{})
	p.db.CreateAccount(address)
}

func (p *TracerProxy) SubBalance(address common.Address, amount *uint256.Int, reason tracing.BalanceChangeReason) uint256.Int {
	// TBD: find an encoding for amount and reason in the form of a byte sequence
	data := []byte{} 
	p.ctx.WriteAddressOp(SubBalanceID, &address, data)
	return p.db.SubBalance(address, amount, reason)
}

func (p *TracerProxy) AddBalance(address common.Address, amount *uint256.Int, reason tracing.BalanceChangeReason) uint256.Int {
	// TBD: find an encoding for amount and reason in the form of a byte sequence
	data := []byte{} 
	p.ctx.WriteAddressOp(AddBalanceID, &address)
	return p.db.AddBalance(address, amount, reason)
}

func (p *TracerProxy) GetBalance(address common.Address) *uint256.Int {
	p.ctx.WriteAddressOp(GetBalanceID, &address, []byte{})
	return p.db.GetBalance(address)
}

func (p *TracerProxy) GetNonce(address common.Address) uint64 {
	p.ctx.WriteAddressOp(GetNonceID, &address)
	return p.db.GetNonce(address)
}

func (p *TracerProxy) SetNonce(address common.Address, nonce uint64, reason tracing.NonceChangeReason) {
	p.ctx.WriteAddressOp(SetNonceID, &address)
	p.db.SetNonce(address, nonce, reason)
}

func (p *TracerProxy) GetCodeHash(address common.Address) common.Hash {
	p.ctx.WriteAddressOp(GetCodeHashID, &address)
	return p.db.GetCodeHash(address)
}

func (p *TracerProxy) GetCode(address common.Address) []byte {
	p.ctx.WriteAddressOp(GetCodeID, &address)
	return p.db.GetCode(address)
}

func (p *TracerProxy) SetCode(address common.Address, code []byte) []byte {
	p.ctx.WriteAddressOp(SetCodeID, &address)
	return p.db.SetCode(address, code)
}

func (p *TracerProxy) GetCodeSize(address common.Address) int {
	p.ctx.WriteAddressOp(GetCodeSizeID, &address)
	return p.db.GetCodeSize(address)
}

func (p *TracerProxy) AddRefund(gas uint64) {
	p.db.AddRefund(gas)
}

func (p *TracerProxy) SubRefund(gas uint64) {
	p.db.SubRefund(gas)
}

func (p *TracerProxy) GetRefund() uint64 {
	return p.db.GetRefund()
}

func (p *TracerProxy) GetCommittedState(address common.Address, key common.Hash) common.Hash {
	p.ctx.WriteKeyOp(GetCommittedStateID, &address, &key)
	return p.db.GetCommittedState(address, key)
}

func (p *TracerProxy) GetState(address common.Address, key common.Hash) common.Hash {
	p.ctx.WriteKeyOp(GetStateID, &address, &key)
	return p.db.GetState(address, key)
}

func (p *TracerProxy) SetState(address common.Address, key common.Hash, value common.Hash) common.Hash {
	p.ctx.WriteValueOp(SetStateID, &address, &key, &value)
	return p.db.SetState(address, key, value)
}
func (p *TracerProxy) SetTransientState(addr common.Address, key common.Hash, value common.Hash) {
	p.ctx.WriteValueOp(SetTransientStateID, &addr, &key, &value)
	p.db.SetTransientState(addr, key, value)
}

func (p *TracerProxy) GetTransientState(addr common.Address, key common.Hash) common.Hash {
	p.ctx.WriteKeyOp(GetTransientStateID, &addr, &key)
	return p.db.GetState(addr, key)
}

func (p *TracerProxy) SelfDestruct(address common.Address) uint256.Int {
	p.ctx.WriteAddressOp(SelfDestructID, &address)
	return p.db.SelfDestruct(address)
}

func (p *TracerProxy) HasSelfDestructed(address common.Address) bool {
	p.ctx.WriteAddressOp(HasSelfDestructedID, &address)
	return p.db.HasSelfDestructed(address)
}

func (p *TracerProxy) Exist(address common.Address) bool {
	p.ctx.WriteAddressOp(ExistID, &address)
	return p.db.Exist(address)
}

func (p *TracerProxy) Empty(address common.Address) bool {
	p.ctx.WriteAddressOp(EmptyID, &address)
	return p.db.Empty(address)
}

func (p *TracerProxy) Prepare(rules params.Rules, sender, coinbase common.Address, dest *common.Address, precompiles []common.Address, txAccesses types.AccessList) {
	p.db.Prepare(rules, sender, coinbase, dest, precompiles, txAccesses)
}

func (p *TracerProxy) AddAddressToAccessList(address common.Address) {
	p.db.AddAddressToAccessList(address)
}

func (p *TracerProxy) AddressInAccessList(address common.Address) bool {
	return p.db.AddressInAccessList(address)
}

func (p *TracerProxy) SlotInAccessList(address common.Address, slot common.Hash) (bool, bool) {
	return p.db.SlotInAccessList(address, slot)
}

func (p *TracerProxy) AddSlotToAccessList(address common.Address, slot common.Hash) {
	p.db.AddSlotToAccessList(address, slot)
}

func (p *TracerProxy) RevertToSnapshot(snapshot int) {
	p.ctx.WriteOp(RevertToSnapshotID)
	p.db.RevertToSnapshot(snapshot)
}

func (p *TracerProxy) Snapshot() int {
	p.ctx.WriteOp(SnapshotID)
	return snapshot
}

func (p *TracerProxy) AddLog(log *types.Log) {
	p.db.AddLog(log)
}

func (p *TracerProxy) GetLogs(hash common.Hash, block uint64, blockHash common.Hash, blkTimestamp uint64) []*types.Log {
	return p.db.GetLogs(hash, block, blockHash, blkTimestamp)
}

func (p *TracerProxy) PointCache() *utils.PointCache {
	return p.db.PointCache()
}

func (p *TracerProxy) Witness() *stateless.Witness {
	return p.db.Witness()
}

func (p *TracerProxy) AddPreimage(address common.Hash, image []byte) {
	p.db.AddPreimage(address, image)
}

func (p *TracerProxy) AccessEvents() *geth_state.AccessEvents {
	return p.db.AccessEvents()
}

func (p *TracerProxy) SetTxContext(thash common.Hash, ti int) {
	p.db.SetTxContext(thash, ti)
}

func (p *TracerProxy) Finalise(deleteEmptyObjects bool) {
	p.db.Finalise(deleteEmptyObjects)
}

func (p *TracerProxy) IntermediateRoot(deleteEmptyObjects bool) common.Hash {
	return p.db.IntermediateRoot(deleteEmptyObjects)
}

func (p *TracerProxy) Commit(block uint64, deleteEmptyObjects bool) (common.Hash, error) {
	return p.db.Commit(block, deleteEmptyObjects)
}

func (p *TracerProxy) GetHash() (common.Hash, error) {
	return p.db.GetHash()
}

func (p *TracerProxy) Error() error {
	return p.db.Error()
}

func (p *TracerProxy) GetSubstatePostAlloc() txctx.WorldState {
	return p.db.GetSubstatePostAlloc()
}

func (p *TracerProxy) PrepareSubstate(substate txctx.WorldState, block uint64) {
	p.db.PrepareSubstate(substate, block)
}

func (p *TracerProxy) BeginTransaction(number uint32) error {
	p.ctx.WriteOp(BeginTransactionID)
	if err := p.db.BeginTransaction(number); err != nil {
		return err
	}
	return nil
}

func (p *TracerProxy) EndTransaction() error {
	p.ctx.WriteOp(EndTransactionID)
	if err := p.db.EndTransaction(); err != nil {
		return err
	}
	return nil
}

func (p *TracerProxy) BeginBlock(number uint64) error {
	p.ctx.WriteOp(BeginBlockID)
	return p.db.BeginBlock(number)
}

func (p *TracerProxy) EndBlock() error {
	p.ctx.WriteOp(EndBlockID)
	return p.db.EndBlock()
}

func (p *TracerProxy) BeginSyncPeriod(number uint64) {
	p.ctx.WriteOp(BeginSyncPeriodID)
	p.db.BeginSyncPeriod(number)
}

func (p *TracerProxy) EndSyncPeriod() {
	p.ctx.WriteOp(EndSyncPeriodID)
	p.db.EndSyncPeriod()
}

func (p *TracerProxy) Close() error {
	return p.db.Close()
}

func (p *TracerProxy) StartBulkLoad(uint64) (state.BulkLoad, error) {
	panic("StartBulkLoad not supported by TracerProxy")
}

func (p *TracerProxy) GetMemoryUsage() *state.MemoryUsage {
	return p.db.GetMemoryUsage()
}

func (p *TracerProxy) GetArchiveState(block uint64) (state.NonCommittableStateDB, error) {
	return p.db.GetArchiveState(block)
}

func (p *TracerProxy) GetArchiveBlockHeight() (uint64, bool, error) {
	return p.db.GetArchiveBlockHeight()
}

func (p *TracerProxy) GetShadowDB() state.StateDB {
	return p.db.GetShadowDB()
}

func (p *TracerProxy) CreateContract(addr common.Address) {
	p.ctx.WriteAddressOp(CreateContractID, addr)
	p.db.CreateContract(addr)
}

func (p *TracerProxy) SelfDestruct6780(addr common.Address) (uint256.Int, bool) {
	p.ctx.WriteAddressOp(SelfDestruct6780ID, addr)
	return p.db.SelfDestruct6780(addr)
}

func (p *TracerProxy) GetStorageRoot(addr common.Address) common.Hash {
	p.ctx.WriteAddressOp(CreateContractID, addr)
	return p.db.GetStorageRoot(addr)
}
