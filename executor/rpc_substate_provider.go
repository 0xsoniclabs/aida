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

//go:generate mockgen -source rpc_substate_provider.go -destination rpc_substate_provider_mocks.go -package executor

import (
	"encoding/hex"
	"fmt"
	"math"
	"math/big"
	"strconv"

	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/txcontext"
	substatecontext "github.com/0xsoniclabs/aida/txcontext/substate"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/substate/substate"
	"github.com/0xsoniclabs/substate/types"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/urfave/cli/v2"
)

// ----------------------------------------------------------------------------
//                              Implementation
// ----------------------------------------------------------------------------

// OpenRPCSubstateProvider opens a substate database as configured in the given parameters.
func OpenRPCSubstateProvider(cfg *utils.Config, ctxt *cli.Context) (Provider[txcontext.TxContext], error) {
	ipcPath := cfg.OperaDb + "/sonic.ipc"

	log := logger.NewLogger("info", "RPCSubstateProvider")
	client, err := utils.GetRpcOrIpcClient(ctxt.Context, cfg.ChainID, ipcPath, log)
	if err != nil {
		return nil, err
	}
	return &rpcSubstateProvider{
		client:              client,
		ctxt:                ctxt,
		numParallelDecoders: cfg.Workers,
	}, nil
}

// rpcSubstateProvider is an adapter of Aida's RPCRpcsubstateProvider interface defined above to the
// current substate implementation offered by github.com/0xsoniclabs/substate.
type rpcSubstateProvider struct {
	client              *rpc.Client
	ctxt                *cli.Context
	numParallelDecoders int
}

func (s rpcSubstateProvider) Run(from int, to int, consumer Consumer[txcontext.TxContext]) error {
	if to == -1 {
		return fmt.Errorf("substate recording doesn't support 'last' as block range boundary")
	}
	for blk := from; blk < to; blk++ {
		err := s.fetchBlockTxs(blk, consumer)
		if err != nil {
			return fmt.Errorf("failed to fetch block %d txs; %w", blk, err)
		}
	}
	return nil
}

func (s rpcSubstateProvider) Close() {
	s.client.Close()
}

