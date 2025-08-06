package logger

import (
	"github.com/0xsoniclabs/aida/config"
	"testing"

	"github.com/0xsoniclabs/aida/ethtest"
	"github.com/0xsoniclabs/aida/executor"
	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/txcontext"
	"go.uber.org/mock/gomock"
)

func TestEthStateTestLogger_PreBlockLogsProgress(t *testing.T) {
	ctrl := gomock.NewController(t)
	log := logger.NewMockLogger(ctrl)
	ext := makeEthStateTestLogger(log, 2)
	s := executor.State[txcontext.TxContext]{Data: ethtest.CreateTestTransaction(t)}

	gomock.InOrder(
		log.EXPECT().Infof("Currently running:\n%s", s.Data),
		log.EXPECT().Infof("Currently running:\n%s", s.Data),
		log.EXPECT().Noticef("%v tests has been processed so far...", 2),
	)

	err := ext.PreBlock(s, &executor.Context{})
	if err != nil {
		t.Fatalf("pre-tx failed: %v", err)
	}

	err = ext.PreBlock(s, &executor.Context{})
	if err != nil {
		t.Fatalf("pre-tx failed: %v", err)
	}
}

func TestEthStateTestLogger_PostRunLogsOverall(t *testing.T) {
	ctrl := gomock.NewController(t)
	log := logger.NewMockLogger(ctrl)
	ext := makeEthStateTestLogger(log, 0)
	ext.overall = 2
	s := executor.State[txcontext.TxContext]{Data: ethtest.CreateTestTransaction(t)}

	gomock.InOrder(
		log.EXPECT().Noticef("Total %v tests processed.", 2),
	)

	err := ext.PostRun(s, &executor.Context{}, nil)
	if err != nil {
		t.Fatalf("post-run failed: %v", err)
	}

}

func TestEthStateTestLogger_MakeEthStateTestLogger(t *testing.T) {
	cfg := &config.Config{}
	ext := MakeEthStateTestLogger(cfg, 0)

	if _, ok := ext.(*ethStateTestLogger); !ok {
		t.Fatal("unexpected extension type")
	}
	if ext.(*ethStateTestLogger).reportFrequency != defaultReportFrequency {
		t.Fatal("default report frequency is not set")
	}
}
