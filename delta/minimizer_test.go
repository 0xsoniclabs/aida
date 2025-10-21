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
