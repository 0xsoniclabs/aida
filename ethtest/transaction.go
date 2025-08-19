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

package ethtest

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/holiman/uint256"
)

// stTransaction indicates the executed transaction.
// Only one value for Data, GasLimit and Value is valid for each transaction.
// Any other value is shared among all transactions within one test file.
// Correct index is marked in stJSON.stPost.Index...
type stTransaction struct {
	GasPrice             *BigInt             `json:"gasPrice"`
	MaxFeePerGas         *BigInt             `json:"maxFeePerGas"`
	MaxPriorityFeePerGas *BigInt             `json:"maxPriorityFeePerGas"`
	Nonce                *BigInt             `json:"nonce"`
	To                   string              `json:"to"`
	Data                 []string            `json:"data"`
	AccessLists          []*types.AccessList `json:"accessLists,omitempty"`
	GasLimit             []*BigInt           `json:"gasLimit"`
	Value                []string            `json:"value"`
	PrivateKey           hexutil.Bytes       `json:"secretKey"`
	BlobGasFeeCap        *BigInt             `json:"maxFeePerBlobGas"`
	BlobHashes           []common.Hash       `json:"blobVersionedHashes"`
	Sender               *common.Address     `json:"sender"`
	AuthorizationList    []*stAuthorization  `json:"authorizationList,omitempty"`
}

type stAuthorization struct {
	ChainID *big.Int       `json:"chainId" gencodec:"required"`
	Address common.Address `json:"address" gencodec:"required"`
	Nonce   uint64         `json:"nonce" gencodec:"required"`
	V       uint8          `json:"v" gencodec:"required"`
	R       *big.Int       `json:"r" gencodec:"required"`
	S       *big.Int       `json:"s" gencodec:"required"`
}

// MarshalJSON marshals as JSON.
func (s *stAuthorization) MarshalJSON() ([]byte, error) {
	type stAuthorization struct {
		ChainID *math.HexOrDecimal256 `json:"chainId" gencodec:"required"`
		Address common.Address        `json:"address" gencodec:"required"`
		Nonce   math.HexOrDecimal64   `json:"nonce" gencodec:"required"`
		V       math.HexOrDecimal64   `json:"v" gencodec:"required"`
		R       *math.HexOrDecimal256 `json:"r" gencodec:"required"`
		S       *math.HexOrDecimal256 `json:"s" gencodec:"required"`
	}
	var enc stAuthorization
	enc.ChainID = (*math.HexOrDecimal256)(s.ChainID)
	enc.Address = s.Address
	enc.Nonce = math.HexOrDecimal64(s.Nonce)
	enc.V = math.HexOrDecimal64(s.V)
	enc.R = (*math.HexOrDecimal256)(s.R)
	enc.S = (*math.HexOrDecimal256)(s.S)
	return json.Marshal(&enc)
}

// UnmarshalJSON unmarshals from JSON.
func (s *stAuthorization) UnmarshalJSON(input []byte) error {
	type stAuthorization struct {
		ChainID *math.HexOrDecimal256 `json:"chainId" gencodec:"required"`
		Address *common.Address       `json:"address" gencodec:"required"`
		Nonce   *math.HexOrDecimal64  `json:"nonce" gencodec:"required"`
		V       *math.HexOrDecimal64  `json:"v" gencodec:"required"`
		R       *math.HexOrDecimal256 `json:"r" gencodec:"required"`
		S       *math.HexOrDecimal256 `json:"s" gencodec:"required"`
	}
	var dec stAuthorization
	if err := json.Unmarshal(input, &dec); err != nil {
		return err
	}
	if dec.ChainID == nil {
		return errors.New("missing required field 'chainId' for stAuthorization")
	}
	s.ChainID = (*big.Int)(dec.ChainID)
	if dec.Address == nil {
		return errors.New("missing required field 'address' for stAuthorization")
	}
	s.Address = *dec.Address
	if dec.Nonce == nil {
		return errors.New("missing required field 'nonce' for stAuthorization")
	}
	s.Nonce = uint64(*dec.Nonce)
	if dec.V == nil {
		return errors.New("missing required field 'v' for stAuthorization")
	}
	s.V = uint8(*dec.V)
	if dec.R == nil {
		return errors.New("missing required field 'r' for stAuthorization")
	}
	s.R = (*big.Int)(dec.R)
	if dec.S == nil {
		return errors.New("missing required field 's' for stAuthorization")
	}
	s.S = (*big.Int)(dec.S)
	return nil
}

