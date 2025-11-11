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
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestCollectMetadata_AssignsContracts(t *testing.T) {
	addr := common.HexToAddress("0x1")
	ops := []TraceOp{
		{Raw: "BeginBlock, 1", Kind: "BeginBlock", HasBlock: true, Block: 1},
		{Raw: "BeginTransaction, 0", Kind: "BeginTransaction", HasBlock: true, Block: 1},
		{
			Raw:         fmt.Sprintf("SetState, %s, 0x0, 0x0, 0x0", addr.Hex()),
			Kind:        "SetState",
			HasContract: true,
			Contract:    addr,
			HasBlock:    true,
			Block:       1,
		},
		{Raw: "GetStateLcls", Kind: "GetStateLcls", HasBlock: true, Block: 1},
		{Raw: "EndTransaction", Kind: "EndTransaction", HasBlock: true, Block: 1},
		{Raw: "EndBlock", Kind: "EndBlock", HasBlock: true, Block: 1},
	}

	meta := collectMetadata(ops, defaultMandatoryKinds())

	require.Len(t, meta, len(ops), "metadata should match operations count")
	require.True(t, meta[2].HasContract, "SetState should have contract metadata")
	require.Equal(t, addr, meta[2].Contract)
	require.True(t, meta[3].HasContract, "GetStateLcls should inherit contract")
	require.Equal(t, addr, meta[3].Contract)
}

func TestMinimize_PrefixReduction(t *testing.T) {
	addr := common.HexToAddress("0x2")
	ops := []TraceOp{
		{Raw: "BeginBlock, 1", Kind: "BeginBlock", HasBlock: true, Block: 1},
		{Raw: "BeginTransaction, 0", Kind: "BeginTransaction", HasBlock: true, Block: 1},
		{
			Raw:         fmt.Sprintf("GetBalance, %s, 0", addr.Hex()),
			Kind:        "GetBalance",
			HasContract: true,
			Contract:    addr,
			HasBlock:    true,
			Block:       1,
		},
		{
			Raw:         fmt.Sprintf("SetState, %s, 0x0, 0x0, 0x0", addr.Hex()),
			Kind:        "SetState",
			HasContract: true,
			Contract:    addr,
			HasBlock:    true,
			Block:       1,
		},
		{Raw: "EndTransaction", Kind: "EndTransaction", HasBlock: true, Block: 1},
		{Raw: "EndBlock", Kind: "EndBlock", HasBlock: true, Block: 1},
	}

	test := func(_ context.Context, candidate []TraceOp) (Outcome, error) {
		for _, op := range candidate {
			if op.Kind == "SetState" {
				return OutcomeFail, nil
			}
		}
		return OutcomePass, nil
	}

	m := NewMinimizer(MinimizerConfig{
		AddressSampleRuns: 1,
		RandSeed:          1,
	})

	result, err := m.Minimize(context.Background(), ops, test)
	require.NoError(t, err)

	hasSetState := false
	for _, op := range result {
		require.NotEqual(t, "GetBalance", op.Kind, "GetBalance should be pruned")
		if op.Kind == "SetState" {
			hasSetState = true
		}
	}
	require.True(t, hasSetState, "SetState must remain in minimized trace")
}

func TestMinimize_AddressReduction(t *testing.T) {
	addrA := common.HexToAddress("0x3")
	addrB := common.HexToAddress("0x4")

	ops := []TraceOp{
		{Raw: "BeginBlock, 1", Kind: "BeginBlock", HasBlock: true, Block: 1},
		{Raw: "BeginTransaction, 0", Kind: "BeginTransaction", HasBlock: true, Block: 1},
		{
			Raw:         fmt.Sprintf("SetState, %s, 0x0, 0x0, 0x0", addrA.Hex()),
			Kind:        "SetState",
			HasContract: true,
			Contract:    addrA,
			HasBlock:    true,
			Block:       1,
		},
		{Raw: "GetStateLcls", Kind: "GetStateLcls", HasBlock: true, Block: 1},
		{
			Raw:         fmt.Sprintf("SetState, %s, 0x0, 0x0, 0x0", addrB.Hex()),
			Kind:        "SetState",
			HasContract: true,
			Contract:    addrB,
			HasBlock:    true,
			Block:       1,
		},
		{Raw: "EndTransaction", Kind: "EndTransaction", HasBlock: true, Block: 1},
		{Raw: "EndBlock", Kind: "EndBlock", HasBlock: true, Block: 1},
	}

	test := func(_ context.Context, candidate []TraceOp) (Outcome, error) {
		for _, op := range candidate {
			if op.Kind == "SetState" && op.Contract == addrA {
				return OutcomeFail, nil
			}
		}
		return OutcomePass, nil
	}

	m := NewMinimizer(MinimizerConfig{
		AddressSampleRuns: 5,
		RandSeed:          2,
	})

	result, err := m.Minimize(context.Background(), ops, test)
	require.NoError(t, err)

	for _, op := range result {
		require.NotEqual(t, addrB, op.Contract, "contract B should be removed")
	}
}

