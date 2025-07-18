package proxy

import (
	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/state"
	"github.com/0xsoniclabs/aida/tracer/context"
	"github.com/0xsoniclabs/aida/tracer/operation"
	"github.com/0xsoniclabs/aida/utils/analytics"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"sync"
	"testing"
)

func getAllProxyImpls(t *testing.T, base state.StateDB) map[string]state.StateDB {
	t.Helper()

	delChan := make(chan ContractLiveliness, 10)
	logChan := make(chan string, 10)
	// discard everything
	go func() {
		for {
			<-delChan
			<-logChan
		}
	}()
	proxies := make(map[string]state.StateDB)
	proxies["Deletion"] = NewDeletionProxy(base, delChan, "info")
	wg := new(sync.WaitGroup)
	proxies["Logger"] = NewLoggerProxy(base, logger.NewLogger("info", "Proxy Logger"), logChan, wg)
	proxies["Profiler"] = NewProfilerProxy(base, analytics.NewIncrementalAnalytics(len(operation.CreateIdLabelMap())), "info")
	traceFile := t.TempDir() + "trace"
	recordCtx, err := context.NewRecord(traceFile, 0)
	assert.NoError(t, err, "failed to create record context")
	proxies["Recorder"] = NewRecorderProxy(base, recordCtx)
	proxies["Shadow"] = NewShadowProxy(base, base, true)
	return proxies
}

// TODO test all proxy calls

func TestProxies_GetLogs(t *testing.T) {
	ctrl := gomock.NewController(t)
	base := state.NewMockStateDB(ctrl)
	proxies := getAllProxyImpls(t, base)
	addr := common.Address{0x11}
	hash := common.Hash{0x12}
	key := common.Hash{0x13}
	val := common.Hash{0x14}
	blk := uint64(2)
	blkHash := common.Hash{2}
	blkTimestamp := uint64(13)

	// everything is .Times(len+1) because shadow proxy calls twice (prime + shadow)
	base.EXPECT().CreateAccount(addr).Times(len(proxies) + 1)
	for name, proxy := range proxies {
		t.Run(name+"_CreateAccount", func(t *testing.T) {
			proxy.CreateAccount(addr)
		})
	}
	base.EXPECT().CreateContract(addr).Times(len(proxies) + 1)
	for name, proxy := range proxies {
		t.Run(name+"_CreateContract", func(t *testing.T) {
			proxy.CreateContract(addr)
		})
	}
	base.EXPECT().Exist(addr).Times(len(proxies) + 1)
	for name, proxy := range proxies {
		t.Run(name+"_CreateAccount", func(t *testing.T) {
			proxy.Exist(addr)
		})
	}
	base.EXPECT().GetLogs(hash, blk, blkHash, blkTimestamp).Times(len(proxies) + 1)
	for name, proxy := range proxies {
		t.Run(name+"_GetLogs", func(t *testing.T) {
			proxy.GetLogs(hash, blk, blkHash, blkTimestamp)
		})
	}
	base.EXPECT().SetTransientState(addr, key, val).Times(len(proxies) + 1)
	for name, proxy := range proxies {
		t.Run(name+"_SetTransientState", func(t *testing.T) {
			proxy.SetTransientState(addr, key, val)
		})
	}
}
