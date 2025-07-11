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

package tracer

import (
	"errors"
	"github.com/0xsoniclabs/aida/state"
	"github.com/0xsoniclabs/aida/txcontext"
	"github.com/Fantom-foundation/lachesis-base/common/bigendian"
	"github.com/ethereum/go-ethereum/common"
	geth_state "github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/stateless"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/trie/utils"
	"github.com/holiman/uint256"
)

// Proxy data structure for writing StateDB operations
type Proxy struct {
	db       state.StateDB    // StateDB object
	ctx      *ArgumentContext // context that keeps track of the argument history
	writeErr error
}

// NewTracerProxy creates a new StateDB proxy for recording and writing events.
func NewTracerProxy(db state.StateDB, ctx *ArgumentContext) *Proxy {
	return &Proxy{
		db:  db,
		ctx: ctx,
	}
}

func (p *Proxy) CreateAccount(address common.Address) {
	p.writeErr = errors.Join(p.writeErr, p.ctx.WriteAddressOp(CreateAccountID, &address, []byte{}))
	p.db.CreateAccount(address)
}

func (p *Proxy) SubBalance(address common.Address, amount *uint256.Int, reason tracing.BalanceChangeReason) uint256.Int {
	data := append(amount.Bytes(), byte(reason))
	p.writeErr = errors.Join(p.writeErr, p.ctx.WriteAddressOp(SubBalanceID, &address, data))
	return p.db.SubBalance(address, amount, reason)
}

func (p *Proxy) AddBalance(address common.Address, amount *uint256.Int, reason tracing.BalanceChangeReason) uint256.Int {
	data := append(amount.Bytes(), byte(reason))
	p.writeErr = errors.Join(p.writeErr, p.ctx.WriteAddressOp(AddBalanceID, &address, data))
	return p.db.AddBalance(address, amount, reason)
}

func (p *Proxy) GetBalance(address common.Address) *uint256.Int {
	p.writeErr = errors.Join(p.writeErr, p.ctx.WriteAddressOp(GetBalanceID, &address, []byte{}))
	return p.db.GetBalance(address)
}

func (p *Proxy) GetNonce(address common.Address) uint64 {
	p.writeErr = errors.Join(p.writeErr, p.ctx.WriteAddressOp(GetNonceID, &address, []byte{}))
	return p.db.GetNonce(address)
}

func (p *Proxy) SetNonce(address common.Address, nonce uint64, reason tracing.NonceChangeReason) {
	// TBD: find an encoding for nonce and reason in the form of a byte sequence
	data := []byte{}
	p.writeErr = errors.Join(p.writeErr, p.ctx.WriteAddressOp(SetNonceID, &address, data))
	p.db.SetNonce(address, nonce, reason)
}

func (p *Proxy) GetCodeHash(address common.Address) common.Hash {
	p.writeErr = errors.Join(p.writeErr, p.ctx.WriteAddressOp(GetCodeHashID, &address, []byte{}))
	return p.db.GetCodeHash(address)
}

func (p *Proxy) GetCode(address common.Address) []byte {
	p.writeErr = errors.Join(p.writeErr, p.ctx.WriteAddressOp(GetCodeID, &address, []byte{}))
	return p.db.GetCode(address)
}

func (p *Proxy) SetCode(address common.Address, code []byte) []byte {
	p.writeErr = errors.Join(p.writeErr, p.ctx.WriteAddressOp(SetCodeID, &address, code))
	return p.db.SetCode(address, code)
}

func (p *Proxy) GetCodeSize(address common.Address) int {
	p.writeErr = errors.Join(p.writeErr, p.ctx.WriteAddressOp(GetCodeSizeID, &address, []byte{}))
	return p.db.GetCodeSize(address)
}

func (p *Proxy) AddRefund(gas uint64) {
	data := bigendian.Uint64ToBytes(gas)
	p.writeErr = errors.Join(p.writeErr, p.ctx.WriteOp(GetCodeSizeID, data))
	p.db.AddRefund(gas)
}

func (p *Proxy) SubRefund(gas uint64) {
	data := bigendian.Uint64ToBytes(gas)
	p.writeErr = errors.Join(p.writeErr, p.ctx.WriteOp(SubRefundID, data))
	p.db.SubRefund(gas)
}

func (p *Proxy) GetRefund() uint64 {
	p.writeErr = errors.Join(p.writeErr, p.ctx.WriteOp(GetRefundID, []byte{}))
	return p.db.GetRefund()
}

