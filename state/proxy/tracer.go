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
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"github.com/0xsoniclabs/aida/state"
	"github.com/0xsoniclabs/aida/tracer"
	"github.com/0xsoniclabs/aida/txcontext"
	"github.com/Fantom-foundation/lachesis-base/common/bigendian"
	"github.com/ethereum/go-ethereum/common"
	geth_state "github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/stateless"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	geth_utils "github.com/ethereum/go-ethereum/trie/utils"
	"github.com/holiman/uint256"
)

// TracerProxy data structure for writing StateDB operations
type TracerProxy struct {
	db       state.StateDB
	ctx      tracer.Context
	writeErr error
}

// NewTracerProxy creates a new StateDB proxy for recording and writing events.
func NewTracerProxy(db state.StateDB, ctx tracer.Context) *TracerProxy {
	return &TracerProxy{
		db:  db,
		ctx: ctx,
	}
}

func (p *TracerProxy) CreateAccount(address common.Address) {
	p.writeErr = errors.Join(p.writeErr, p.ctx.WriteAddressOp(tracer.CreateAccountID, &address, []byte{}))
	p.db.CreateAccount(address)
}

func (p *TracerProxy) SubBalance(address common.Address, amount *uint256.Int, reason tracing.BalanceChangeReason) uint256.Int {
	if address.Cmp(common.HexToAddress("0x6fc240cb0cb82a0323b92a8053f7b6d1329f4075")) == 0 {
		b := p.db.GetBalance(address)
		fmt.Println(b.Uint64())
	}

	size := uint32(amount.ByteLen())
	data := append(bigendian.Uint32ToBytes(size), amount.Bytes()...)
	data = append(data, byte(reason))
	p.writeErr = errors.Join(p.writeErr, p.ctx.WriteAddressOp(tracer.SubBalanceID, &address, data))
	return p.db.SubBalance(address, amount, reason)
}

func (p *TracerProxy) AddBalance(address common.Address, amount *uint256.Int, reason tracing.BalanceChangeReason) uint256.Int {
	size := uint32(amount.ByteLen())
	data := append(bigendian.Uint32ToBytes(size), amount.Bytes()...)
	data = append(data, byte(reason))
	p.writeErr = errors.Join(p.writeErr, p.ctx.WriteAddressOp(tracer.AddBalanceID, &address, data))
	return p.db.AddBalance(address, amount, reason)
}

func (p *TracerProxy) GetBalance(address common.Address) *uint256.Int {
	p.writeErr = errors.Join(p.writeErr, p.ctx.WriteAddressOp(tracer.GetBalanceID, &address, []byte{}))
	return p.db.GetBalance(address)
}

func (p *TracerProxy) GetNonce(address common.Address) uint64 {
	p.writeErr = errors.Join(p.writeErr, p.ctx.WriteAddressOp(tracer.GetNonceID, &address, []byte{}))
	return p.db.GetNonce(address)
}

func (p *TracerProxy) SetNonce(address common.Address, nonce uint64, reason tracing.NonceChangeReason) {
	data := append(bigendian.Uint64ToBytes(nonce), byte(reason))
	p.writeErr = errors.Join(p.writeErr, p.ctx.WriteAddressOp(tracer.SetNonceID, &address, data))
	p.db.SetNonce(address, nonce, reason)
}

func (p *TracerProxy) GetCodeHash(address common.Address) common.Hash {
	p.writeErr = errors.Join(p.writeErr, p.ctx.WriteAddressOp(tracer.GetCodeHashID, &address, []byte{}))
	return p.db.GetCodeHash(address)
}

func (p *TracerProxy) GetCode(address common.Address) []byte {
	p.writeErr = errors.Join(p.writeErr, p.ctx.WriteAddressOp(tracer.GetCodeID, &address, []byte{}))
	return p.db.GetCode(address)
}

func (p *TracerProxy) SetCode(address common.Address, code []byte) []byte {
	// As of Osaka fork the max code size is 24_576 bytes
	// Max uint32 is 4_294_967_295
	size := uint32(len(code))
	encodedSize := bigendian.Uint32ToBytes(size)
	data := append(encodedSize, code...)
	p.writeErr = errors.Join(p.writeErr, p.ctx.WriteAddressOp(tracer.SetCodeID, &address, data))
	return p.db.SetCode(address, code)
}

