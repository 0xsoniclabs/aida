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
	proxies["deletion"] = NewDeletionProxy(base, delChan, "info")
	wg := new(sync.WaitGroup)
	proxies["logger"] = NewLoggerProxy(proxies["deletion"], logger.NewLogger("info", "Proxy Logger"), logChan, wg)
	proxies["profiler"] = NewProfilerProxy(proxies["logger"], analytics.NewIncrementalAnalytics(len(operation.CreateIdLabelMap())), "info")
	traceFile := t.TempDir() + "trace"
	recordCtx, err := context.NewRecord(traceFile, 0)
	assert.NoError(t, err, "failed to create record context")
	proxies["recorder"] = NewRecorderProxy(proxies["profiler"], recordCtx)
	proxies["shadow"] = NewShadowProxy(base, proxies["recorder"], true)
	return proxies
}

// TODO test all proxy calls

func TestProxies_GetLogs(t *testing.T) {
	ctrl := gomock.NewController(t)
	base := state.NewMockStateDB(ctrl)
	proxies := getAllProxyImpls(t, base)
	hash := common.Hash{0x12}
	blk := uint64(2)
	blkHash := common.Hash{2}
	blkTimestamp := uint64(13)
	base.EXPECT().GetLogs(hash, blk, blkHash, blkTimestamp).Times(len(proxies) + 1) // +1 because shadow proxy calls twice
	for name, proxy := range proxies {
		t.Run(name, func(t *testing.T) {
			proxy.GetLogs(hash, blk, blkHash, blkTimestamp)
		})
	}
}
