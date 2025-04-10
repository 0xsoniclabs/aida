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

package substate

import (
	"github.com/0xsoniclabs/aida/txcontext"
	"github.com/0xsoniclabs/substate/substate"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
)

func NewTxContext(data *substate.Substate) txcontext.TxContext {
	return &substateData{data}
}

type substateData struct {
	*substate.Substate
}

func (t *substateData) GetLogsHash() common.Hash {
	return common.Hash{}
}

func (t *substateData) GetStateHash() common.Hash {
	return common.Hash{}
}

func (t *substateData) GetInputState() txcontext.WorldState {
	return NewWorldState(t.InputSubstate)
}

func (t *substateData) GetOutputState() txcontext.WorldState {
	return NewWorldState(t.OutputSubstate)
}

func (t *substateData) GetBlockEnvironment() txcontext.BlockEnvironment {
	return NewBlockEnvironment(t.Env)
}

func (t *substateData) GetMessage() *core.Message {
	// todo remove iteration once fantom types are created
	var list types.AccessList
	for _, tuple := range t.Message.AccessList {
		var keys []common.Hash
		for _, key := range tuple.StorageKeys {
			keys = append(keys, common.Hash(key))
		}
		list = append(list, types.AccessTuple{Address: common.Address(tuple.Address), StorageKeys: keys})
	}

	var blobHashes []common.Hash
	for _, hash := range t.Message.BlobHashes {
		blobHashes = append(blobHashes, common.Hash(hash))
	}

	var authorizations []types.SetCodeAuthorization
	for _, a := range t.Message.SetCodeAuthorizations {
		authorizations = append(authorizations, types.SetCodeAuthorization{ChainID: a.ChainID, Address: common.Address(a.Address), Nonce: a.Nonce, V: a.V, R: a.R, S: a.S})
	}

	return &core.Message{
		To:                    (*common.Address)(t.Message.To),
		From:                  common.Address(t.Message.From),
		Nonce:                 t.Message.Nonce,
		Value:                 t.Message.Value,
		GasLimit:              t.Message.Gas,
		GasPrice:              t.Message.GasPrice,
		GasFeeCap:             t.Message.GasFeeCap,
		GasTipCap:             t.Message.GasTipCap,
		Data:                  t.Message.Data,
		AccessList:            list,
		BlobGasFeeCap:         t.Message.BlobGasFeeCap,
		BlobHashes:            blobHashes,
		SkipNonceChecks:       !t.Message.CheckNonce,
		SetCodeAuthorizations: authorizations,
	}
}

func (t *substateData) GetResult() txcontext.Result {
	return NewReceipt(t.Result)
}