func (p *TracerProxy) GetCodeSize(address common.Address) int {
	p.writeErr = errors.Join(p.writeErr, p.ctx.WriteAddressOp(tracer.GetCodeSizeID, &address, []byte{}))
	return p.db.GetCodeSize(address)
}

func (p *TracerProxy) AddRefund(gas uint64) {
	data := bigendian.Uint64ToBytes(gas)
	p.writeErr = errors.Join(p.writeErr, p.ctx.WriteOp(tracer.AddRefundID, data))
	p.db.AddRefund(gas)
}

func (p *TracerProxy) SubRefund(gas uint64) {
	data := bigendian.Uint64ToBytes(gas)
	p.writeErr = errors.Join(p.writeErr, p.ctx.WriteOp(tracer.SubRefundID, data))
	p.db.SubRefund(gas)
}

func (p *TracerProxy) GetRefund() uint64 {
	p.writeErr = errors.Join(p.writeErr, p.ctx.WriteOp(tracer.GetRefundID, []byte{}))
	return p.db.GetRefund()
}

func (p *TracerProxy) GetCommittedState(address common.Address, key common.Hash) common.Hash {
	p.writeErr = errors.Join(p.writeErr, p.ctx.WriteKeyOp(tracer.GetCommittedStateID, &address, &key, []byte{}))
	return p.db.GetCommittedState(address, key)
}

func (p *TracerProxy) GetState(address common.Address, key common.Hash) common.Hash {
	p.writeErr = errors.Join(p.writeErr, p.ctx.WriteKeyOp(tracer.GetStateID, &address, &key, []byte{}))
	return p.db.GetState(address, key)
}

func (p *TracerProxy) SetState(address common.Address, key common.Hash, value common.Hash) common.Hash {
	p.writeErr = errors.Join(p.writeErr, p.ctx.WriteValueOp(tracer.SetStateID, &address, &key, &value))
	return p.db.SetState(address, key, value)
}
func (p *TracerProxy) SetTransientState(addr common.Address, key common.Hash, value common.Hash) {
	p.writeErr = errors.Join(p.writeErr, p.ctx.WriteValueOp(tracer.SetTransientStateID, &addr, &key, &value))
	p.db.SetTransientState(addr, key, value)
}

func (p *TracerProxy) GetTransientState(addr common.Address, key common.Hash) common.Hash {
	p.writeErr = errors.Join(p.writeErr, p.ctx.WriteKeyOp(tracer.GetTransientStateID, &addr, &key, []byte{}))
	return p.db.GetTransientState(addr, key)
}

func (p *TracerProxy) SelfDestruct(address common.Address) uint256.Int {
	p.writeErr = errors.Join(p.writeErr, p.ctx.WriteAddressOp(tracer.SelfDestructID, &address, []byte{}))
	return p.db.SelfDestruct(address)
}

func (p *TracerProxy) HasSelfDestructed(address common.Address) bool {
	p.writeErr = errors.Join(p.writeErr, p.ctx.WriteAddressOp(tracer.HasSelfDestructedID, &address, []byte{}))
	return p.db.HasSelfDestructed(address)
}

func (p *TracerProxy) Exist(address common.Address) bool {
	p.writeErr = errors.Join(p.writeErr, p.ctx.WriteAddressOp(tracer.ExistID, &address, []byte{}))
	return p.db.Exist(address)
}

func (p *TracerProxy) Empty(address common.Address) bool {
	p.writeErr = errors.Join(p.writeErr, p.ctx.WriteAddressOp(tracer.EmptyID, &address, []byte{}))
	return p.db.Empty(address)
}

