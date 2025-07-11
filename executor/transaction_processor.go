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

package executor

import (
	"errors"
	"fmt"
	"math/big"
	"slices"
	"strings"
	"sync/atomic"

	"github.com/ethereum/go-ethereum/params"

	"github.com/0xsoniclabs/aida/ethtest"
	"github.com/ethereum/go-ethereum/core"
	"golang.org/x/exp/maps"

	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/state"
	"github.com/0xsoniclabs/aida/txcontext"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/tosca/go/tosca"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/holiman/uint256"

	_ "github.com/0xsoniclabs/tosca/go/processor/floria"
	_ "github.com/0xsoniclabs/tosca/go/processor/geth"
	_ "github.com/0xsoniclabs/tosca/go/processor/geth_eth"
	_ "github.com/0xsoniclabs/tosca/go/processor/opera"
)

// MakeLiveDbTxProcessor creates a executor.Processor which processes transaction into LIVE StateDb.
func MakeLiveDbTxProcessor(cfg *utils.Config) (*LiveDbTxProcessor, error) {
	processor, err := MakeTxProcessor(cfg)
	if err != nil {
		return nil, err
	}
	return &LiveDbTxProcessor{processor}, nil
}

type LiveDbTxProcessor struct {
	*TxProcessor
}

// Process transaction inside state into given LIVE StateDb
func (p *LiveDbTxProcessor) Process(state State[txcontext.TxContext], ctx *Context) error {
	var err error

	ctx.ExecutionResult, err = p.ProcessTransaction(ctx.State, state.Block, state.Transaction, state.Data)
	if err == nil {
		return nil
	}

	if !p.isErrFatal() {
		ctx.ErrorInput <- fmt.Errorf("live-db processor failed; %v", err)
		return nil
	}

	return err
}

// MakeArchiveDbTxProcessor creates a executor.Processor which processes transaction into ARCHIVE StateDb.
func MakeArchiveDbTxProcessor(cfg *utils.Config) (*ArchiveDbTxProcessor, error) {
	processor, err := MakeTxProcessor(cfg)
	if err != nil {
		return nil, err
	}
	return &ArchiveDbTxProcessor{processor}, nil
}

type ArchiveDbTxProcessor struct {
	*TxProcessor
}

// Process transaction inside state into given ARCHIVE StateDb
func (p *ArchiveDbTxProcessor) Process(state State[txcontext.TxContext], ctx *Context) error {
	var err error

	ctx.ExecutionResult, err = p.ProcessTransaction(ctx.Archive, state.Block, state.Transaction, state.Data)
	if err == nil {
		return nil
	}

	if !p.isErrFatal() {
		ctx.ErrorInput <- fmt.Errorf("archive-db processor failed; %v", err)
		return nil
	}

	return err
}

// MakeEthTestProcessor creates an executor.Processor which processes transaction created from ethereum test package.
func MakeEthTestProcessor(cfg *utils.Config) (*ethTestProcessor, error) {
	processor, err := MakeTxProcessor(cfg)
	if err != nil {
		return nil, err
	}
	return &ethTestProcessor{processor}, nil
}

type ethTestProcessor struct {
	*TxProcessor
}