func (p *Proxy) GetCommittedState(address common.Address, key common.Hash) common.Hash {
	p.writeErr = errors.Join(p.writeErr, p.ctx.WriteKeyOp(GetCommittedStateID, &address, &key, []byte{}))
	return p.db.GetCommittedState(address, key)
}

func (p *Proxy) GetState(address common.Address, key common.Hash) common.Hash {
	p.writeErr = errors.Join(p.writeErr, p.ctx.WriteKeyOp(GetStateID, &address, &key, []byte{}))
	return p.db.GetState(address, key)
}

func (p *Proxy) SetState(address common.Address, key common.Hash, value common.Hash) common.Hash {
	p.writeErr = errors.Join(p.writeErr, p.ctx.WriteValueOp(SetStateID, &address, &key, &value))
	return p.db.SetState(address, key, value)
}
func (p *Proxy) SetTransientState(addr common.Address, key common.Hash, value common.Hash) {
	p.writeErr = errors.Join(p.writeErr, p.ctx.WriteValueOp(SetTransientStateID, &addr, &key, &value))
	p.db.SetTransientState(addr, key, value)
}

func (p *Proxy) GetTransientState(addr common.Address, key common.Hash) common.Hash {
	p.writeErr = errors.Join(p.writeErr, p.ctx.WriteKeyOp(GetTransientStateID, &addr, &key, []byte{}))
	return p.db.GetState(addr, key)
}

func (p *Proxy) SelfDestruct(address common.Address) uint256.Int {
	p.writeErr = errors.Join(p.writeErr, p.ctx.WriteAddressOp(SelfDestructID, &address, []byte{}))
	return p.db.SelfDestruct(address)
}

func (p *Proxy) HasSelfDestructed(address common.Address) bool {
	p.writeErr = errors.Join(p.writeErr, p.ctx.WriteAddressOp(HasSelfDestructedID, &address, []byte{}))
	return p.db.HasSelfDestructed(address)
}

func (p *Proxy) Exist(address common.Address) bool {
	p.writeErr = errors.Join(p.writeErr, p.ctx.WriteAddressOp(ExistID, &address, []byte{}))
	return p.db.Exist(address)
}

func (p *Proxy) Empty(address common.Address) bool {
	p.writeErr = errors.Join(p.writeErr, p.ctx.WriteAddressOp(EmptyID, &address, []byte{}))
	return p.db.Empty(address)
}

func (p *Proxy) Prepare(rules params.Rules, sender, coinbase common.Address, dest *common.Address, precompiles []common.Address, txAccesses types.AccessList) {
	p.db.Prepare(rules, sender, coinbase, dest, precompiles, txAccesses)
}

func (p *Proxy) AddAddressToAccessList(address common.Address) {
	p.db.AddAddressToAccessList(address)
}

func (p *Proxy) AddressInAccessList(address common.Address) bool {
	return p.db.AddressInAccessList(address)
}

func (p *Proxy) SlotInAccessList(address common.Address, slot common.Hash) (bool, bool) {
	return p.db.SlotInAccessList(address, slot)
}

func (p *Proxy) AddSlotToAccessList(address common.Address, slot common.Hash) {
	p.db.AddSlotToAccessList(address, slot)
}

func (p *Proxy) RevertToSnapshot(snapshot int) {
	data := bigendian.Uint32ToBytes(uint32(snapshot))
	p.writeErr = errors.Join(p.writeErr, p.ctx.WriteOp(RevertToSnapshotID, data))
	p.db.RevertToSnapshot(snapshot)
}

func (p *Proxy) Snapshot() int {
	p.writeErr = errors.Join(p.writeErr, p.ctx.WriteOp(SnapshotID, []byte{}))
	return p.db.Snapshot()
}

func (p *Proxy) AddLog(log *types.Log) {
	p.db.AddLog(log)
}

func (p *Proxy) GetLogs(hash common.Hash, block uint64, blockHash common.Hash, blkTimestamp uint64) []*types.Log {
	return p.db.GetLogs(hash, block, blockHash, blkTimestamp)
}

func (p *Proxy) PointCache() *utils.PointCache {
	return p.db.PointCache()
}

func (p *Proxy) Witness() *stateless.Witness {
	return p.db.Witness()
}

func (p *Proxy) AddPreimage(address common.Hash, image []byte) {
	p.db.AddPreimage(address, image)
}