func (p *TracerProxy) Prepare(rules params.Rules, sender, coinbase common.Address, dest *common.Address, precompiles []common.Address, txAccesses types.AccessList) {
	buf := bytes.Buffer{}
	enc := gob.NewEncoder(&buf)

	err := enc.Encode(rules)
	if err != nil {
		p.writeErr = errors.Join(p.writeErr, fmt.Errorf("cannot encode rules: %w", err))
		return
	}
	encodedSize := bigendian.Uint32ToBytes(uint32(buf.Len()))
	// Append size of rules and rules
	data := append(encodedSize, buf.Bytes()...)
	// Append sender
	data = append(data, sender.Bytes()...)
	// Append coinbase
	data = append(data, coinbase.Bytes()...)
	// Append dest
	if dest != nil {
		// signal that dest is present
		data = append(data, byte(1))
		data = append(data, dest.Bytes()...)
	} else {
		// signal that dest is not present
		data = append(data, byte(0))
	}
	// Append number of addrs in precompiles
	numAddrs := uint16(len(precompiles))
	data = append(data, bigendian.Uint16ToBytes(numAddrs)...)
	// Append addr one by one
	for _, addr := range precompiles {
		data = append(data, addr.Bytes()...)
	}
	// Append size of txAccesses and txAccesses
	if len(txAccesses) == 0 {
		// If txAccesses is empty, we append a zero size
		data = append(data, bigendian.Uint32ToBytes(0)...)
	} else {
		// RESET BUFFER TO REUSE IT!!!
		buf.Reset()
		err = enc.Encode(txAccesses)
		if err != nil {
			p.writeErr = errors.Join(p.writeErr, fmt.Errorf("cannot encode txAccesses: %w", err))
			return
		}
		encodedSize = bigendian.Uint32ToBytes(uint32(buf.Len()))
		data = append(data, encodedSize...)
		data = append(data, buf.Bytes()...)
	}
	p.writeErr = errors.Join(p.writeErr, p.ctx.WriteOp(tracer.PrepareID, data))
	p.db.Prepare(rules, sender, coinbase, dest, precompiles, txAccesses)
}

func (p *TracerProxy) AddAddressToAccessList(address common.Address) {
	p.writeErr = errors.Join(p.writeErr, p.ctx.WriteAddressOp(tracer.AddAddressToAccessListID, &address, []byte{}))
	p.db.AddAddressToAccessList(address)
}

func (p *TracerProxy) AddressInAccessList(address common.Address) bool {
	p.writeErr = errors.Join(p.writeErr, p.ctx.WriteAddressOp(tracer.AddressInAccessListID, &address, []byte{}))
	return p.db.AddressInAccessList(address)
}

func (p *TracerProxy) SlotInAccessList(address common.Address, slot common.Hash) (bool, bool) {
	p.writeErr = errors.Join(p.writeErr, p.ctx.WriteKeyOp(tracer.SlotInAccessListID, &address, &slot, []byte{}))
	return p.db.SlotInAccessList(address, slot)
}

func (p *TracerProxy) AddSlotToAccessList(address common.Address, slot common.Hash) {
	p.writeErr = errors.Join(p.writeErr, p.ctx.WriteKeyOp(tracer.AddSlotToAccessListID, &address, &slot, []byte{}))
	p.db.AddSlotToAccessList(address, slot)
}

func (p *TracerProxy) RevertToSnapshot(snapshot int) {
	data := bigendian.Uint32ToBytes(uint32(snapshot))
	p.writeErr = errors.Join(p.writeErr, p.ctx.WriteOp(tracer.RevertToSnapshotID, data))
	p.db.RevertToSnapshot(snapshot)
}

func (p *TracerProxy) Snapshot() int {
	p.writeErr = errors.Join(p.writeErr, p.ctx.WriteOp(tracer.SnapshotID, []byte{}))
	return p.db.Snapshot()
}

func (p *TracerProxy) AddLog(log *types.Log) {
	buf := bytes.Buffer{}
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(log)
	if err != nil {
		p.writeErr = errors.Join(p.writeErr, fmt.Errorf("cannot encode log: %w", err))
		return
	}
	size := bigendian.Uint32ToBytes(uint32(buf.Len()))
	data := append(size, buf.Bytes()...)
	p.writeErr = errors.Join(p.writeErr, p.ctx.WriteOp(tracer.AddLogID, data))
	p.db.AddLog(log)
}