// Process transaction inside state into given LIVE StateDb
func (p *ethTestProcessor) Process(state State[txcontext.TxContext], ctx *Context) error {
	// These checks need to be done before ApplyMessage is called to identify invalid transactions. Invalid
	// transactions are expected to be filtered out before running them (via ApplyMessage). If they
	// got processed, they would at least update the nonce of the transaction sender, thereby influencing
	// the resulting state hash. This would be detected as a failed test case.
	msg := state.Data.GetMessage()

	// Compute the maximum blob gas limit per block.
	maxBlobTransactions := 0
	switch fork := strings.ToLower(state.Data.GetBlockEnvironment().GetFork()); fork {
	case "istanbul":
	case "berlin":
	case "london":
	case "merge":
	case "paris":
	case "shanghai":
		maxBlobTransactions = 0
	case "cancun":
		maxBlobTransactions = 6
	case "prague":
		maxBlobTransactions = 9
	default:
		return fmt.Errorf("unknown fork: %s", fork)
	}
	maxBlobGasPerBlock := maxBlobTransactions * params.BlobTxBlobGasPerBlob

	// Make sure the block's blob gas limit is not exceeded.
	if len(msg.BlobHashes)*params.BlobTxBlobGasPerBlob > maxBlobGasPerBlock {
		ctx.ExecutionResult = newTransactionResult([]*types.Log{}, msg, nil, errors.New("blob gas exceeds maximum"), msg.From)
		return nil
	}

	txBytes := state.Data.(*ethtest.StateTestContext).GetTxBytes()

	if len(txBytes) != 0 {
		var ttx types.Transaction
		err := ttx.UnmarshalBinary(txBytes)
		if err != nil {
			ctx.ExecutionResult = newTransactionResult(
				[]*types.Log{},
				msg,
				nil,
				fmt.Errorf("cannot unmarshal tx-bytes: %w", err),
				msg.From,
			)
			return nil
		}

		chainCfg, err := p.cfg.GetChainConfig(state.Data.GetBlockEnvironment().GetFork())
		if err != nil {
			return err
		}
		if _, err = types.Sender(types.LatestSigner(chainCfg), &ttx); err != nil {
			ctx.ExecutionResult = newTransactionResult(
				[]*types.Log{},
				msg,
				nil,
				fmt.Errorf("cannot validate sender: %w", err),
				msg.From,
			)
			return nil
		}
	}

	// We ignore error in this case, because some tests require the processor to fail,
	// ethStateTestValidator decides whether error is fatal.
	ctx.ExecutionResult, _ = p.ProcessTransaction(ctx.State, state.Block, state.Transaction, state.Data)
	return nil
}

type TxProcessor struct {
	cfg       *utils.Config
	numErrors *atomic.Int32 // transactions can be processed in parallel, so this needs to be thread safe
	log       logger.Logger
	processor processor
}

func MakeTxProcessor(cfg *utils.Config) (*TxProcessor, error) {
	var processor processor
	switch strings.ToLower(cfg.EvmImpl) {
	case "", "opera", "ethereum":
		processor = makeAidaProcessor(cfg)
	default:
		interpreter, err := tosca.NewInterpreter(cfg.VmImpl)
		if err != nil {
			available := maps.Keys(tosca.GetAllRegisteredInterpreters())
			return nil, fmt.Errorf("failed to create interpreter %s, error %v, supported: %v", cfg.VmImpl, err, available)
		}
		evm := tosca.GetProcessor(cfg.EvmImpl, interpreter)
		if evm == nil {
			available := maps.Keys(tosca.GetAllRegisteredProcessorFactories())
			available = append(available, "opera", "ethereum")
			slices.Sort(available)
			return nil, fmt.Errorf("unknown EVM implementation: %s, supported: %v", cfg.EvmImpl, available)
		}

		processor = &toscaProcessor{
			processor: evm,
			cfg:       cfg,
			log:       logger.NewLogger(cfg.LogLevel, fmt.Sprintf("ToscaProcessor-%s-%s", cfg.EvmImpl, cfg.VmImpl)),
		}
	}

	return &TxProcessor{
		cfg:       cfg,
		numErrors: new(atomic.Int32),
		log:       logger.NewLogger(cfg.LogLevel, "TxProcessor"),
		processor: processor,
	}, nil
}

func (s *TxProcessor) isErrFatal() bool {
	if !s.cfg.ContinueOnFailure {
		return true
	}

	// check this first, so we don't have to access atomic value
	if s.cfg.MaxNumErrors <= 0 {
		return false
	}

	if s.numErrors.Load() < int32(s.cfg.MaxNumErrors) {
		s.numErrors.Add(1)
		return false
	}

	return true
}

func (s *TxProcessor) ProcessTransaction(db state.VmStateDB, block int, tx int, st txcontext.TxContext) (txcontext.Result, error) {
	if tx >= utils.PseudoTx {
		return s.processPseudoTx(st.GetOutputState(), db), nil
	}
	return s.processor.processRegularTx(db, block, tx, st)
}

type processor interface {
	processRegularTx(db state.VmStateDB, block int, tx int, st txcontext.TxContext) (transactionResult, error)
}

type aidaProcessor struct {
	cfg *utils.Config
	log logger.Logger
}

