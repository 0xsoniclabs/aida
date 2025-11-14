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

package delta

import (
	"context"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestNewStateTester_DefaultConfig(t *testing.T) {
	cfg := StateTesterConfig{
		DbImpl: "geth",
	}

	tester, err := NewStateTester(cfg)
	require.NoError(t, err)
	require.NotNil(t, tester)
}

func TestNewStateTester_CustomConfig(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := StateTesterConfig{
		DbImpl:       "geth",
		Variant:      "go-file",
		TmpDir:       tmpDir,
		CarmenSchema: 5,
		LogLevel:     "DEBUG",
		ChainID:      250,
	}

	tester, err := NewStateTester(cfg)
	require.NoError(t, err)
	require.NotNil(t, tester)
}

func TestStateTester_PassingTrace(t *testing.T) {
	cfg := StateTesterConfig{
		DbImpl:  "geth",
		TmpDir:  t.TempDir(),
		ChainID: 250,
	}

	tester, err := NewStateTester(cfg)
	require.NoError(t, err)

	addr := common.HexToAddress("0x1")
	ops := []TraceOp{
		{Kind: "BeginBlock", Args: []string{"1"}},
		{Kind: "BeginTransaction", Args: []string{"0"}},
		{Kind: "CreateAccount", Args: []string{addr.Hex()}},
		{Kind: "SetNonce", Args: []string{addr.Hex(), "1", "NonceChangeUnspecified"}},
		{Kind: "EndTransaction"},
		{Kind: "EndBlock"},
	}

	outcome, err := tester(context.Background(), ops)
	require.NoError(t, err)
	require.Equal(t, outcomePass, outcome)
}

func TestStateTester_FailingTrace(t *testing.T) {
	cfg := StateTesterConfig{
		DbImpl:  "geth",
		TmpDir:  t.TempDir(),
		ChainID: 250,
	}

	tester, err := NewStateTester(cfg)
	require.NoError(t, err)

	ops := []TraceOp{
		{Kind: "BeginBlock", Args: []string{"1"}},
		{Kind: "SetNonce", Args: []string{}},
	}

	outcome, err := tester(context.Background(), ops)
	require.NoError(t, err)
	require.Equal(t, outcomeFail, outcome)
}

func TestStateTester_UnsupportedOperation(t *testing.T) {
	cfg := StateTesterConfig{
		DbImpl:  "geth",
		TmpDir:  t.TempDir(),
		ChainID: 250,
	}

	tester, err := NewStateTester(cfg)
	require.NoError(t, err)

	ops := []TraceOp{
		{Kind: "BeginBlock", Args: []string{"1"}},
		{Kind: "GetCodeHashLc"},
	}

	outcome, err := tester(context.Background(), ops)
	require.NoError(t, err)
	require.Equal(t, outcomeFail, outcome)
}

func TestStateTester_ContextCancellation(t *testing.T) {
	cfg := StateTesterConfig{
		DbImpl:  "geth",
		TmpDir:  t.TempDir(),
		ChainID: 250,
	}

	tester, err := NewStateTester(cfg)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	ops := []TraceOp{
		{Kind: "BeginBlock", Args: []string{"1"}},
		{Kind: "EndBlock"},
	}

	outcome, err := tester(ctx, ops)
	require.Error(t, err)
	require.Equal(t, outcomeUnresolved, outcome)
}

func TestStateTester_EmptyDbImpl(t *testing.T) {
	cfg := StateTesterConfig{
		DbImpl: "",
		TmpDir: t.TempDir(),
	}

	tester, err := NewStateTester(cfg)
	require.NoError(t, err)
	require.NotNil(t, tester)
}

func TestStateTester_ComplexTrace(t *testing.T) {
	cfg := StateTesterConfig{
		DbImpl:  "geth",
		TmpDir:  t.TempDir(),
		ChainID: 250,
	}

	tester, err := NewStateTester(cfg)
	require.NoError(t, err)

	addr := common.HexToAddress("0x1234567890123456789012345678901234567890")
	key := common.HexToHash("0x0")
	value := common.HexToHash("0x42")

	ops := []TraceOp{
		{Kind: "BeginBlock", Args: []string{"100"}},
		{Kind: "BeginTransaction", Args: []string{"0"}},
		{Kind: "CreateAccount", Args: []string{addr.Hex()}},
		{Kind: "SetNonce", Args: []string{addr.Hex(), "1", "NonceChangeUnspecified"}},
		{Kind: "AddBalance", Args: []string{addr.Hex(), "1000000000000000000", "0", "BalanceChangeUnspecified", "1000000000000000000"}},
		{Kind: "SetState", Args: []string{addr.Hex(), key.Hex(), value.Hex()}},
		{Kind: "GetState", Args: []string{addr.Hex(), key.Hex()}},
		{Kind: "EndTransaction"},
		{Kind: "EndBlock"},
	}

	outcome, err := tester(context.Background(), ops)
	require.NoError(t, err)
	require.Equal(t, outcomePass, outcome)
}