func (p *TracerProxy) GetLogs(hash common.Hash, block uint64, blockHash common.Hash, blkTimestamp uint64) []*types.Log {
	data := append(hash.Bytes(), bigendian.Uint64ToBytes(block)...)
	data = append(data, blockHash.Bytes()...)
	data = append(data, bigendian.Uint64ToBytes(blkTimestamp)...)
	p.writeErr = errors.Join(p.writeErr, p.ctx.WriteOp(tracer.GetLogsID, data))
	return p.db.GetLogs(hash, block, blockHash, blkTimestamp)
}

func (p *TracerProxy) PointCache() *geth_utils.PointCache {
	p.writeErr = errors.Join(p.writeErr, p.ctx.WriteOp(tracer.PointCacheID, []byte{}))
	return p.db.PointCache()
}

func (p *TracerProxy) Witness() *stateless.Witness {
	p.writeErr = errors.Join(p.writeErr, p.ctx.WriteOp(tracer.WitnessID, []byte{}))
	return p.db.Witness()
}

func (p *TracerProxy) AddPreimage(hash common.Hash, image []byte) {
	size := len(image)
	encodedSize := bigendian.Uint32ToBytes(uint32(size))
	data := append(hash.Bytes(), encodedSize...)
	if size > 0 {
		data = append(data, image...)
	}
	p.writeErr = errors.Join(p.writeErr, p.ctx.WriteOp(tracer.AddPreimageID, data))
	p.db.AddPreimage(hash, image)
}

func (p *TracerProxy) AccessEvents() *geth_state.AccessEvents {
	p.writeErr = errors.Join(p.writeErr, p.ctx.WriteOp(tracer.AccessEventsID, []byte{}))
	return p.db.AccessEvents()
}

func (p *TracerProxy) SetTxContext(hash common.Hash, ti int) {
	data := append(hash.Bytes(), bigendian.Uint32ToBytes(uint32(ti))...)
	p.writeErr = errors.Join(p.writeErr, p.ctx.WriteOp(tracer.SetTxContextID, data))
	p.db.SetTxContext(hash, ti)
}

func (p *TracerProxy) Finalise(deleteEmptyObjects bool) {
	data := []byte{0}
	if deleteEmptyObjects {
		data[0] = 1
	}
	p.writeErr = errors.Join(p.writeErr, p.ctx.WriteOp(tracer.FinaliseID, data))
	p.db.Finalise(deleteEmptyObjects)
}

func (p *TracerProxy) IntermediateRoot(deleteEmptyObjects bool) common.Hash {
	data := []byte{0}
	if deleteEmptyObjects {
		data[0] = 1
	}
	p.writeErr = errors.Join(p.writeErr, p.ctx.WriteOp(tracer.IntermediateRootID, data))
	return p.db.IntermediateRoot(deleteEmptyObjects)
}

func (p *TracerProxy) Commit(block uint64, deleteEmptyObjects bool) (common.Hash, error) {
	data := []byte{0}
	if deleteEmptyObjects {
		data[0] = 1
	}
	data = append(data, bigendian.Uint64ToBytes(block)...)
	p.writeErr = errors.Join(p.writeErr, p.ctx.WriteOp(tracer.CommitID, data))
	return p.db.Commit(block, deleteEmptyObjects)
}

func (p *TracerProxy) GetHash() (common.Hash, error) {
	p.writeErr = errors.Join(p.writeErr, p.ctx.WriteOp(tracer.GetHashID, []byte{}))
	return p.db.GetHash()
}

func (p *TracerProxy) Error() error {
	return errors.Join(p.writeErr, p.db.Error())
}

func (p *TracerProxy) GetSubstatePostAlloc() txcontext.WorldState {
	p.writeErr = errors.Join(p.writeErr, p.ctx.WriteOp(tracer.GetSubstatePostAllocID, []byte{}))
	return p.db.GetSubstatePostAlloc()
}