// for testing purposes
func makeAidaProcessor(cfg *utils.Config) *aidaProcessor {
	evmImpl := strings.ToLower(cfg.EvmImpl)
	return &aidaProcessor{
		cfg: cfg,
		log: logger.NewLogger(cfg.LogLevel, fmt.Sprintf("AidaProcessor(%s)", evmImpl)),
	}
}

// executionResult is a wrapper around ExecutionResult so both types from core and evmcore can be used.
type executionResult interface {
	Failed() bool
	Return() []byte
	GetGasUsed() uint64
	GetError() error
}

// messageResult is a basic implementation of execution result which
// contains data owned by ExecutionResult from both evmcore and core.
type messageResult struct {
	failed     bool
	returnData []byte
	gasUsed    uint64
	err        error
}

func (w messageResult) Failed() bool {
	return w.failed
}

func (w messageResult) Return() []byte {
	return w.returnData
}

func (w messageResult) GetGasUsed() uint64 {
	return w.gasUsed
}

func (w messageResult) GetError() error {
	return w.err
}

// processRegularTx executes VM on a chosen storage system.
func (s *aidaProcessor) processRegularTx(db state.VmStateDB, block int, tx int, st txcontext.TxContext) (res transactionResult, finalError error) {
	var (
		txHash    = common.HexToHash(fmt.Sprintf("0x%016d%016d", block, tx))
		inputEnv  = st.GetBlockEnvironment()
		msg       = st.GetMessage()
		hashError error
	)

	chainCfg, err := s.cfg.GetChainConfig(inputEnv.GetFork())
	// Return early if chain config cannot be created.
	if err != nil {
		return res, fmt.Errorf("cannot get chain config: %w", err)
	}

	db.SetTxContext(txHash, tx)
	snapshot := db.Snapshot()
	blockCtx := utils.PrepareBlockCtx(inputEnv, &hashError)
	evm := vm.NewEVM(*blockCtx, db, chainCfg, s.cfg.VmCfg)

	var msgResult messageResult
	var gasPool = new(core.GasPool)
	gasPool.AddGas(inputEnv.GetGasLimit())
	executionResult, err := core.ApplyMessage(evm, msg, gasPool)
	if err != nil {
		// if transaction fails, revert to the first snapshot.
		db.RevertToSnapshot(snapshot)
		finalError = errors.Join(fmt.Errorf("block: %v transaction: %v", block, tx), err)
	} else {
		// result should only be created if error was not returned
		msgResult = messageResult{
			failed:     executionResult.Failed(),
			returnData: executionResult.Return(),
			gasUsed:    executionResult.UsedGas,
			err:        executionResult.Err,
		}
	}

	// inform about failing transaction
	if msgResult.Failed() {
		s.log.Debugf("Block: %v\nTransaction %v\n Status: Failed; %v", block, tx, msgResult.GetError().Error())
	}

	// check whether getHash func produced an error
	if hashError != nil {
		finalError = errors.Join(finalError, hashError)
	}

	blockHash := common.HexToHash(fmt.Sprintf("0x%016d", block))
	// if no prior error, create result and pass it to the data.
	res = newTransactionResult(db.GetLogs(txHash, uint64(block), blockHash, inputEnv.GetTimestamp()), msg, msgResult, finalError, msg.From)
	return
}

// processPseudoTx processes pseudo transactions in Lachesis by applying the change in db state.
// The pseudo transactions includes Lachesis SFC, lachesis genesis and lachesis-opera transition.
func (s *TxProcessor) processPseudoTx(ws txcontext.WorldState, db state.VmStateDB) txcontext.Result {
	ws.ForEachAccount(func(addr common.Address, acc txcontext.Account) {
		db.SubBalance(addr, db.GetBalance(addr), tracing.BalanceChangeUnspecified)
		db.AddBalance(addr, acc.GetBalance(), tracing.BalanceChangeUnspecified)
		db.SetNonce(addr, acc.GetNonce(), tracing.NonceChangeUnspecified)
		db.SetCode(addr, acc.GetCode())
		acc.ForEachStorage(func(keyHash common.Hash, valueHash common.Hash) {
			db.SetState(addr, keyHash, valueHash)
		})
	})
	return newPseudoExecutionResult()
}

type toscaProcessor struct {
	processor tosca.Processor
	cfg       *utils.Config
	log       logger.Logger
}

