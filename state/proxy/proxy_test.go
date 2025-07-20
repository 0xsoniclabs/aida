package proxy

import (
	"sync"
	"testing"

	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/state"
	"github.com/0xsoniclabs/aida/tracer/context"
	"github.com/0xsoniclabs/aida/tracer/operation"
	"github.com/0xsoniclabs/aida/txcontext"
	"github.com/0xsoniclabs/aida/utils/analytics"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func getAllProxyImpls(t *testing.T, base state.StateDB) map[string]state.StateDB {
	t.Helper()

	delChan := make(chan ContractLiveliness, 10)
	logChan := make(chan string, 10)
	// discard everything
	go func() {
		for {
			select {
			case <-delChan:
			case <-logChan:
			}

		}
	}()
	proxies := make(map[string]state.StateDB)
	proxies["Deletion"] = NewDeletionProxy(base, delChan, "CRITICAL")
	wg := new(sync.WaitGroup)
	proxies["Logger"] = NewLoggerProxy(base, logger.NewLogger("CRITICAL", "Proxy Logger"), logChan, wg)
	proxies["Profiler"] = NewProfilerProxy(base, analytics.NewIncrementalAnalytics(len(operation.CreateIdLabelMap())), "info")
	traceFile := t.TempDir() + "trace"
	recordCtx, err := context.NewRecord(traceFile, 0)
	assert.NoError(t, err, "failed to create record context")
	proxies["Recorder"] = NewRecorderProxy(base, recordCtx)
	proxies["Shadow"] = NewShadowProxy(base, base, true)
	return proxies
}

// TODO test all proxy calls

func TestProxies_AllCalls(t *testing.T) {
	ctrl := gomock.NewController(t)
	base := state.NewMockStateDB(ctrl)
	proxies := getAllProxyImpls(t, base)
	addr := common.Address{0x11}
	hash := common.Hash{0x12}
	hash2 := common.Hash{0x22}
	key := common.Hash{0x13}
	val := common.Hash{0x14}
	uint64Val := uint64(42)
	intVal := 0
	code := []byte{0x01, 0x02}
	log := &types.Log{}
	image := []byte{0xAA}
	boolVal := true
	amount := uint256.NewInt(1000)
	balanceReason := tracing.BalanceChangeUnspecified
	nonceReason := tracing.NonceChangeUnspecified
	ws := txcontext.NewWorldState(make(map[common.Address]txcontext.Account))

	// CreateContract
	base.EXPECT().CreateContract(addr).Times(len(proxies) + 1)
	for name, proxy := range proxies {
		t.Run(name+"_CreateContract", func(t *testing.T) {
			proxy.CreateContract(addr)
		})
	}

	// SubBalance
	base.EXPECT().SubBalance(addr, amount, balanceReason).Times(len(proxies) + 1)
	// LoggerProxy calls GetBalance to log current balance
	base.EXPECT().GetBalance(addr)
	for name, proxy := range proxies {
		t.Run(name+"_SubBalance", func(t *testing.T) {
			proxy.SubBalance(addr, amount, balanceReason)
		})
	}

	// AddBalance
	// LoggerProxy calls GetBalance to log current balance
	base.EXPECT().GetBalance(addr)
	base.EXPECT().AddBalance(addr, amount, balanceReason).Times(len(proxies) + 1)
	for name, proxy := range proxies {
		t.Run(name+"_AddBalance", func(t *testing.T) {
			proxy.AddBalance(addr, amount, balanceReason)
		})
	}

	// GetBalance
	base.EXPECT().GetBalance(addr).Times(len(proxies) + 1).Return(amount)
	for name, proxy := range proxies {
		t.Run(name+"_GetBalance", func(t *testing.T) {
			proxy.GetBalance(addr)
		})
	}

	// GetNonce
	base.EXPECT().GetNonce(addr).Times(len(proxies) + 1)
	for name, proxy := range proxies {
		t.Run(name+"_GetNonce", func(t *testing.T) {
			proxy.GetNonce(addr)
		})
	}

	// SetNonce
	base.EXPECT().SetNonce(addr, uint64Val, nonceReason).Times(len(proxies) + 1)
	for name, proxy := range proxies {
		t.Run(name+"SetNonce", func(t *testing.T) {
			proxy.SetNonce(addr, uint64Val, nonceReason)
		})
	}

	// SelfDestruct
	base.EXPECT().SelfDestruct(addr).Times(len(proxies) + 1)
	for name, proxy := range proxies {
		t.Run(name+"_SelfDestruct", func(t *testing.T) {
			proxy.SelfDestruct(addr)
		})
	}

	// SelfDestruct6780
	base.EXPECT().SelfDestruct6780(addr).Times(len(proxies) + 1)
	for name, proxy := range proxies {
		t.Run(name+"_SelfDestruct6780", func(t *testing.T) {
			proxy.SelfDestruct6780(addr)
		})
	}

	// HasSelfDestructed
	base.EXPECT().HasSelfDestructed(addr).Times(len(proxies) + 1)
	for name, proxy := range proxies {
		t.Run(name+"_HasSelfDestructed", func(t *testing.T) {
			proxy.HasSelfDestructed(addr)
		})
	}

	// Empty
	base.EXPECT().Exist(addr).Times(len(proxies) + 1)
	for name, proxy := range proxies {
		t.Run(name+"_Exist", func(t *testing.T) {
			proxy.Exist(addr)
		})
	}

	// Empty
	base.EXPECT().Empty(addr).Times(len(proxies) + 1)
	for name, proxy := range proxies {
		t.Run(name+"_Empty", func(t *testing.T) {
			proxy.Empty(addr)
		})
	}

	// GetCommittedState
	base.EXPECT().GetCommittedState(addr, key).Times(len(proxies) + 1)
	for name, proxy := range proxies {
		t.Run(name+"_GetCommittedState", func(t *testing.T) {
			proxy.GetCommittedState(addr, key)
		})
	}

	// Commit
	base.EXPECT().Commit(uint64Val, boolVal).Times(len(proxies) + 1)
	for name, proxy := range proxies {
		t.Run(name+"_Commit", func(t *testing.T) {
			proxy.Commit(uint64Val, boolVal)
		})
	}

	// GetState
	base.EXPECT().GetState(addr, key).Times(len(proxies) + 1)
	for name, proxy := range proxies {
		t.Run(name+"_GetState", func(t *testing.T) {
			proxy.GetState(addr, key)
		})
	}

	// GetHash
	base.EXPECT().GetHash().Times(len(proxies) + 1)
	for name, proxy := range proxies {
		t.Run(name+"_GetHash", func(t *testing.T) {
			proxy.GetHash()
		})
	}

	// SetState
	base.EXPECT().SetState(addr, key, val).Times(len(proxies) + 1)
	for name, proxy := range proxies {
		t.Run(name+"_SetState", func(t *testing.T) {
			proxy.SetState(addr, key, val)
		})
	}

	// GetStorageRoot
	base.EXPECT().GetStorageRoot(addr).Times(len(proxies) + 1)
	for name, proxy := range proxies {
		t.Run(name+"_GetStorageRoot", func(t *testing.T) {
			proxy.GetStorageRoot(addr)
		})
	}

	// Error
	// ShadowProxy does not implement this method
	base.EXPECT().Error().Times(len(proxies) - 1)
	for name, proxy := range proxies {
		t.Run(name+"_Error", func(t *testing.T) {
			proxy.Error()
		})
	}

	// SetTransientState
	base.EXPECT().SetTransientState(addr, key, val).Times(len(proxies) + 1)
	for name, proxy := range proxies {
		t.Run(name+"_SetTransientState", func(t *testing.T) {
			proxy.SetTransientState(addr, key, val)
		})
	}

	// GetTransientState
	base.EXPECT().GetTransientState(addr, key).Times(len(proxies) + 1)
	for name, proxy := range proxies {
		t.Run(name+"_GetTransientState", func(t *testing.T) {
			proxy.GetTransientState(addr, key)
		})
	}

	// GetCodeHash
	base.EXPECT().GetCodeHash(addr).Times(len(proxies) + 1)
	for name, proxy := range proxies {
		t.Run(name+"_GetCodeHash", func(t *testing.T) {
			proxy.GetCodeHash(addr)
		})
	}

	// GetCode
	base.EXPECT().GetCode(addr).Times(len(proxies) + 1)
	for name, proxy := range proxies {
		t.Run(name+"_GetCode", func(t *testing.T) {
			proxy.GetCode(addr)
		})
	}

	// SetCode
	base.EXPECT().SetCode(addr, code).Times(len(proxies) + 1)
	for name, proxy := range proxies {
		t.Run(name+"_SetCode", func(t *testing.T) {
			proxy.SetCode(addr, code)
		})
	}

	// GetCodeSize
	base.EXPECT().GetCodeSize(addr).Times(len(proxies) + 1)
	for name, proxy := range proxies {
		t.Run(name+"_GetCodeSize", func(t *testing.T) {
			proxy.GetCodeSize(addr)
		})
	}

	// AddRefund
	base.EXPECT().AddRefund(uint64Val).Times(len(proxies) + 1)
	// ShadowProxy and Logger calls GetRefund to check state correctness
	base.EXPECT().GetRefund().Times(3)
	for name, proxy := range proxies {
		t.Run(name+"_AddRefund", func(t *testing.T) {
			proxy.AddRefund(uint64Val)
		})
	}

	// SubRefund
	base.EXPECT().SubRefund(uint64Val).Times(len(proxies) + 1)
	// ShadowProxy and Logger calls GetRefund to check state correctness
	base.EXPECT().GetRefund().Times(3)
	for name, proxy := range proxies {
		t.Run(name+"_SubRefund", func(t *testing.T) {
			proxy.SubRefund(uint64Val)
		})
	}

	// GetRefund
	base.EXPECT().GetRefund().Times(len(proxies) + 1)
	for name, proxy := range proxies {
		t.Run(name+"_GetRefund", func(t *testing.T) {
			proxy.GetRefund()
		})
	}

	// Prepare
	base.EXPECT().Prepare(gomock.Any(), addr, addr, gomock.Nil(), gomock.Any(), gomock.Any()).Times(len(proxies) + 1)
	for name, proxy := range proxies {
		t.Run(name+"_Prepare", func(t *testing.T) {
			proxy.Prepare(params.Rules{}, addr, addr, nil, []common.Address{}, types.AccessList{})
		})
	}

	// AddressInAccessList
	base.EXPECT().AddressInAccessList(addr).Times(len(proxies) + 1)
	for name, proxy := range proxies {
		t.Run(name+"_AddressInAccessList", func(t *testing.T) {
			proxy.AddressInAccessList(addr)
		})
	}

	// SlotInAccessList
	base.EXPECT().SlotInAccessList(addr, key).Times(len(proxies) + 1)
	for name, proxy := range proxies {
		t.Run(name+"_SlotInAccessList", func(t *testing.T) {
			proxy.SlotInAccessList(addr, key)
		})
	}

	// AddAddressToAccessList
	base.EXPECT().AddAddressToAccessList(addr).Times(len(proxies) + 1)
	for name, proxy := range proxies {
		t.Run(name+"_AddAddressToAccessList", func(t *testing.T) {
			proxy.AddAddressToAccessList(addr)
		})
	}

	// AddSlotToAccessList
	base.EXPECT().AddSlotToAccessList(addr, key).Times(len(proxies) + 1)
	for name, proxy := range proxies {
		t.Run(name+"_AddSlotToAccessList", func(t *testing.T) {
			proxy.AddSlotToAccessList(addr, key)
		})
	}

	// AddLog
	base.EXPECT().AddLog(log).Times(len(proxies) + 1)
	for name, proxy := range proxies {
		t.Run(name+"_AddLog", func(t *testing.T) {
			proxy.AddLog(log)
		})
	}

	// GetLogs
	base.EXPECT().GetLogs(hash, uint64Val, hash2, uint64Val).Times(len(proxies) + 1)
	for name, proxy := range proxies {
		t.Run(name+"_GetLogs", func(t *testing.T) {
			proxy.GetLogs(hash, uint64Val, hash2, uint64Val)
		})
	}

	// Snapshot
	base.EXPECT().Snapshot().Times(len(proxies) + 1)
	for name, proxy := range proxies {
		t.Run(name+"_Snapshot", func(t *testing.T) {
			proxy.Snapshot()
		})
	}

	// RevertToSnapshot
	base.EXPECT().RevertToSnapshot(intVal).Times(len(proxies) + 1)
	for name, proxy := range proxies {
		t.Run(name+"_RevertToSnapshot", func(t *testing.T) {
			proxy.RevertToSnapshot(intVal)
		})
	}

	// PointCache
	// ShadowProxy calls this only on Prime todo is this correct?
	base.EXPECT().PointCache().Times(len(proxies))
	for name, proxy := range proxies {
		t.Run(name+"_PointCache", func(t *testing.T) {
			proxy.PointCache()
		})
	}

	// Witness
	// ShadowProxy calls this only on Prime todo is this correct?
	base.EXPECT().Witness().Times(len(proxies))
	for name, proxy := range proxies {
		t.Run(name+"_Witness", func(t *testing.T) {
			proxy.Witness()
		})
	}

	// SetTxContext
	base.EXPECT().SetTxContext(hash, int(uint64Val)).Times(len(proxies) + 1)
	for name, proxy := range proxies {
		t.Run(name+"_SetTxContext", func(t *testing.T) {
			proxy.SetTxContext(hash, int(uint64Val))
		})
	}

	// Finalise
	base.EXPECT().Finalise(boolVal).Times(len(proxies) + 1)
	for name, proxy := range proxies {
		t.Run(name+"_Finalise", func(t *testing.T) {
			proxy.Finalise(boolVal)
		})
	}

	// IntermediateRoot
	base.EXPECT().IntermediateRoot(boolVal).Times(len(proxies) + 1)
	for name, proxy := range proxies {
		t.Run(name+"_IntermediateRoot", func(t *testing.T) {
			proxy.IntermediateRoot(boolVal)
		})
	}

	// AddPreimage
	base.EXPECT().AddPreimage(hash, image).Times(len(proxies) + 1)
	for name, proxy := range proxies {
		t.Run(name+"_AddPreimage", func(t *testing.T) {
			proxy.AddPreimage(hash, image)
		})
	}

	// AccessEvents
	// ShadowProxy calls this only on Prime todo is this correct?
	base.EXPECT().AccessEvents().Times(len(proxies))
	for name, proxy := range proxies {
		t.Run(name+"_AccessEvents", func(t *testing.T) {
			proxy.AccessEvents()
		})
	}

	// GetSubstatePostAlloc
	base.EXPECT().GetSubstatePostAlloc().Times(len(proxies) + 1).Return(ws)
	for name, proxy := range proxies {
		t.Run(name+"_GetSubstatePostAlloc", func(t *testing.T) {
			proxy.GetSubstatePostAlloc()
		})
	}

	// PrepareSubstate
	base.EXPECT().PrepareSubstate(ws, uint64Val).Times(len(proxies) + 1)
	for name, proxy := range proxies {
		t.Run(name+"_PrepareSubstate", func(t *testing.T) {
			proxy.PrepareSubstate(ws, uint64Val)
		})
	}

	// BeginBlock
	base.EXPECT().BeginBlock(uint64Val).Times(len(proxies) + 1)
	for name, proxy := range proxies {
		t.Run(name+"_BeginBlock", func(t *testing.T) {
			proxy.BeginBlock(uint64Val)
		})
	}

	// BeginTransaction
	txNum := uint32(55)
	base.EXPECT().BeginTransaction(txNum).Times(len(proxies) + 1)
	for name, proxy := range proxies {
		t.Run(name+"_BeginTransaction", func(t *testing.T) {
			proxy.BeginTransaction(txNum)
		})
	}

	// EndTransaction
	base.EXPECT().EndTransaction().Times(len(proxies) + 1)
	for name, proxy := range proxies {
		t.Run(name+"_EndTransaction", func(t *testing.T) {
			proxy.EndTransaction()
		})
	}

	// EndBlock
	base.EXPECT().EndBlock().Times(len(proxies) + 1)
	for name, proxy := range proxies {
		t.Run(name+"_EndBlock", func(t *testing.T) {
			proxy.EndBlock()
		})
	}

	// BeginSyncPeriod
	base.EXPECT().BeginSyncPeriod(uint64Val).Times(len(proxies) + 1)
	for name, proxy := range proxies {
		t.Run(name+"_BeginSyncPeriod", func(t *testing.T) {
			proxy.BeginSyncPeriod(uint64Val)
		})
	}

	// EndSyncPeriod
	base.EXPECT().EndSyncPeriod().Times(len(proxies) + 1)
	for name, proxy := range proxies {
		t.Run(name+"_EndSyncPeriod", func(t *testing.T) {
			proxy.EndSyncPeriod()
		})
	}

	// Close
	base.EXPECT().Close().Times(len(proxies) + 1)
	for name, proxy := range proxies {
		t.Run(name+"_Close", func(t *testing.T) {
			proxy.Close()
		})
	}

	// GetMemoryUsage
	base.EXPECT().GetMemoryUsage().Times(len(proxies) + 1)
	for name, proxy := range proxies {
		t.Run(name+"_GetMemoryUsage", func(t *testing.T) {
			proxy.GetMemoryUsage()
		})
	}

	// GetShadowDB
	// ShadowProxy does not implement this method
	base.EXPECT().GetShadowDB().Times(len(proxies) - 1)
	for name, proxy := range proxies {
		t.Run(name+"_GetShadowDB", func(t *testing.T) {
			proxy.GetShadowDB()
		})
	}
}
