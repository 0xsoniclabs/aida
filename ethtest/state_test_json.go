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

package ethtest

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
)

// stJSON serves as a 'middleman' into which are data unmarshalled from geth test files.
type stJSON struct {
	path      string
	testLabel string
	Env       stBlockEnvironment  `json:"env"`
	Pre       types.GenesisAlloc  `json:"pre"`
	Tx        stTransaction       `json:"transaction"`
	Out       hexutil.Bytes       `json:"out"`
	Post      map[string][]stPost `json:"post"`
}

func (s *stJSON) setPath(path string) {
	s.path = path
}

func (s *stJSON) setTestLabel(testLabel string) {
	s.testLabel = testLabel
}

func (s *stJSON) CreateEnv(chainCfg *params.ChainConfig, fork string) *stBlockEnvironment {
	// Create copy as each tx needs its own env
	env := s.Env
	env.chainCfg = chainCfg
	env.fork = fork
	return &env
}

// stPost indicates data for each transaction.
type stPost struct {
	// RootHash holds expected state hash after a transaction is executed.
	RootHash common.Hash `json:"hash"`
	// LogsHash holds expected logs hash (Bloom) after a transaction is executed.
	LogsHash        common.Hash   `json:"logs"`
	TxBytes         hexutil.Bytes `json:"txbytes"`
	ExpectException string        `json:"expectException"`
	Indexes         Index         `json:"indexes"`
}

// Index indicates position of data, gas, value for executed transaction.
type Index struct {
	Data  int `json:"data"`
	Gas   int `json:"gas"`
	Value int `json:"value"`
}
