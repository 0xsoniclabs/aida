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
	substatetypes "github.com/0xsoniclabs/substate/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/uint256"
)

func NewAccount(acc *substate.Account) txcontext.Account {
	return &account{acc}
}

type account struct {
	*substate.Account
}

func (a *account) GetNonce() uint64 {
	return a.Nonce
}

func (a *account) GetBalance() *uint256.Int {
	return a.Balance
}

func (a *account) HasStorageAt(key common.Hash) bool {
	_, ok := a.Storage[substatetypes.Hash(key)]
	return ok
}

func (a *account) GetStorageAt(hash common.Hash) common.Hash {
	return common.Hash(a.Storage[substatetypes.Hash(hash)])
}

func (a *account) GetCode() []byte {
	return a.Code
}

func (a *account) GetStorageSize() int {
	return len(a.Storage)
}

func (a *account) ForEachStorage(h txcontext.StorageHandler) {
	for keyHash, valueHash := range a.Storage {
		h(common.Hash(keyHash), common.Hash(valueHash))
	}
}

func (a *account) String() string {
	return txcontext.AccountString(a)
}