func (tx *stTransaction) toMessage(ps stPost, baseFee *BigInt) (*core.Message, error) {
	var from common.Address
	// If 'sender' field is present, use that
	if tx.Sender != nil {
		from = *tx.Sender
	} else if len(tx.PrivateKey) > 0 {
		// Derive sender from private key if needed.
		key, err := crypto.ToECDSA(tx.PrivateKey)
		if err != nil {
			return nil, fmt.Errorf("invalid private key: %v", err)
		}
		from = crypto.PubkeyToAddress(key.PublicKey)
	}
	// Parse recipient if present.
	var to *common.Address
	if tx.To != "" {
		to = new(common.Address)
		if err := to.UnmarshalText([]byte(tx.To)); err != nil {
			return nil, fmt.Errorf("invalid to address: %v", err)
		}
	}

	// Get values specific to this post state.
	if ps.Indexes.Data > len(tx.Data) {
		return nil, fmt.Errorf("tx data index %d out of bounds", ps.Indexes.Data)
	}
	if ps.Indexes.Value > len(tx.Value) {
		return nil, fmt.Errorf("tx value index %d out of bounds", ps.Indexes.Value)
	}
	if ps.Indexes.Gas > len(tx.GasLimit) {
		return nil, fmt.Errorf("tx gas limit index %d out of bounds", ps.Indexes.Gas)
	}
	dataHex := tx.Data[ps.Indexes.Data]
	valueHex := tx.Value[ps.Indexes.Value]
	gasLimit := tx.GasLimit[ps.Indexes.Gas]

	value := new(big.Int)
	if valueHex != "0x" {
		v, ok := math.ParseBig256(valueHex)
		if !ok {
			return nil, fmt.Errorf("invalid tx value %q", valueHex)
		}
		value = v
	}
	data, err := hex.DecodeString(strings.TrimPrefix(dataHex, "0x"))
	if err != nil {
		return nil, fmt.Errorf("invalid tx data %q", dataHex)
	}
	var accessList types.AccessList
	if tx.AccessLists != nil && tx.AccessLists[ps.Indexes.Data] != nil {
		accessList = *tx.AccessLists[ps.Indexes.Data]
	}
	// If baseFee provided, set gasPrice to effectiveGasPrice.
	gasPrice := tx.GasPrice
	if baseFee != nil {
		if tx.MaxFeePerGas == nil {
			tx.MaxFeePerGas = gasPrice
		}
		if tx.MaxFeePerGas == nil {
			tx.MaxFeePerGas = new(BigInt)
		}
		if tx.MaxPriorityFeePerGas == nil {
			tx.MaxPriorityFeePerGas = tx.MaxFeePerGas
		}

		gasPrice = &BigInt{*bigMin(new(big.Int).Add(tx.MaxPriorityFeePerGas.Convert(), baseFee.Convert()),
			tx.MaxFeePerGas.Convert())}
	}
	if gasPrice == nil {
		return nil, fmt.Errorf("no gas price provided")
	}
	var authList []types.SetCodeAuthorization
	if tx.AuthorizationList != nil {
		authList = make([]types.SetCodeAuthorization, len(tx.AuthorizationList))
		for i, auth := range tx.AuthorizationList {
			authList[i] = types.SetCodeAuthorization{
				ChainID: *uint256.MustFromBig(auth.ChainID),
				Address: auth.Address,
				Nonce:   auth.Nonce,
				V:       auth.V,
				R:       *uint256.MustFromBig(auth.R),
				S:       *uint256.MustFromBig(auth.S),
			}
		}
	}

	msg := &core.Message{
		To:                    to,
		From:                  from,
		Nonce:                 tx.Nonce.Uint64(),
		Value:                 value,
		GasLimit:              gasLimit.Uint64(),
		GasPrice:              gasPrice.Convert(),
		GasFeeCap:             tx.MaxFeePerGas.Convert(),
		GasTipCap:             tx.MaxPriorityFeePerGas.Convert(),
		Data:                  data,
		AccessList:            accessList,
		BlobGasFeeCap:         tx.BlobGasFeeCap.Convert(),
		BlobHashes:            tx.BlobHashes,
		SetCodeAuthorizations: authList,
	}
	return msg, nil
}

func bigMin(a, b *big.Int) *big.Int {
	if a.Cmp(b) < 0 {
		return a
	}
	return b
}