func (p *Proxy) AccessEvents() *geth_state.AccessEvents {
	return p.db.AccessEvents()
}

func (p *Proxy) SetTxContext(hash common.Hash, ti int) {
	p.db.SetTxContext(hash, ti)
}

func (p *Proxy) Finalise(deleteEmptyObjects bool) {
	p.db.Finalise(deleteEmptyObjects)
}

func (p *Proxy) IntermediateRoot(deleteEmptyObjects bool) common.Hash {
	return p.db.IntermediateRoot(deleteEmptyObjects)
}

func (p *Proxy) Commit(block uint64, deleteEmptyObjects bool) (common.Hash, error) {
	return p.db.Commit(block, deleteEmptyObjects)
}

func (p *Proxy) GetHash() (common.Hash, error) {
	return p.db.GetHash()
}

func (p *Proxy) Error() error {
	return errors.Join(errors.Unwrap(p.writeErr), p.db.Error())
}

func (p *Proxy) GetSubstatePostAlloc() txcontext.WorldState {
	return p.db.GetSubstatePostAlloc()
}

func (p *Proxy) PrepareSubstate(substate txcontext.WorldState, block uint64) {
	p.db.PrepareSubstate(substate, block)
}

func (p *Proxy) BeginTransaction(number uint32) error {
	data := bigendian.Uint32ToBytes(number)
	p.writeErr = errors.Join(p.writeErr, p.ctx.WriteOp(BeginTransactionID, data))
	if err := p.db.BeginTransaction(number); err != nil {
		return err
	}
	return nil
}

func (p *Proxy) EndTransaction() error {
	p.writeErr = errors.Join(p.writeErr, p.ctx.WriteOp(EndTransactionID, []byte{}))
	if err := p.db.EndTransaction(); err != nil {
		return err
	}
	return nil
}

func (p *Proxy) BeginBlock(number uint64) error {
	data := bigendian.Uint64ToBytes(number)
	p.writeErr = errors.Join(p.writeErr, p.ctx.WriteOp(BeginBlockID, data))
	return p.db.BeginBlock(number)
}

func (p *Proxy) EndBlock() error {
	p.writeErr = errors.Join(p.writeErr, p.ctx.WriteOp(EndBlockID, []byte{}))
	return p.db.EndBlock()
}

func (p *Proxy) BeginSyncPeriod(number uint64) {
	data := bigendian.Uint64ToBytes(number)
	p.writeErr = errors.Join(p.writeErr, p.ctx.WriteOp(BeginSyncPeriodID, data))
	p.db.BeginSyncPeriod(number)
}

func (p *Proxy) EndSyncPeriod() {
	p.writeErr = errors.Join(p.writeErr, p.ctx.WriteOp(EndSyncPeriodID, []byte{}))
	p.db.EndSyncPeriod()
}

func (p *Proxy) Close() error {
	return errors.Join(p.ctx.Close(), p.db.Close())
}

func (p *Proxy) StartBulkLoad(uint64) (state.BulkLoad, error) {
	panic("StartBulkLoad not supported by TracerProxy")
}

func (p *Proxy) GetMemoryUsage() *state.MemoryUsage {
	return p.db.GetMemoryUsage()
}

func (p *Proxy) GetArchiveState(block uint64) (state.NonCommittableStateDB, error) {
	return p.db.GetArchiveState(block)
}

func (p *Proxy) GetArchiveBlockHeight() (uint64, bool, error) {
	return p.db.GetArchiveBlockHeight()
}

func (p *Proxy) GetShadowDB() state.StateDB {
	return p.db.GetShadowDB()
}

func (p *Proxy) CreateContract(addr common.Address) {
	p.writeErr = errors.Join(p.writeErr, p.ctx.WriteAddressOp(CreateContractID, &addr, []byte{}))
	p.db.CreateContract(addr)
}

func (p *Proxy) SelfDestruct6780(addr common.Address) (uint256.Int, bool) {
	p.writeErr = errors.Join(p.writeErr, p.ctx.WriteAddressOp(SelfDestruct6780ID, &addr, []byte{}))
	return p.db.SelfDestruct6780(addr)
}

func (p *Proxy) GetStorageRoot(addr common.Address) common.Hash {
	p.writeErr = errors.Join(p.writeErr, p.ctx.WriteAddressOp(CreateContractID, &addr, []byte{}))
	return p.db.GetStorageRoot(addr)
}
