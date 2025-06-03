package statedb

import (
	"fmt"
	"github.com/0xsoniclabs/aida/executor"
	"github.com/0xsoniclabs/aida/executor/extension"
	"github.com/0xsoniclabs/aida/txcontext"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/sonic/evmcore"
	"github.com/ethereum/go-ethereum/core/vm"
	"math/big"
)

// NewParentBlockHashProcessor creates a new instance of parent block hash processor which saves the
// parent block hash in the blockchain. This is required for Prague fork and later (https://eips.ethereum.org/EIPS/eip-2935).
func NewParentBlockHashProcessor(cfg *utils.Config) executor.Extension[txcontext.TxContext] {
	return &parentBlockHashProcessor{
		cfg:          cfg,
		NilExtension: extension.NilExtension[txcontext.TxContext]{},
	}
}

type parentBlockHashProcessor struct {
	cfg          *utils.Config
	hashProvider utils.StateHashProvider
	extension.NilExtension[txcontext.TxContext]
}

func (p *parentBlockHashProcessor) PreRun(_ executor.State[txcontext.TxContext], ctx *executor.Context) error {
	p.hashProvider = utils.MakeStateHashProvider(ctx.AidaDb)
	return nil
}

// PreBlock processes parent block hash.
func (p *parentBlockHashProcessor) PreBlock(state executor.State[txcontext.TxContext], ctx *executor.Context) error {
	// We are saving historic block hashes, first block must be skipped because
	// there is no history at this point
	if uint64(state.Block) == utils.KeywordBlocks[p.cfg.ChainID]["first"] {
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

	prevBlockHash, err := p.hashProvider.GetStateHash(state.Block - 1)
	if err != nil {
		return fmt.Errorf("cannot get previous block hash: %w", err)
	}

	var hashError error
	blockCtx := utils.PrepareBlockCtx(inputEnv, &hashError)
	evm := vm.NewEVM(*blockCtx, ctx.State, chainCfg, p.cfg.VmCfg)
	evmcore.ProcessParentBlockHash(
		prevBlockHash,
		evm,
	)

	if hashError != nil {
		return fmt.Errorf("hash error while processing parent block hash: %v", err)
	}
	return nil
}