func (t *toscaProcessor) processRegularTx(db state.VmStateDB, int, tx int, st txcontext.TxContext) (res transactionResult, finalError error) {
	// The main task of this function is to link the context provided through parameters
	// with the context required by a Tosca Processor implementation to execute a transaction.
	processor := t.processor

	blockEnvironment := st.GetBlockEnvironment()
	message := st.GetMessage()

	chainCfg, err := t.cfg.GetChainConfig(blockEnvironment.GetFork())
	if err != nil {
		return res, fmt.Errorf("cannot get chain config: %w", err)
	}

	block := blockEnvironment.GetNumber()
	baseFee := blockEnvironment.GetBaseFee()
	if message.GasPrice.Cmp(big.NewInt(0)) == 0 &&
		blockEnvironment.GetBaseFee() != nil &&
		blockEnvironment.GetBaseFee().Cmp(big.NewInt(0)) != 0 {
		// Base fee can not be lower than gas price
		baseFee = big.NewInt(0)
	}

	revision := tosca.R07_Istanbul
	if chainCfg.BerlinBlock != nil && block >= chainCfg.BerlinBlock.Uint64() {
		revision = tosca.R09_Berlin
	}
	if chainCfg.LondonBlock != nil && block >= chainCfg.LondonBlock.Uint64() {
		revision = tosca.R10_London
	}
	if chainCfg.MergeNetsplitBlock != nil && block >= chainCfg.MergeNetsplitBlock.Uint64() {
		revision = tosca.R11_Paris
	}
	if chainCfg.ShanghaiTime != nil && st.GetBlockEnvironment().GetTimestamp() >= *chainCfg.ShanghaiTime {
		revision = tosca.R12_Shanghai
	}
	if chainCfg.CancunTime != nil && st.GetBlockEnvironment().GetTimestamp() >= *chainCfg.CancunTime {
		revision = tosca.R13_Cancun
	}

	randao := tosca.Hash(bigToValue(blockEnvironment.GetDifficulty()))
	if revision >= tosca.R11_Paris {
		randao = tosca.Hash(*blockEnvironment.GetRandom())
	}

	blockParams := tosca.BlockParameters{
		BlockNumber: int64(block),
		Timestamp:   int64(blockEnvironment.GetTimestamp()),
		GasLimit:    tosca.Gas(blockEnvironment.GetGasLimit()),
		Coinbase:    tosca.Address(blockEnvironment.GetCoinbase()),
		ChainID:     tosca.Word(bigToValue(chainCfg.ChainID)),
		PrevRandao:  tosca.Hash(randao),
		BaseFee:     bigToValue(baseFee),
		BlobBaseFee: bigToValue(blockEnvironment.GetBlobBaseFee()),
		Revision:    revision,
	}

	transaction := messageToTransaction(message)

	context := &toscaTxContext{
		blockEnvironment: blockEnvironment,
		db:               db,
	}

	receipt, err := processor.Run(blockParams, transaction, context)
	if err != nil {
		return transactionResult{err: err}, err
	}

	log := []*types.Log{}
	for _, l := range receipt.Logs {
		topics := make([]common.Hash, len(l.Topics))
		for i, t := range l.Topics {
			topics[i] = common.Hash(t)
		}
		log = append(log, &types.Log{
			Address: common.Address(l.Address),
			Topics:  topics,
			Data:    l.Data,
		})
	}
	msg := st.GetMessage()

	if !receipt.Success {
		// The actual error is not relevant. Anything
		// that is not equal to nil will be considered
		// as a failed execution that got rolled back.
		err = fmt.Errorf("transaction failed")
	}

	result := &messageResult{
		gasUsed:    uint64(receipt.GasUsed),
		err:        err,
		returnData: receipt.Output,
		failed:     !receipt.Success,
	}

	return newTransactionResult(log, msg, result, finalError, msg.From), nil
}