func (p *TracerProxy) PrepareSubstate(substate txcontext.WorldState, block uint64) {
	var data []byte
	if substate.Len() > 0 {
		buf := bytes.Buffer{}
		enc := gob.NewEncoder(&buf)
		err := enc.Encode(substate)
		if err != nil {
			p.writeErr = errors.Join(p.writeErr, fmt.Errorf("cannot encode substate: %w", err))
			return
		}
		data = append(bigendian.Uint32ToBytes(uint32(buf.Len())), buf.Bytes()...)
	} else {
		data = bigendian.Uint32ToBytes(0) // empty substate
	}
	p.writeErr = errors.Join(p.writeErr, p.ctx.WriteOp(tracer.PrepareSubstateID, data))
	p.db.PrepareSubstate(substate, block)
}

func (p *TracerProxy) BeginTransaction(number uint32) error {
	data := bigendian.Uint32ToBytes(number)
	p.writeErr = errors.Join(p.writeErr, p.ctx.WriteOp(tracer.BeginTransactionID, data))
	return p.db.BeginTransaction(number)
}

func (p *TracerProxy) EndTransaction() error {
	p.writeErr = errors.Join(p.writeErr, p.ctx.WriteOp(tracer.EndTransactionID, []byte{}))
	return p.db.EndTransaction()
}

func (p *TracerProxy) BeginBlock(number uint64) error {
	data := bigendian.Uint64ToBytes(number)
	p.writeErr = errors.Join(p.writeErr, p.ctx.WriteOp(tracer.BeginBlockID, data))
	return p.db.BeginBlock(number)
}

func (p *TracerProxy) EndBlock() error {
	p.writeErr = errors.Join(p.writeErr, p.ctx.WriteOp(tracer.EndBlockID, []byte{}))
	return p.db.EndBlock()
}

func (p *TracerProxy) BeginSyncPeriod(number uint64) {
	data := bigendian.Uint64ToBytes(number)
	p.writeErr = errors.Join(p.writeErr, p.ctx.WriteOp(tracer.BeginSyncPeriodID, data))
	p.db.BeginSyncPeriod(number)
}

func (p *TracerProxy) EndSyncPeriod() {
	p.writeErr = errors.Join(p.writeErr, p.ctx.WriteOp(tracer.EndSyncPeriodID, []byte{}))
	p.db.EndSyncPeriod()
}

func (p *TracerProxy) Close() error {
	p.writeErr = errors.Join(p.writeErr, p.ctx.WriteOp(tracer.CloseID, []byte{}))
	return errors.Join(p.ctx.Close(), p.db.Close())
}

func (p *TracerProxy) StartBulkLoad(uint64) (state.BulkLoad, error) {
	panic("StartBulkLoad not supported by TracerProxy")
}

func (p *TracerProxy) GetMemoryUsage() *state.MemoryUsage {
	// ignored
	return p.db.GetMemoryUsage()
}

func (p *TracerProxy) GetArchiveState(block uint64) (state.NonCommittableStateDB, error) {
	p.writeErr = errors.Join(p.writeErr, p.ctx.WriteOp(tracer.GetArchiveStateID, bigendian.Uint64ToBytes(block)))
	return p.db.GetArchiveState(block)
}

func (p *TracerProxy) GetArchiveBlockHeight() (uint64, bool, error) {
	p.writeErr = errors.Join(p.writeErr, p.ctx.WriteOp(tracer.GetArchiveBlockHeightID, []byte{}))
	return p.db.GetArchiveBlockHeight()
}

func (p *TracerProxy) GetShadowDB() state.StateDB {
	// ignored
	return p.db.GetShadowDB()
}

func (p *TracerProxy) CreateContract(addr common.Address) {
	p.writeErr = errors.Join(p.writeErr, p.ctx.WriteAddressOp(tracer.CreateContractID, &addr, []byte{}))
	p.db.CreateContract(addr)
}

func (p *TracerProxy) SelfDestruct6780(addr common.Address) (uint256.Int, bool) {
	p.writeErr = errors.Join(p.writeErr, p.ctx.WriteAddressOp(tracer.SelfDestruct6780ID, &addr, []byte{}))
	return p.db.SelfDestruct6780(addr)
}

func (p *TracerProxy) GetStorageRoot(addr common.Address) common.Hash {
	p.writeErr = errors.Join(p.writeErr, p.ctx.WriteAddressOp(tracer.GetStorageRootID, &addr, []byte{}))
	return p.db.GetStorageRoot(addr)
}
