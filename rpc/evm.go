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

package rpc

import (
	"fmt"
	"math"
	"math/big"
	"strings"

	"github.com/0xsoniclabs/aida/state"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/sonic/ethapi"
	"github.com/0xsoniclabs/sonic/opera"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params"
	"github.com/status-im/keycard-go/hexutils"
)

// EvmExecutor represents requests executed over Ethereum Virtual Machine
type EvmExecutor struct {
	args      ethapi.TransactionArgs
	archive   state.NonCommittableStateDB
	timestamp uint64 // EVM requests require timestamp for correct execution
	chainCfg  *params.ChainConfig
	vmImpl    vm.InterpreterFactory
	blockId   *big.Int
	rules     opera.EconomyRules
}

const maxGasLimit = 9995800     // used when request does not specify gas
const globalGasCap = 50_000_000 // highest gas allowance used for estimateGas

// newEvmExecutor creates EvmExecutor for executing requests into StateDB that demand usage of EVM
func newEvmExecutor(blockID uint64, archive state.NonCommittableStateDB, cfg *utils.Config, params map[string]interface{}, timestamp uint64) (*EvmExecutor, error) {
	factory, err := cfg.GetInterpreterFactory()
	if err != nil {
		return nil, fmt.Errorf("cannot get interpreter factory: %w", err)
	}
	chainCfg, err := cfg.GetChainConfig("")
	if err != nil {
		return nil, fmt.Errorf("cannot get chain config: %w", err)
	}

	return &EvmExecutor{
		args:      newTxArgs(params),
		archive:   archive,
		timestamp: timestamp,
		chainCfg:  chainCfg,
		vmImpl:    factory,
		blockId:   new(big.Int).SetUint64(blockID),
		rules:     opera.DefaultEconomyRules(),
	}, nil
}

// newTxArgs decodes recorded params into ethapi.TransactionArgs
func newTxArgs(params map[string]interface{}) ethapi.TransactionArgs {
	var args ethapi.TransactionArgs

	if v, ok := params["from"]; ok && v != nil {
		args.From = new(common.Address)
		*args.From = common.HexToAddress(v.(string))
	}

	if v, ok := params["to"]; ok && v != nil {
		args.To = new(common.Address)
		*args.To = common.HexToAddress(v.(string))
	}

	if v, ok := params["value"]; ok && v != nil {
		value := new(big.Int)
		value.SetString(strings.TrimPrefix(v.(string), "0x"), 16)
		args.Value = (*hexutil.Big)(value)
	}

	args.Gas = new(hexutil.Uint64)
	if v, ok := params["gas"]; ok && v != nil {
		gas := new(big.Int)
		gas.SetString(strings.TrimPrefix(v.(string), "0x"), 16)
		*args.Gas = hexutil.Uint64(gas.Uint64())
	} else {
		// if gas cap is not specified, we use maxGasLimit
		*args.Gas = hexutil.Uint64(maxGasLimit)
	}

	if v, ok := params["gasPrice"]; ok && v != nil {
		gasPrice := new(big.Int)
		gasPrice.SetString(strings.TrimPrefix(v.(string), "0x"), 16)
		args.GasPrice = new(hexutil.Big)
		args.GasPrice = (*hexutil.Big)(gasPrice)
	}

	if v, ok := params["data"]; ok && v != nil {
		s := strings.TrimPrefix(v.(string), "0x")
		data := hexutils.HexToBytes(s)
		args.Data = new(hexutil.Bytes)
		args.Data = (*hexutil.Bytes)(&data)
	}

	return args
}

// newEVM creates new instance of EVM with given parameters
func (e *EvmExecutor) newEVM(msg *core.Message, hashErr *error) *vm.EVM {
	var (
		getHash  func(uint64) common.Hash
		blockCtx vm.BlockContext
		vmConfig vm.Config
	)

	getHash = func(_ uint64) common.Hash {
		h, err := e.archive.GetHash()
		*hashErr = err
		return h
	}

	blockCtx = vm.BlockContext{
		CanTransfer: core.CanTransfer,
		Transfer:    core.Transfer,
		Coinbase:    common.Address{}, // opera based value
		BlockNumber: e.blockId,
		Difficulty:  big.NewInt(1),  // evmcore/evm.go
		GasLimit:    math.MaxUint64, // evmcore/dummy_block.go
		GetHash:     getHash,
		BaseFee:     e.rules.MinGasPrice, // big.NewInt(1e9)
		Time:        e.timestamp,
	}

	// The default rules only work until there are blocks that have been created
	// using the single-proposer mode. The crucial difference in the VM setup is
	// that in the single-proposer mode the charging of excess gas is disabled,
	// while in the distributed-proposer mode (the default mode), it is enabled.
	defaultVmConfig := opera.GetVmConfig(opera.Rules{})
	vmConfig = defaultVmConfig
	vmConfig.NoBaseFee = true
	vmConfig.Interpreter = e.vmImpl

	return vm.NewEVM(blockCtx, e.archive, e.chainCfg, vmConfig)
}

// sendCall executes the call method in the EvmExecutor with given archive
func (e *EvmExecutor) sendCall() (*core.ExecutionResult, error) {
	var (
		gp              *core.GasPool
		executionResult *core.ExecutionResult
		err             error
		msg             *core.Message
		evm             *vm.EVM
	)

	gp = new(core.GasPool).AddGas(math.MaxUint64) // based in opera
	msg, err = e.args.ToMessage(globalGasCap, e.rules.MinGasPrice)
	if err != nil {
		return nil, err
	}

	var hashErr *error
	evm = e.newEVM(msg, hashErr)

	executionResult, err = core.ApplyMessage(evm, msg, gp)
	if err != nil {
		return executionResult, fmt.Errorf("err: %v (supplied gas %v)", err, e.args.Gas)
	}
	if executionResult.Err != nil {
		return nil, fmt.Errorf("execution returned err; %w", executionResult.Err)
	}

	if hashErr != nil {
		return nil, fmt.Errorf("cannot get state hash; %w", *hashErr)
	}

	// If the timer caused an abort, return an appropriate error message
	if evm.Cancelled() {
		return nil, fmt.Errorf("execution aborted: timeout")
	}

	return executionResult, nil

}

// sendEstimateGas executes estimateGas method in the EvmExecutor
// It calculates how much gas would transaction need if it was executed
func (e *EvmExecutor) sendEstimateGas() (hexutil.Uint64, error) {
	panic("not implemented")
}

// executable tries to execute call with given gas into EVM. This func is used for estimateGas calculation
func (e *EvmExecutor) executable(gas uint64) (bool, *core.ExecutionResult, error) {
	e.args.Gas = (*hexutil.Uint64)(&gas)

	result, err := e.sendCall()

	if err != nil {
		if strings.Contains(err.Error(), "intrinsic gas too low") {
			return true, nil, nil // Special case, raise gas limit
		}
		return true, nil, err // Bail out
	}
	return result.Failed(), result, nil
}
