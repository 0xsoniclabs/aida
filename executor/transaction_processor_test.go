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
	"github.com/stretchr/testify/require"
	"math/big"
	"strings"
	"testing"

	"github.com/0xsoniclabs/aida/ethtest"
	"github.com/0xsoniclabs/aida/state"
	"github.com/0xsoniclabs/aida/txcontext"
	substatecontext "github.com/0xsoniclabs/aida/txcontext/substate"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/substate/substate"
	"github.com/0xsoniclabs/tosca/go/tosca"
	"go.uber.org/mock/gomock"
)

// TestPrepareBlockCtx tests a creation of block context from substate environment.
func TestPrepareBlockCtx(t *testing.T) {
	gaslimit := uint64(10000000)
	blocknum := uint64(4600000)
	basefee := big.NewInt(12345)
	env := substatecontext.NewBlockEnvironment(&substate.Env{Difficulty: big.NewInt(1), GasLimit: gaslimit, Number: blocknum, Timestamp: 1675961395, BaseFee: basefee})

	var hashError error
	// BlockHashes are nil, expect an error
	blockCtx := utils.PrepareBlockCtx(env, &hashError)

	if blocknum != blockCtx.BlockNumber.Uint64() {
		t.Fatalf("Wrong block number")
	}
	if gaslimit != blockCtx.GasLimit {
		t.Fatalf("Wrong amount of gas limit")
	}
	if basefee.Cmp(blockCtx.BaseFee) != 0 {
		t.Fatalf("Wrong base fee")
	}
	if hashError != nil {
		t.Fatalf("Hash error; %v", hashError)
	}
}

func TestMakeTxProcessor_CanSelectBetweenProcessorImplementations(t *testing.T) {
	isAida := func(t *testing.T, p processor, name string) {
		_, ok := p.(*aidaProcessor)
		if !ok {
			t.Fatalf("Expected aidaProcessor from '%s', got %T", name, p)
		}
	}
	isTosca := func(t *testing.T, p processor, name string) {
		if _, ok := p.(*toscaProcessor); !ok {
			t.Fatalf("Expected toscaProcessor from '%s', got %T", name, p)
		}
	}

	tests := map[string]func(*testing.T, processor, string){
		"":         isAida,
		"opera":    isAida,
		"ethereum": isAida,
	}

	for name := range tosca.GetAllRegisteredProcessorFactories() {
		if _, present := tests[name]; !present {
			tests[name] = isTosca
		}
	}

	for name, check := range tests {
		t.Run(name, func(t *testing.T) {
			cfg := &utils.Config{
				ChainID: utils.MainnetChainID,
				EvmImpl: name,
				VmImpl:  "geth",
			}
			p, err := MakeTxProcessor(cfg)
			if err != nil {
				t.Fatalf("Failed to create tx processor; %v", err)
			}
			check(t, p.processor, name)
		})
	}

}

func TestEthTestProcessor_DoesNotExecuteTransactionWhenBlobGasCouldExceed(t *testing.T) {
	p, err := MakeEthTestProcessor(&utils.Config{})
	if err != nil {
		t.Fatalf("cannot make eth test processor: %v", err)
	}
	ctrl := gomock.NewController(t)
	// Process is returned early - nothing is expected
	stateDb := state.NewMockStateDB(ctrl)

	ctx := &Context{State: stateDb}
	err = p.Process(State[txcontext.TxContext]{Data: ethtest.CreateTransactionThatFailsBlobGasExceedCheck(t)}, ctx)
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}

	_, got := ctx.ExecutionResult.GetRawResult()
	want := "blob gas exceeds maximum"
	if !strings.EqualFold(got.Error(), want) {
		t.Errorf("unexpected error, got: %v, want: %v", got, want)
	}
}

func TestEthTestProcessor_DoesNotExecuteTransactionWithInvalidTxBytes(t *testing.T) {
	tests := []struct {
		name          string
		expectedError string
		data          txcontext.TxContext
	}{
		{
			name:          "fails_unmarshal",
			expectedError: "cannot unmarshal tx-bytes",
			data:          ethtest.CreateTransactionWithInvalidTxBytes(t),
		},
		{
			name:          "fails_validation",
			expectedError: "cannot validate sender",
			data:          ethtest.CreateTransactionThatFailsSenderValidation(t),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			p, err := MakeEthTestProcessor(&utils.Config{ChainID: utils.EthTestsChainID})
			if err != nil {
				t.Fatalf("cannot make eth test processor: %v", err)
			}
			ctrl := gomock.NewController(t)
			// Process is returned early - no calls are expected
			stateDb := state.NewMockStateDB(ctrl)

			ctx := &Context{State: stateDb}
			err = p.Process(State[txcontext.TxContext]{Data: test.data}, ctx)
			if err != nil {
				t.Fatalf("run failed: %v", err)
			}

			_, got := ctx.ExecutionResult.GetRawResult()
			if !strings.Contains(got.Error(), test.expectedError) {
				t.Errorf("unexpected error, got: %v, want: %v", got, test.expectedError)
			}
		})
	}
}

func TestMessageResult(t *testing.T) {
	e := errors.New("error")
	res := executionResult(messageResult{
		failed:     true,
		returnData: []byte{0x12},
		gasUsed:    10,
		err:        e,
	})

	require.True(t, res.Failed())
	require.Equal(t, res.Return(), []byte{0x12})
	require.Equal(t, res.GetGasUsed(), uint64(10))
	require.ErrorIs(t, res.GetError(), e)
}