func TestOutcome_String(t *testing.T) {
	require.Equal(t, "pass", OutcomePass.String())
	require.Equal(t, "fail", OutcomeFail.String())
	require.Equal(t, "unresolved", OutcomeUnresolved.String())
	require.Equal(t, "unknown(99)", Outcome(99).String())
}

func TestMinimize_NilTestFunc(t *testing.T) {
	ops := []TraceOp{
		{Kind: "BeginBlock"},
	}
	m := NewMinimizer(MinimizerConfig{})
	_, err := m.Minimize(context.Background(), ops, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "test function must be provided")
}

func TestMinimize_EmptyTrace(t *testing.T) {
	m := NewMinimizer(MinimizerConfig{})
	test := func(_ context.Context, ops []TraceOp) (Outcome, error) {
		return OutcomeFail, nil
	}
	_, err := m.Minimize(context.Background(), []TraceOp{}, test)
	require.Error(t, err)
	require.Contains(t, err.Error(), "trace is empty")
}

func TestMinimize_ReducePrefixError(t *testing.T) {
	ops := []TraceOp{
		{Kind: "BeginBlock"},
	}
	m := NewMinimizer(MinimizerConfig{})
	test := func(_ context.Context, ops []TraceOp) (Outcome, error) {
		return OutcomeFail, fmt.Errorf("test error")
	}
	_, err := m.Minimize(context.Background(), ops, test)
	require.Error(t, err)
	require.Contains(t, err.Error(), "test error")
}

func TestMinimize_ReduceAddressesError(t *testing.T) {
	addr := common.HexToAddress("0x1")
	ops := []TraceOp{
		{Kind: "BeginBlock"},
		{Kind: "SetState", HasContract: true, Contract: addr},
	}
	m := NewMinimizer(MinimizerConfig{})
	callCount := 0
	test := func(_ context.Context, ops []TraceOp) (Outcome, error) {
		callCount++
		if callCount == 1 {
			return OutcomeFail, nil
		}
		return OutcomeFail, fmt.Errorf("address reduction error")
	}
	_, err := m.Minimize(context.Background(), ops, test)
	require.Error(t, err)
	require.Contains(t, err.Error(), "address reduction error")
}

func TestReducePrefix_EmptyMetadata(t *testing.T) {
	m := NewMinimizer(MinimizerConfig{})
	test := func(_ context.Context, ops []TraceOp) (Outcome, error) {
		return OutcomeFail, nil
	}
	_, _, err := m.reducePrefix(context.Background(), []OperationMeta{}, test)
	require.Error(t, err)
	require.Contains(t, err.Error(), "empty metadata")
}

func TestReducePrefix_TestFunctionError(t *testing.T) {
	meta := []OperationMeta{
		{Op: TraceOp{Kind: "BeginBlock"}, Kind: "BeginBlock"},
	}
	m := NewMinimizer(MinimizerConfig{})
	test := func(_ context.Context, ops []TraceOp) (Outcome, error) {
		return OutcomeFail, fmt.Errorf("test function error")
	}
	_, _, err := m.reducePrefix(context.Background(), meta, test)
	require.Error(t, err)
	require.Contains(t, err.Error(), "test function error")
}

func TestReducePrefix_InputDoesNotFail(t *testing.T) {
	meta := []OperationMeta{
		{Op: TraceOp{Kind: "BeginBlock"}, Kind: "BeginBlock"},
	}
	m := NewMinimizer(MinimizerConfig{})
	test := func(_ context.Context, ops []TraceOp) (Outcome, error) {
		return OutcomePass, nil
	}
	_, _, err := m.reducePrefix(context.Background(), meta, test)
	require.Error(t, err)
	require.Equal(t, ErrInputDoesNotFail, err)
}

