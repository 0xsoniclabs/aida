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

package statedb

//go:generate mockgen -source parent_block_hash_processor.go -destination mocks/parent_block_hash_processor_mock.go -package mocks

import (
	"fmt"
	"math/big"

	"github.com/0xsoniclabs/substate/db"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/params"

	"github.com/0xsoniclabs/aida/executor"
	"github.com/0xsoniclabs/aida/executor/extension"
	"github.com/0xsoniclabs/aida/state"
	"github.com/0xsoniclabs/aida/txcontext"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/sonic/evmcore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
)

// NewParentBlockHashProcessor creates a new instance of parent block hash processor which saves the
// parent block hash in the blockchain. This is required for Prague fork and later (https://eips.ethereum.org/EIPS/eip-2935).
func NewParentBlockHashProcessor(cfg *utils.Config) executor.Extension[txcontext.TxContext] {
	return &parentBlockHashProcessor{
		processor:    evmProcessor{},
		cfg:          cfg,
		NilExtension: extension.NilExtension[txcontext.TxContext]{},
	}
}

type parentBlockHashProcessor struct {
	hashProvider       db.HashProvider
	processor          iEvmProcessor
	cfg                *utils.Config
	lastProcessedBlock uint64 // Substate may skip blocks, so we need to track the last processed block to fill the gaps
	extension.NilExtension[txcontext.TxContext]
}

// iEvmProcessor is an interface that defines the method to process the parent block hash.
type iEvmProcessor interface {
	// ProcessParentBlockHash saves prevHash in the blockchain.
	ProcessParentBlockHash(common.Hash, *vm.EVM, state.StateDB) error
}

// evmProcessor is a wrapper around evmcore.ProcessParentBlockHash.
type evmProcessor struct{}

// ProcessParentBlockHash saves prevHash in the blockchain by calling the history storage contract.
// Copied from sonic codebase.
func (p evmProcessor) ProcessParentBlockHash(prevHash common.Hash, evm *vm.EVM, state state.StateDB) error {
	msg := &core.Message{
		From:      params.SystemAddress,
		GasLimit:  30_000_000,
		GasPrice:  common.Big0,
		GasFeeCap: common.Big0,
		GasTipCap: common.Big0,
		To:        &params.HistoryStorageAddress,
		Data:      prevHash.Bytes(),
	}

	state.AddAddressToAccessList(params.HistoryStorageAddress)
	txContext := evmcore.NewEVMTxContext(msg)
	if evm != nil {
		evm.SetTxContext(txContext)
		_, _, _ = evm.Call(msg.From, *msg.To, msg.Data, 30_000_000, common.U2560)
	}
	state.Finalise(true)
	return state.EndTransaction()
}

func (p *parentBlockHashProcessor) PreRun(_ executor.State[txcontext.TxContext], ctx *executor.Context) error {
	p.hashProvider = db.MakeHashProvider(ctx.AidaDb)
	// initialized the last processed block
	if p.cfg.First > 0 {
		p.lastProcessedBlock = p.cfg.First - 1
	}
	return nil
}

// PreBlock processes parent block hash. It loops from the previous processed
// block to the current block to fill any gaps caused by skipped blocks in substate.
func (p *parentBlockHashProcessor) PreBlock(state executor.State[txcontext.TxContext], ctx *executor.Context) error {
	defer func() {
		p.lastProcessedBlock = uint64(state.Block)
	}()

	// We are saving historic block hashes, first block must be skipped because
	// there is no history at this point
	firstBlock := utils.KeywordBlocks[p.cfg.ChainID]["first"]
	if uint64(state.Block) <= firstBlock {
		return nil
	}

	inputEnv := state.Data.GetBlockEnvironment()
	chainCfg, err := p.cfg.GetChainConfig(inputEnv.GetFork())
	if err != nil {
		return fmt.Errorf("cannot get chain config: %w", err)
	}

	if !chainCfg.IsPrague(new(big.Int).SetUint64(inputEnv.GetNumber()), inputEnv.GetTimestamp()) {
		return nil
	}

	startBlock := int(p.lastProcessedBlock) + 1

	for b := startBlock; b <= state.Block; b++ {
		prevBlockHash, err := p.hashProvider.GetBlockHash(b - 1)
		if err != nil {
			return fmt.Errorf("cannot get block hash for block %d: %w", b-1, err)
		}

		if err = ctx.State.BeginTransaction(utils.PseudoTx); err != nil {
			return fmt.Errorf("cannot begin transaction: %w", err)
		}

		var hashError error
		blockCtx := utils.PrepareBlockCtx(inputEnv, &hashError)
		blockCtx.BlockNumber = new(big.Int).SetUint64(uint64(b))
		evm := vm.NewEVM(*blockCtx, ctx.State, chainCfg, p.cfg.VmCfg)
		err = p.processor.ProcessParentBlockHash(common.Hash(prevBlockHash), evm, ctx.State)
		if err != nil {
			return err
		}
		if hashError != nil {
			return fmt.Errorf("hash error while processing parent block hash for block %d: %v", b, hashError)
		}
	}

	return nil
}