func messageToTransaction(message *core.Message) tosca.Transaction {
	gasFeeCap := message.GasFeeCap
	gasTipCap := message.GasTipCap
	if message.GasPrice.Cmp(big.NewInt(0)) != 0 &&
		gasFeeCap.Cmp(big.NewInt(0)) == 0 &&
		gasTipCap.Cmp(big.NewInt(0)) == 0 {
		// Legacy transaction do not specify gas fee cap and gas tip cap but the gas price
		gasFeeCap = message.GasPrice
		gasTipCap = message.GasPrice
	}

	accessList := []tosca.AccessTuple{}
	for _, tuple := range message.AccessList {
		keys := make([]tosca.Key, len(tuple.StorageKeys))
		for i, key := range tuple.StorageKeys {
			keys[i] = tosca.Key(key)
		}
		accessList = append(accessList, tosca.AccessTuple{
			Address: tosca.Address(tuple.Address),
			Keys:    keys,
		})
	}

	var blobHashes []tosca.Hash
	if message.BlobHashes != nil {
		blobHashes = make([]tosca.Hash, len(message.BlobHashes))
		for i, hash := range message.BlobHashes {
			blobHashes[i] = tosca.Hash(hash)
		}
	}

	var recipient *tosca.Address
	if message.To != nil {
		recipient = (*tosca.Address)(message.To)
	}

	transaction := tosca.Transaction{
		Sender:        tosca.Address(message.From),
		Recipient:     recipient,
		Nonce:         message.Nonce,
		Input:         message.Data,
		Value:         bigToValue(message.Value),
		GasFeeCap:     bigToValue(gasFeeCap),
		GasTipCap:     bigToValue(gasTipCap),
		GasLimit:      tosca.Gas(message.GasLimit),
		BlobGasFeeCap: bigToValue(message.BlobGasFeeCap),
		BlobHashes:    blobHashes,
		AccessList:    accessList,
	}

	return transaction
}

// toscaTxContext is a bridge between Tosca's transaction context and the one provided by the executor.
type toscaTxContext struct {
	blockEnvironment txcontext.BlockEnvironment
	db               state.VmStateDB
}

func (a *toscaTxContext) CreateAccount(addr tosca.Address) {
	if !a.db.Exist(common.Address(addr)) {
		a.db.CreateAccount(common.Address(addr))
	}
	a.db.CreateContract(common.Address(addr))
}

func (a *toscaTxContext) HasEmptyStorage(addr tosca.Address) bool {
	rootHash := a.db.GetStorageRoot(common.Address(addr))
	return rootHash == common.Hash{} || rootHash == types.EmptyRootHash
}

func (a *toscaTxContext) AccountExists(addr tosca.Address) bool {
	return a.db.Exist(common.Address(addr))
}

func (a *toscaTxContext) GetBalance(addr tosca.Address) tosca.Value {
	return uint256ToValue(a.db.GetBalance(common.Address(addr)))
}

func (a *toscaTxContext) SetBalance(addr tosca.Address, balance tosca.Value) {
	want := balance.ToUint256()
	account := common.Address(addr)
	cur := a.db.GetBalance(account)
	if cur.Cmp(want) == 0 {
		return
	}
	if cur.Cmp(want) > 0 {
		diff := new(uint256.Int).Sub(cur, want)
		a.db.SubBalance(account, diff, 0 /*unknown tracing*/)
	} else {
		diff := new(uint256.Int).Sub(want, cur)
		a.db.AddBalance(account, diff, 0 /*unknown tracing*/)
	}
}

func (a *toscaTxContext) GetNonce(addr tosca.Address) uint64 {
	return a.db.GetNonce(common.Address(addr))
}

func (a *toscaTxContext) SetNonce(addr tosca.Address, nonce uint64) {
	a.db.SetNonce(common.Address(addr), nonce, tracing.NonceChangeUnspecified)
}

func (a *toscaTxContext) GetCodeSize(addr tosca.Address) int {
	return a.db.GetCodeSize(common.Address(addr))
}

func (a *toscaTxContext) GetCodeHash(addr tosca.Address) tosca.Hash {
	return tosca.Hash(a.db.GetCodeHash(common.Address(addr)))
}

func (a *toscaTxContext) GetCode(addr tosca.Address) tosca.Code {
	return a.db.GetCode(common.Address(addr))
}

func (a *toscaTxContext) SetCode(addr tosca.Address, code tosca.Code) {
	a.db.SetCode(common.Address(addr), code)
}

func (a *toscaTxContext) GetStorage(addr tosca.Address, key tosca.Key) tosca.Word {
	return tosca.Word(a.db.GetState(common.Address(addr), common.Hash(key)))
}