func (s rpcSubstateProvider) fetchBlockTxs(blk int, consumer Consumer[txcontext.TxContext]) error {
	res, err := utils.RetrieveBlock(s.client, fmt.Sprintf("0x%x", blk), true)
	if err != nil {
		return fmt.Errorf("failed to retrieve block %d; %w", blk, err)
	}

	fmt.Printf("Block %d: %s\n", blk, res)
	//TODO store stateroot
	stateRoot := res["stateRoot"].(string)
	fmt.Printf("stateroot %s\n", stateRoot)

	txs := res["transactions"].([]interface{})
	for _, txI := range txs {
		tx := txI.(map[string]interface{})

		txHash := tx["hash"].(string)
		var receipt map[string]interface{}
		receipt, err = utils.RetrieveTxReceipt(s.client, txHash)

		var txIndex int64
		txIndex, err = strconv.ParseInt(tx["transactionIndex"].(string), 0, 32)

		var coinbase types.Address
		// TODO probably incorrect determine when "coinbase" and when "miner"
		coinbase = types.HexToAddress(res["miner"].(string))

		var difficulty *big.Int
		difficulty = new(big.Int)
		difficulty.SetString(res["difficulty"].(string)[2:], 16)

		var gasLimit uint64
		gasLimit, err = strconv.ParseUint(res["gasLimit"].(string), 0, 64)
		if err != nil {
			return fmt.Errorf("failed to parse block gas limit; %w", err)
		}

		var number = uint64(blk)
		var timestamp uint64
		timestamp, err = strconv.ParseUint(res["timestamp"].(string), 0, 64)
		if err != nil {
			return fmt.Errorf("failed to parse block timestamp; %w", err)
		}

		var baseFee *big.Int
		baseFee = new(big.Int)
		baseFee.SetString(res["baseFeePerGas"].(string)[2:], 16)

		var blobBaseFee *big.Int
		blobBaseFee = big.NewInt(1)

		var blockHashes map[uint64]types.Hash
		// TODO probably not needed

		var random *types.Hash
		// TODO probably not needed
		random = new(types.Hash)
		random.SetBytes([]byte(""))

		env := substate.NewEnv(coinbase, difficulty, gasLimit, number, timestamp, baseFee, blobBaseFee, blockHashes, random)

		var nonce uint64
		nonce, err = strconv.ParseUint(tx["nonce"].(string), 0, 64)
		if err != nil {
			return fmt.Errorf("failed to parse transaction nonce; %w", err)
		}

		var checkNonce bool
		// TODO probably not needed

		var gasPrice *big.Int
		gasPrice = new(big.Int)
		gasPrice.SetString(tx["gasPrice"].(string), 16)

		var gas uint64
		gas, err = strconv.ParseUint(tx["gas"].(string), 0, 64)
		if err != nil {
			return fmt.Errorf("failed to parse transaction gas; %w", err)
		}

		var from types.Address
		from = types.HexToAddress(tx["from"].(string))

		var to *types.Address
		if tx["to"] != nil {
			toA := types.HexToAddress(tx["to"].(string))
			to = &toA
		}

		var value *big.Int
		value = new(big.Int)
		value.SetString(tx["value"].(string), 16)

		var data []byte
		data, err = hex.DecodeString(tx["input"].(string)[2:])
		if err != nil {
			return fmt.Errorf("failed to decode input data; %w", err)
		}

		var dataHash *types.Hash
		// TODO hash data from above

		var ProtobufTxType *int32
		var typ uint64
		typ, err = strconv.ParseUint(tx["type"].(string), 0, 64)
		if err != nil {
			return fmt.Errorf("failed to parse block type; %w", err)
		}
		if typ > math.MaxInt32 {
			return fmt.Errorf("block type value out of range for int32: %d", typ)
		}
		ProtobufTxTypeI := int32(typ)
		ProtobufTxType = &ProtobufTxTypeI

		var accessList types.AccessList
		var gasFeeCap *big.Int
		gasFeeCap = big.NewInt(0)
		var gasTipCap *big.Int
		gasTipCap = big.NewInt(0)
		var blobGasFeeCap *big.Int
		var blobHashes []types.Hash
		msg := substate.NewMessage(nonce, checkNonce, gasPrice, gas, from, to, value, data, dataHash, ProtobufTxType, accessList, gasFeeCap, gasTipCap, blobGasFeeCap, blobHashes)

		var status uint64
		status, err = strconv.ParseUint(receipt["status"].(string), 0, 16)
		if err != nil {
			return fmt.Errorf("failed to parse transaction status; %w", err)
		}

		var bloom types.Bloom
		logsBloomStr := receipt["logsBloom"].(string)
		var bts []byte
		bts, err = hex.DecodeString(logsBloomStr[2:])
		bloom.SetBytes(bts)

		var logs []*types.Log
		//for _, logI := range receipt["logs"].([]interface{}) {
		//	log := logI.(map[string]interface{})
		//
		//	var address types.Address
		//	address.SetBytes([]byte(log["address"].(string)))
		//
		//	var topics []types.Hash
		//	for _, topicI := range log["topics"].([]interface{}) {
		//		topic := topicI.(string)
		//		var hash types.Hash
		//		hash.SetBytes([]byte(topic))
		//		topics = append(topics, hash)
		//	}
		//
		//	var data []byte
		//	data, err = hex.DecodeString(log["data"].(string)[2:])
		//	if err != nil {
		//		return fmt.Errorf("failed to decode log data; %w", err)
		//	}
		//
		//	logs = append(logs, types.Log{
		//}

		var contractAddress types.Address
		if receipt["contractAddress"] != nil {
			contractAddress.SetBytes([]byte(receipt["contractAddress"].(string)))
		}
		var gasUsed uint64
		gasUsed, err = strconv.ParseUint(receipt["gasUsed"].(string), 0, 64)
		if err != nil {
			return fmt.Errorf("failed to parse transaction gas used; %w", err)

		}

		result := substate.NewResult(status, bloom, logs, contractAddress, gasUsed)

		txSubstate := substate.NewSubstate(nil, nil, env, msg, result, uint64(blk), int(txIndex))
		err = consumer(TransactionInfo[txcontext.TxContext]{blk, int(txIndex), substatecontext.NewTxContext(txSubstate)})
		if err != nil {
			return err
		}
	}
	return nil
}