func TestReducePrefix_ContextCancellation(t *testing.T) {
	meta := make([]OperationMeta, 10)
	for i := range meta {
		meta[i] = OperationMeta{
			Op:    TraceOp{Kind: "GetBalance"},
			Kind:  "GetBalance",
			Index: i,
		}
	}

	m := NewMinimizer(MinimizerConfig{})
	test := func(_ context.Context, ops []TraceOp) (Outcome, error) {
		return OutcomeFail, nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, _, err := m.reducePrefix(ctx, meta, test)
	require.Error(t, err)
	require.Equal(t, context.Canceled, err)
}

func TestReducePrefix_BinarySearchTestError(t *testing.T) {
	meta := make([]OperationMeta, 10)
	for i := range meta {
		meta[i] = OperationMeta{
			Op:    TraceOp{Kind: "GetBalance"},
			Kind:  "GetBalance",
			Index: i,
		}
	}

	m := NewMinimizer(MinimizerConfig{})
	callCount := 0
	test := func(_ context.Context, ops []TraceOp) (Outcome, error) {
		callCount++
		if callCount == 1 {
			return OutcomeFail, nil
		}
		return OutcomeFail, fmt.Errorf("binary search error")
	}

	_, _, err := m.reducePrefix(context.Background(), meta, test)
	require.Error(t, err)
	require.Contains(t, err.Error(), "binary search error")
}

func TestReduceAddresses_SingleAddress(t *testing.T) {
	addr := common.HexToAddress("0x1")
	ops := []TraceOp{
		{Kind: "SetState", HasContract: true, Contract: addr},
	}
	meta := collectMetadata(ops, defaultMandatoryKinds())

	m := NewMinimizer(MinimizerConfig{})
	test := func(_ context.Context, ops []TraceOp) (Outcome, error) {
		return OutcomeFail, nil
	}

	result, _, err := m.reduceAddresses(context.Background(), ops, meta, test)
	require.NoError(t, err)
	require.Equal(t, ops, result)
}

func TestReduceAddresses_ContextCancellation(t *testing.T) {
	addr1 := common.HexToAddress("0x1")
	addr2 := common.HexToAddress("0x2")
	ops := []TraceOp{
		{Kind: "SetState", HasContract: true, Contract: addr1},
		{Kind: "SetState", HasContract: true, Contract: addr2},
	}
	meta := collectMetadata(ops, defaultMandatoryKinds())

	m := NewMinimizer(MinimizerConfig{})
	test := func(_ context.Context, ops []TraceOp) (Outcome, error) {
		return OutcomeFail, nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, _, err := m.reduceAddresses(ctx, ops, meta, test)
	require.Error(t, err)
	require.Equal(t, context.Canceled, err)
}

func TestReduceAddresses_SampleSizeZero(t *testing.T) {
	addr1 := common.HexToAddress("0x1")
	ops := []TraceOp{
		{Kind: "SetState", HasContract: true, Contract: addr1},
	}
	meta := collectMetadata(ops, defaultMandatoryKinds())

	m := NewMinimizer(MinimizerConfig{
		MaxFactor:         10,
		AddressSampleRuns: 1,
	})
	testCalled := false
	test := func(_ context.Context, ops []TraceOp) (Outcome, error) {
		testCalled = true
		return OutcomeFail, nil
	}

	result, _, err := m.reduceAddresses(context.Background(), ops, meta, test)
	require.NoError(t, err)
	require.Equal(t, ops, result)
	require.False(t, testCalled)
}

func TestReduceAddresses_SampleSizeEqualsAddressCount(t *testing.T) {
	addr1 := common.HexToAddress("0x1")
	addr2 := common.HexToAddress("0x2")
	ops := []TraceOp{
		{Kind: "SetState", HasContract: true, Contract: addr1},
		{Kind: "SetState", HasContract: true, Contract: addr2},
	}
	meta := collectMetadata(ops, defaultMandatoryKinds())

	m := NewMinimizer(MinimizerConfig{
		MaxFactor:         1,
		AddressSampleRuns: 1,
	})
	test := func(_ context.Context, ops []TraceOp) (Outcome, error) {
		return OutcomeFail, nil
	}

	result, _, err := m.reduceAddresses(context.Background(), ops, meta, test)
	require.NoError(t, err)
	require.Equal(t, ops, result)
}

func TestReduceAddresses_TestErrorInLoop(t *testing.T) {
	addr1 := common.HexToAddress("0x1")
	addr2 := common.HexToAddress("0x2")
	ops := []TraceOp{
		{Kind: "SetState", HasContract: true, Contract: addr1},
		{Kind: "SetState", HasContract: true, Contract: addr2},
	}
	meta := collectMetadata(ops, defaultMandatoryKinds())

	m := NewMinimizer(MinimizerConfig{
		AddressSampleRuns: 5,
	})
	test := func(_ context.Context, ops []TraceOp) (Outcome, error) {
		return OutcomeFail, fmt.Errorf("test error in loop")
	}

	_, _, err := m.reduceAddresses(context.Background(), ops, meta, test)
	require.Error(t, err)
	require.Contains(t, err.Error(), "test error in loop")
}

func TestReduceAddresses_ContextCancelInLoop(t *testing.T) {
	addr1 := common.HexToAddress("0x1")
	addr2 := common.HexToAddress("0x2")
	ops := []TraceOp{
		{Kind: "SetState", HasContract: true, Contract: addr1},
		{Kind: "SetState", HasContract: true, Contract: addr2},
	}
	meta := collectMetadata(ops, defaultMandatoryKinds())

	m := NewMinimizer(MinimizerConfig{
		AddressSampleRuns: 5,
		RandSeed:          1,
	})

	callCount := 0
	test := func(ctx context.Context, ops []TraceOp) (Outcome, error) {
		callCount++
		if callCount > 1 {
			return OutcomeFail, context.Canceled
		}
		return OutcomePass, nil
	}

	ctx := context.Background()
	_, _, err := m.reduceAddresses(ctx, ops, meta, test)
	require.Error(t, err)
}

func TestReduceAddresses_NoReduction(t *testing.T) {
	addr1 := common.HexToAddress("0x1")
	addr2 := common.HexToAddress("0x2")
	ops := []TraceOp{
		{Kind: "SetState", HasContract: true, Contract: addr1},
		{Kind: "SetState", HasContract: true, Contract: addr2},
	}
	meta := collectMetadata(ops, defaultMandatoryKinds())

	m := NewMinimizer(MinimizerConfig{
		AddressSampleRuns: 2,
		RandSeed:          1,
	})
	test := func(_ context.Context, ops []TraceOp) (Outcome, error) {
		return OutcomePass, nil
	}

	result, _, err := m.reduceAddresses(context.Background(), ops, meta, test)
	require.NoError(t, err)
	require.Equal(t, ops, result)
}

func TestSampleAddresses_ZeroSize(t *testing.T) {
	m := NewMinimizer(MinimizerConfig{})
	addresses := []common.Address{
		common.HexToAddress("0x1"),
		common.HexToAddress("0x2"),
	}
	result := m.sampleAddresses(addresses, 0)
	require.Nil(t, result)
}

func TestSampleAddresses_EmptyAddresses(t *testing.T) {
	m := NewMinimizer(MinimizerConfig{})
	result := m.sampleAddresses([]common.Address{}, 5)
	require.Nil(t, result)
}

func TestSampleAddresses_SizeExceedsLength(t *testing.T) {
	m := NewMinimizer(MinimizerConfig{RandSeed: 1})
	addresses := []common.Address{
		common.HexToAddress("0x1"),
		common.HexToAddress("0x2"),
	}
	result := m.sampleAddresses(addresses, 10)
	require.Len(t, result, 2)
}

func TestLog_WithLogger(t *testing.T) {
	var logged string
	m := NewMinimizer(MinimizerConfig{
		Logger: func(format string, args ...any) {
			logged = fmt.Sprintf(format, args...)
		},
	})
	m.log("test %s %d", "message", 42)
	require.Equal(t, "test message 42", logged)
}

func TestLog_NoLogger(t *testing.T) {
	m := NewMinimizer(MinimizerConfig{})
	m.log("test message")
}

func TestNewMinimizer_DefaultAddressSampleRuns(t *testing.T) {
	m := NewMinimizer(MinimizerConfig{})
	require.Equal(t, 5, m.cfg.AddressSampleRuns)
}

func TestNewMinimizer_DefaultMandatoryKinds(t *testing.T) {
	m := NewMinimizer(MinimizerConfig{})
	require.NotNil(t, m.cfg.MandatoryKinds)
	_, ok := m.cfg.MandatoryKinds["BeginBlock"]
	require.True(t, ok)
}

func TestNewMinimizer_DefaultRandSeed(t *testing.T) {
	m := NewMinimizer(MinimizerConfig{RandSeed: 0})
	require.NotNil(t, m.rand)
}

func TestUniqueContracts(t *testing.T) {
	addr1 := common.HexToAddress("0x1")
	addr2 := common.HexToAddress("0x2")
	ops := []TraceOp{
		{Kind: "SetState", HasContract: true, Contract: addr1},
		{Kind: "SetState", HasContract: true, Contract: addr2},
		{Kind: "GetBalance", HasContract: true, Contract: addr1},
	}

	result := UniqueContracts(ops)
	require.Len(t, result, 2)
}

func TestReduceAddresses_SampleSizeZeroWithMultipleAddresses(t *testing.T) {
	addresses := make([]common.Address, 2)
	for i := range addresses {
		addresses[i] = common.HexToAddress(fmt.Sprintf("0x%d", i+1))
	}

	ops := make([]TraceOp, len(addresses))
	for i, addr := range addresses {
		ops[i] = TraceOp{Kind: "SetState", HasContract: true, Contract: addr}
	}
	meta := collectMetadata(ops, defaultMandatoryKinds())

	m := NewMinimizer(MinimizerConfig{
		MaxFactor:         3,
		AddressSampleRuns: 1,
	})
	test := func(_ context.Context, ops []TraceOp) (Outcome, error) {
		return OutcomePass, nil
	}

	result, _, err := m.reduceAddresses(context.Background(), ops, meta, test)
	require.NoError(t, err)
	require.Equal(t, ops, result)
}

func TestReduceAddresses_SampleSizeEqualsAddresses(t *testing.T) {
	addresses := make([]common.Address, 3)
	for i := range addresses {
		addresses[i] = common.HexToAddress(fmt.Sprintf("0x%d", i+1))
	}

	ops := make([]TraceOp, len(addresses))
	for i, addr := range addresses {
		ops[i] = TraceOp{Kind: "SetState", HasContract: true, Contract: addr}
	}
	meta := collectMetadata(ops, defaultMandatoryKinds())

	m := NewMinimizer(MinimizerConfig{
		MaxFactor:         1,
		AddressSampleRuns: 1,
	})
	test := func(_ context.Context, ops []TraceOp) (Outcome, error) {
		return OutcomePass, nil
	}

	result, _, err := m.reduceAddresses(context.Background(), ops, meta, test)
	require.NoError(t, err)
	require.Equal(t, ops, result)
}

func TestReduceAddresses_ContextCancelInInnerLoop(t *testing.T) {
	addr1 := common.HexToAddress("0x1")
	addr2 := common.HexToAddress("0x2")
	addr3 := common.HexToAddress("0x3")
	ops := []TraceOp{
		{Kind: "SetState", HasContract: true, Contract: addr1},
		{Kind: "SetState", HasContract: true, Contract: addr2},
		{Kind: "SetState", HasContract: true, Contract: addr3},
	}
	meta := collectMetadata(ops, defaultMandatoryKinds())

	m := NewMinimizer(MinimizerConfig{
		AddressSampleRuns: 10,
		RandSeed:          1,
	})

	ctx, cancel := context.WithCancel(context.Background())
	callCount := 0
	test := func(ctx context.Context, ops []TraceOp) (Outcome, error) {
		callCount++
		if callCount > 1 {
			cancel()
		}
		return OutcomePass, nil
	}

	_, _, err := m.reduceAddresses(ctx, ops, meta, test)
	require.Error(t, err)
	require.Equal(t, context.Canceled, err)
}

func TestReduceAddresses_SampleSizeZeroDuringIteration(t *testing.T) {
	addr1 := common.HexToAddress("0x1")
	addr2 := common.HexToAddress("0x2")
	ops := []TraceOp{
		{Kind: "SetState", HasContract: true, Contract: addr1},
		{Kind: "SetState", HasContract: true, Contract: addr2},
	}
	meta := collectMetadata(ops, defaultMandatoryKinds())

	logged := []string{}
	m := NewMinimizer(MinimizerConfig{
		MaxFactor:         10,
		AddressSampleRuns: 1,
		Logger: func(format string, args ...any) {
			logged = append(logged, fmt.Sprintf(format, args...))
		},
	})
	test := func(_ context.Context, ops []TraceOp) (Outcome, error) {
		return OutcomePass, nil
	}

	result, _, err := m.reduceAddresses(context.Background(), ops, meta, test)
	require.NoError(t, err)
	require.Equal(t, ops, result)
	require.NotEmpty(t, logged)
}