func (a *toscaTxContext) GetCommittedStorage(addr tosca.Address, key tosca.Key) tosca.Word {
	return tosca.Word(a.db.GetCommittedState(common.Address(addr), common.Hash(key)))
}

func (a *toscaTxContext) SetStorage(addr tosca.Address, key tosca.Key, value tosca.Word) tosca.StorageStatus {
	original := a.GetCommittedStorage(addr, key)
	current := a.GetStorage(addr, key)
	a.db.SetState(common.Address(addr), common.Hash(key), common.Hash(value))
	return tosca.GetStorageStatus(original, current, value)
}

func (a *toscaTxContext) GetTransientStorage(addr tosca.Address, key tosca.Key) tosca.Word {
	return tosca.Word(a.db.GetTransientState(common.Address(addr), common.Hash(key)))
}

func (a *toscaTxContext) SetTransientStorage(addr tosca.Address, key tosca.Key, value tosca.Word) {
	a.db.SetTransientState(common.Address(addr), common.Hash(key), common.Hash(value))
}

func (a *toscaTxContext) GetBlockHash(number int64) tosca.Hash {
	h, _ := a.blockEnvironment.GetBlockHash(uint64(number))
	return tosca.Hash(h)
}

func (a *toscaTxContext) EmitLog(log tosca.Log) {
	tpcs := make([]common.Hash, len(log.Topics))
	for i, t := range log.Topics {
		tpcs[i] = common.Hash(t)
	}

	a.db.AddLog(&types.Log{
		Address: common.Address(log.Address),
		Topics:  tpcs,
		Data:    log.Data,
	})
}

func (a *toscaTxContext) GetLogs() []tosca.Log {
	res := []tosca.Log{}
	for _, l := range a.db.GetLogs(common.Hash{}, 0, common.Hash{}, 0) {
		topics := make([]tosca.Hash, len(l.Topics))
		for i, t := range l.Topics {
			topics[i] = tosca.Hash(t)
		}
		res = append(res, tosca.Log{
			Address: tosca.Address(l.Address),
			Topics:  topics,
			Data:    l.Data,
		})
	}
	return res
}

func (a *toscaTxContext) SelfDestruct(addr tosca.Address, beneficiary tosca.Address) bool {
	selfdestructed := !a.db.HasSelfDestructed(common.Address(addr))

	if a.blockEnvironment.GetFork() == tosca.R13_Cancun.String() {
		a.db.SelfDestruct6780(common.Address(addr))
	} else {
		a.db.SelfDestruct(common.Address(addr))
	}
	return selfdestructed
}

func (a *toscaTxContext) AccessAccount(addr tosca.Address) tosca.AccessStatus {
	res := a.IsAddressInAccessList(addr)
	a.db.AddAddressToAccessList(common.Address(addr))
	if res {
		return tosca.WarmAccess
	}
	return tosca.ColdAccess
}

func (a *toscaTxContext) AccessStorage(addr tosca.Address, key tosca.Key) tosca.AccessStatus {
	_, res := a.IsSlotInAccessList(addr, key)
	a.db.AddSlotToAccessList(common.Address(addr), common.Hash(key))
	if res {
		return tosca.WarmAccess
	}
	return tosca.ColdAccess
}

func (a *toscaTxContext) HasSelfDestructed(addr tosca.Address) bool {
	return a.db.HasSelfDestructed(common.Address(addr))
}

func (a *toscaTxContext) CreateSnapshot() tosca.Snapshot {
	return tosca.Snapshot(a.db.Snapshot())
}

func (a *toscaTxContext) RestoreSnapshot(snapshot tosca.Snapshot) {
	a.db.RevertToSnapshot(int(snapshot))
}

func (a *toscaTxContext) IsAddressInAccessList(addr tosca.Address) bool {
	return a.db.AddressInAccessList(common.Address(addr))
}

func (a *toscaTxContext) IsSlotInAccessList(addr tosca.Address, key tosca.Key) (addressPresent, slotPresent bool) {
	return a.db.SlotInAccessList(common.Address(addr), common.Hash(key))
}

func bigToValue(value *big.Int) tosca.Value {
	if value == nil {
		return tosca.Value{}
	}
	var res tosca.Value
	value.FillBytes(res[:])
	return res
}

func uint256ToValue(value *uint256.Int) tosca.Value {
	return tosca.ValueFromUint256(value)
}
