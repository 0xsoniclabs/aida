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
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/rpc"
	"github.com/0xsoniclabs/aida/utils"
	"go.uber.org/mock/gomock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func TestRPCRequestProvider_WorksWithValidResponse(t *testing.T) {
	ctrl := gomock.NewController(t)
	consumer := NewMockRPCReqConsumer(ctrl)
	i := rpc.NewMockIterator(ctrl)

	cfg := &utils.Config{}

	provider := openRpcRecording(i, cfg, logger.NewLogger("critical", "rpc-provider-test"), nil, []string{"testfile"})

	defer provider.Close()

	gomock.InOrder(
		i.EXPECT().Next().Return(true),
		i.EXPECT().Error().Return(nil),
		i.EXPECT().Value().Return(validResp),
		consumer.EXPECT().Consume(10, gomock.Any(), validResp),
		i.EXPECT().Next().Return(false),
		i.EXPECT().Close(),
	)

	if err := provider.Run(10, 11, toRPCConsumer(consumer)); err != nil {
		t.Fatalf("failed to iterate through requests: %v", err)
	}
}

func TestRPCRequestProvider_WorksWithErrorResponse(t *testing.T) {
	ctrl := gomock.NewController(t)
	consumer := NewMockRPCReqConsumer(ctrl)
	i := rpc.NewMockIterator(ctrl)

	cfg := &utils.Config{}

	provider := openRpcRecording(i, cfg, logger.NewLogger("critical", "rpc-provider-test"), nil, []string{"testfile"})

	defer provider.Close()

	gomock.InOrder(
		i.EXPECT().Next().Return(true),
		i.EXPECT().Error().Return(nil),
		i.EXPECT().Value().Return(errResp),
		consumer.EXPECT().Consume(10, gomock.Any(), errResp),
		i.EXPECT().Next().Return(false),
		i.EXPECT().Close(),
	)

	if err := provider.Run(10, 11, toRPCConsumer(consumer)); err != nil {
		t.Fatalf("failed to iterate through requests: %v", err)
	}
}

func TestRPCRequestProvider_NilRequestDoesNotGetToConsumer(t *testing.T) {
	ctrl := gomock.NewController(t)
	consumer := NewMockRPCReqConsumer(ctrl)
	i := rpc.NewMockIterator(ctrl)

	cfg := &utils.Config{}

	provider := openRpcRecording(i, cfg, logger.NewLogger("critical", "rpc-provider-test"), nil, []string{"testfile"})

	defer provider.Close()

	gomock.InOrder(
		i.EXPECT().Next().Return(true),
		i.EXPECT().Error().Return(nil),
		i.EXPECT().Value().Return(nil),
		i.EXPECT().Close(),
	)

	err := provider.Run(10, 11, toRPCConsumer(consumer))
	if err == nil {
		t.Fatal("provider must return error")
	}

	got := err.Error()
	want := "iterator returned nil request"

	if strings.Compare(got, want) != 0 {
		t.Fatalf("unexpected error\ngot: %v\nwant:%v", got, want)
	}

}

func TestRPCRequestProvider_ErrorReturnedByIteratorEndsTheApp(t *testing.T) {
	ctrl := gomock.NewController(t)
	consumer := NewMockRPCReqConsumer(ctrl)
	i := rpc.NewMockIterator(ctrl)

	cfg := &utils.Config{}

	provider := openRpcRecording(i, cfg, logger.NewLogger("critical", "rpc-provider-test"), nil, []string{"testfile"})

	defer provider.Close()

	gomock.InOrder(
		i.EXPECT().Next().Return(true),
		i.EXPECT().Error().Return(errors.New("err")),
		i.EXPECT().Error().Return(errors.New("err")),
		i.EXPECT().Close(),
	)

	if err := provider.Run(10, 11, toRPCConsumer(consumer)); err == nil {
		if strings.Compare(err.Error(), "iterator returned error; err") != 0 {
			t.Fatal("unexpected error returned by the iterator")
		}
		t.Fatal("the test should return an error")
	}
}

func TestRPCRequestProvider_GetLogMethodDoesNotEndIteration(t *testing.T) {
	ctrl := gomock.NewController(t)
	consumer := NewMockRPCReqConsumer(ctrl)
	i := rpc.NewMockIterator(ctrl)

	cfg := &utils.Config{}

	provider := openRpcRecording(i, cfg, logger.NewLogger("critical", "rpc-provider-test"), nil, []string{"testfile"})

	defer provider.Close()

	gomock.InOrder(
		i.EXPECT().Next().Return(true),
		i.EXPECT().Error().Return(nil),
		i.EXPECT().Value().Return(logResp),
		i.EXPECT().Next().Return(true),
		i.EXPECT().Error().Return(nil),
		i.EXPECT().Value().Return(logResp),
		i.EXPECT().Next().Return(false),
		i.EXPECT().Close(),
	)

	if err := provider.Run(10, 11, toRPCConsumer(consumer)); err != nil {
		t.Fatal("test cannot fail")
	}
}

func TestRPCRequestProvider_ReportsAboutRun(t *testing.T) {
	ctrl := gomock.NewController(t)
	consumer := NewMockRPCReqConsumer(ctrl)
	log := logger.NewMockLogger(ctrl)
	i := rpc.NewMockIterator(ctrl)

	cfg := &utils.Config{}
	cfg.RpcRecordingPath = "test_file"

	provider := openRpcRecording(i, cfg, log, nil, []string{cfg.RpcRecordingPath})

	defer provider.Close()

	gomock.InOrder(
		i.EXPECT().Next().Return(true),
		i.EXPECT().Error().Return(nil),
		i.EXPECT().Value().Return(validResp),
		log.EXPECT().Noticef("Iterating file %v/%v path: %v", 1, 1, "test_file"),
		log.EXPECT().Noticef("First block of recording: %v", 10),
		consumer.EXPECT().Consume(10, gomock.Any(), validResp).Return(errors.New("err")),
		log.EXPECT().Infof("Last iterated file: %v", "test_file"),

		//i.EXPECT().Next().Return(false),
		i.EXPECT().Close(),
	)

	if err := provider.Run(10, 11, toRPCConsumer(consumer)); err == nil {
		t.Fatal("run must fail")
	}
}

var validResp = &rpc.RequestAndResults{
	Query: &rpc.Body{},
	Response: &rpc.Response{
		Version:   "2.0",
		ID:        json.RawMessage{1},
		BlockID:   10,
		Timestamp: 10,
		Result:    nil,
		Payload:   nil,
	},
	Error:       nil,
	ParamsRaw:   nil,
	ResponseRaw: nil,
}

var errResp = &rpc.RequestAndResults{
	Query:    &rpc.Body{},
	Response: nil,
	Error: &rpc.ErrorResponse{
		Version:   "2.0",
		Id:        json.RawMessage{1},
		BlockID:   10,
		Timestamp: 10,
		Error: rpc.ErrorMessage{
			Code:    -1,
			Message: "err",
		},
		Payload: nil,
	},
	ParamsRaw:   nil,
	ResponseRaw: nil,
}

var logResp = &rpc.RequestAndResults{
	Query: &rpc.Body{
		MethodBase: "getLogs",
	},
	Response: &rpc.Response{
		Version:   "2.0",
		ID:        json.RawMessage{1},
		BlockID:   10,
		Timestamp: 10,
		Result:    nil,
		Payload:   nil,
	},
	Error:       nil,
	ParamsRaw:   nil,
	ResponseRaw: nil,
}

// Helper to create a cli.Context with a background context
func newTestCliContext() *cli.Context {
	return &cli.Context{Context: context.Background()}
}

func TestOpenRpcRecording(t *testing.T) {
	baseDir := t.TempDir()
	cliCtx := newTestCliContext()

	t.Run("Success_SingleFile", func(t *testing.T) {
		tmpFile, err := os.CreateTemp(baseDir, "singlefile*.rpc")
		require.NoError(t, err)
		_, err = tmpFile.WriteString("dummy rpc content")
		require.NoError(t, err)
		err = tmpFile.Close()
		require.NoError(t, err)

		cfg := &utils.Config{RpcRecordingPath: tmpFile.Name()}

		provider, err := OpenRpcRecording(cfg, cliCtx)
		require.NoError(t, err)
		require.NotNil(t, provider)
		defer provider.Close()

		rpcProv, ok := provider.(*rpcRequestProvider)
		require.True(t, ok, "Provider should be of type *rpcRequestProvider")
		assert.Equal(t, cfg.RpcRecordingPath, rpcProv.fileName)
		assert.Equal(t, []string{tmpFile.Name()}, rpcProv.files)
		assert.NotNil(t, rpcProv.iter)
	})

	t.Run("Success_DirectoryWithOneFile", func(t *testing.T) {
		tmpDir := filepath.Join(baseDir, "singledir")
		require.NoError(t, os.Mkdir(tmpDir, 0755))
		tmpFile, err := os.Create(filepath.Join(tmpDir, "file1.rpc"))
		require.NoError(t, err)
		_, err = tmpFile.WriteString("dummy rpc content")
		require.NoError(t, err)
		err = tmpFile.Close()
		require.NoError(t, err)

		cfg := &utils.Config{RpcRecordingPath: tmpDir}

		provider, err := OpenRpcRecording(cfg, cliCtx)
		require.NoError(t, err)
		require.NotNil(t, provider)
		defer provider.Close()

		rpcProv, ok := provider.(*rpcRequestProvider)
		require.True(t, ok)
		assert.Equal(t, cfg.RpcRecordingPath, rpcProv.fileName)
		assert.Equal(t, []string{tmpFile.Name()}, rpcProv.files)
		assert.NotNil(t, rpcProv.iter)
	})

	t.Run("Success_DirectoryWithMultipleFiles", func(t *testing.T) {
		tmpDir := filepath.Join(baseDir, "multidir")
		require.NoError(t, os.Mkdir(tmpDir, 0755))

		file1, err := os.Create(filepath.Join(tmpDir, "file1.rpc"))
		require.NoError(t, err)
		_, err = file1.WriteString("dummy content1")
		require.NoError(t, err)
		err = file1.Close()
		require.NoError(t, err)

		file2, err := os.Create(filepath.Join(tmpDir, "file2.rpc"))
		require.NoError(t, err)
		_, err = file2.WriteString("dummy content2")
		require.NoError(t, err)
		err = file2.Close()
		require.NoError(t, err)

		expectedFiles := []string{file1.Name(), file2.Name()}
		sort.Strings(expectedFiles)

		cfg := &utils.Config{RpcRecordingPath: tmpDir}

		provider, err := OpenRpcRecording(cfg, cliCtx)
		require.NoError(t, err)
		require.NotNil(t, provider)
		defer provider.Close()

		rpcProv, ok := provider.(*rpcRequestProvider)
		require.True(t, ok)
		assert.Equal(t, cfg.RpcRecordingPath, rpcProv.fileName)

		actualFiles := make([]string, len(rpcProv.files))
		copy(actualFiles, rpcProv.files)
		sort.Strings(actualFiles)
		assert.Equal(t, expectedFiles, actualFiles)
		assert.NotNil(t, rpcProv.iter)
	})

	t.Run("Error_PathNotFound", func(t *testing.T) {
		cfg := &utils.Config{RpcRecordingPath: filepath.Join(baseDir, "nonexistent.rpc")}
		provider, err := OpenRpcRecording(cfg, cliCtx)

		assert.Error(t, err)
		assert.Nil(t, provider)
		assert.Contains(t, err.Error(), "cannot stat the rpc path")
	})

	t.Run("Error_RpcNewFileReader_SingleFile_Unreadable", func(t *testing.T) {
		tmpFile, err := os.CreateTemp(baseDir, "unreadable*.rpc")
		require.NoError(t, err)
		err = tmpFile.Close()
		require.NoError(t, err)

		err = os.Chmod(tmpFile.Name(), 0000)
		if err != nil {
			t.Logf("Skipping unreadable test, could not chmod: %v", err)
			t.SkipNow()
		}
		defer func(name string, mode os.FileMode) {
			e := os.Chmod(name, mode)
			if e != nil {
				t.Fatal(e)
			}
		}(tmpFile.Name(), 0644)

		cfg := &utils.Config{RpcRecordingPath: tmpFile.Name()}
		provider, err := OpenRpcRecording(cfg, cliCtx)

		assert.Error(t, err)
		assert.Nil(t, provider)

		assert.Contains(t, err.Error(), "cannot open rpc recording file")
	})

	t.Run("Error_RpcNewFileReader_Directory_FirstFileUnreadable", func(t *testing.T) {
		tmpDir := filepath.Join(baseDir, "dirunreadable")
		require.NoError(t, os.Mkdir(tmpDir, 0755))
		tmpFile, err := os.Create(filepath.Join(tmpDir, "unreadablefile.rpc"))
		require.NoError(t, err)
		err = tmpFile.Close()
		if err != nil {
			t.Fatal(err)
		}

		err = os.Chmod(tmpFile.Name(), 0000)
		if err != nil {
			t.Logf("Skipping unreadable test, could not chmod: %v", err)
			t.SkipNow()
		}
		defer func(name string, mode os.FileMode) {
			e := os.Chmod(name, mode)
			if e != nil {
				t.Fatal(e)
			}
		}(tmpFile.Name(), 0644)

		cfg := &utils.Config{RpcRecordingPath: tmpDir}
		provider, err := OpenRpcRecording(cfg, cliCtx)

		assert.Error(t, err)
		assert.Nil(t, provider)
		assert.Contains(t, err.Error(), "cannot open rpc recording file")
		assert.Contains(t, err.Error(), tmpFile.Name())
	})

	t.Run("Panic_Directory_GetFilesReturnsEmptySlice", func(t *testing.T) {
		tmpEmptyDir := filepath.Join(baseDir, "emptydir_for_panic")
		require.NoError(t, os.Mkdir(tmpEmptyDir, 0755))

		cfg := &utils.Config{RpcRecordingPath: tmpEmptyDir}

		assert.PanicsWithError(t, "runtime error: index out of range [0] with length 0", func() {
			_, _ = OpenRpcRecording(cfg, cliCtx)
		}, "Expected panic for empty directory leading to files[0] access")
	})

}

func TestRpcRequestProvider_Run_SingleFile_AllItemsConsumed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockIter := rpc.NewMockIterator(ctrl)
	mockConsumer := NewMockRPCReqConsumer(ctrl)
	testLog := logger.NewLogger("info", "rpc-provider-test")
	cliCtx := newTestCliContext()

	cfg := &utils.Config{}

	provider := openRpcRecording(mockIter, cfg, testLog, cliCtx, []string{"file1.rpc"}).(*rpcRequestProvider)
	defer provider.Close()

	req1 := &rpc.RequestAndResults{Query: &rpc.Body{}, Response: &rpc.Response{BlockID: 5}}
	req2 := &rpc.RequestAndResults{Query: &rpc.Body{}, Response: &rpc.Response{BlockID: 10}}

	mockIter.EXPECT().Next().Return(true).Times(1)
	mockIter.EXPECT().Error().Return(nil).Times(1)
	mockIter.EXPECT().Value().Return(req1).Times(1)
	mockConsumer.EXPECT().Consume(5, 0, req1).Return(nil).Times(1)

	mockIter.EXPECT().Next().Return(true).Times(1)
	mockIter.EXPECT().Error().Return(nil).Times(1)
	mockIter.EXPECT().Value().Return(req2).Times(1)
	mockConsumer.EXPECT().Consume(10, 0, req2).Return(nil).Times(1)

	mockIter.EXPECT().Next().Return(false).Times(1)
	mockIter.EXPECT().Close()

	err := provider.Run(0, 20, toRPCConsumer(mockConsumer))
	assert.NoError(t, err)
}

func TestRpcRequestProvider_Run_FilteringWithFromAndTo(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockIter := rpc.NewMockIterator(ctrl)
	mockConsumer := NewMockRPCReqConsumer(ctrl)
	testLog := logger.NewLogger("info", "rpc-provider-test")
	cliCtx := newTestCliContext()
	provider := openRpcRecording(mockIter, &utils.Config{}, testLog, cliCtx, []string{"f1.rpc"}).(*rpcRequestProvider)
	defer provider.Close()

	req1 := &rpc.RequestAndResults{Query: &rpc.Body{}, Response: &rpc.Response{BlockID: 5}}
	req2 := &rpc.RequestAndResults{Query: &rpc.Body{}, Response: &rpc.Response{BlockID: 10}}
	req3 := &rpc.RequestAndResults{Query: &rpc.Body{}, Response: &rpc.Response{BlockID: 15}}
	req4 := &rpc.RequestAndResults{Query: &rpc.Body{}, Response: &rpc.Response{BlockID: 20}}

	mockIter.EXPECT().Next().Return(true)
	mockIter.EXPECT().Error().Return(nil)
	mockIter.EXPECT().Value().Return(req1)

	mockIter.EXPECT().Next().Return(true)
	mockIter.EXPECT().Error().Return(nil)
	mockIter.EXPECT().Value().Return(req2)
	mockConsumer.EXPECT().Consume((10), (0), req2).Return(nil)

	mockIter.EXPECT().Next().Return(true)
	mockIter.EXPECT().Error().Return(nil)
	mockIter.EXPECT().Value().Return(req3)
	mockConsumer.EXPECT().Consume(15, 0, req3).Return(nil)

	mockIter.EXPECT().Next().Return(true)
	mockIter.EXPECT().Error().Return(nil)
	mockIter.EXPECT().Value().Return(req4)

	mockIter.EXPECT().Close()

	err := provider.Run(10, 20, toRPCConsumer(mockConsumer))
	assert.NoError(t, err)
}

func TestRpcRequestProvider_Run_ConsumerErrorInProcessFirst(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockIter := rpc.NewMockIterator(ctrl)
	mockConsumer := NewMockRPCReqConsumer(ctrl)
	testLog := logger.NewLogger("info", "rpc-provider-test")
	cliCtx := newTestCliContext()
	provider := openRpcRecording(mockIter, &utils.Config{}, testLog, cliCtx, []string{"f1.rpc"}).(*rpcRequestProvider)
	defer provider.Close()

	req1 := &rpc.RequestAndResults{Query: &rpc.Body{}, Response: &rpc.Response{BlockID: 5}}
	consumerErr := errors.New("consumer failed in processFirst")

	mockIter.EXPECT().Next().Return(true)
	mockIter.EXPECT().Error().Return(nil)
	mockIter.EXPECT().Value().Return(req1)
	mockConsumer.EXPECT().Consume(5, 0, req1).Return(consumerErr)

	mockIter.EXPECT().Close()

	err := provider.Run(0, 10, toRPCConsumer(mockConsumer))
	assert.Error(t, err)
	assert.Equal(t, consumerErr, err)
}

func TestRpcRequestProvider_Run_ConsumerErrorInLoop(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockIter := rpc.NewMockIterator(ctrl)
	mockConsumer := NewMockRPCReqConsumer(ctrl)
	testLog := logger.NewLogger("info", "rpc-provider-test")
	cliCtx := newTestCliContext()
	provider := openRpcRecording(mockIter, &utils.Config{}, testLog, cliCtx, []string{"f1.rpc"}).(*rpcRequestProvider)
	defer provider.Close()

	req1 := &rpc.RequestAndResults{Query: &rpc.Body{}, Response: &rpc.Response{BlockID: 5}}
	req2 := &rpc.RequestAndResults{Query: &rpc.Body{}, Response: &rpc.Response{BlockID: 6}}
	consumerErr := errors.New("consumer failed in loop")

	mockIter.EXPECT().Next().Return(true)
	mockIter.EXPECT().Error().Return(nil)
	mockIter.EXPECT().Value().Return(req1)
	mockConsumer.EXPECT().Consume(5, 0, req1).Return(nil)

	mockIter.EXPECT().Next().Return(true)
	mockIter.EXPECT().Error().Return(nil)
	mockIter.EXPECT().Value().Return(req2)
	mockConsumer.EXPECT().Consume(6, 0, req2).Return(consumerErr) // Error here

	mockIter.EXPECT().Close()

	err := provider.Run(0, 10, toRPCConsumer(mockConsumer))
	assert.Error(t, err)
	assert.Equal(t, consumerErr, err)
}

func TestRpcRequestProvider_Run_IteratorErrorInLoop(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockIter := rpc.NewMockIterator(ctrl)
	mockConsumer := NewMockRPCReqConsumer(ctrl)
	testLog := logger.NewLogger("info", "rpc-provider-test")
	cliCtx := newTestCliContext()
	provider := openRpcRecording(mockIter, &utils.Config{}, testLog, cliCtx, []string{"f1.rpc"}).(*rpcRequestProvider)
	defer provider.Close()

	req1 := &rpc.RequestAndResults{Query: &rpc.Body{}, Response: &rpc.Response{BlockID: 5}}
	iterErr := errors.New("iterator failed")

	// processFirst
	mockIter.EXPECT().Next().Return(true).AnyTimes()
	mockIter.EXPECT().Error().Return(nil)
	mockIter.EXPECT().Value().Return(req1).AnyTimes()
	mockConsumer.EXPECT().Consume(5, 0, req1).Return(nil).AnyTimes()

	mockIter.EXPECT().Next().Return(true).AnyTimes()
	mockIter.EXPECT().Error().Return(iterErr).AnyTimes()

	mockIter.EXPECT().Close()

	err := provider.Run(0, 10, toRPCConsumer(mockConsumer))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "iterator returned error; iterator failed")
}

func TestRpcRequestProvider_Run_IteratorReturnsNilValueInLoop(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockIter := rpc.NewMockIterator(ctrl)
	mockConsumer := NewMockRPCReqConsumer(ctrl)
	testLog := logger.NewLogger("info", "rpc-provider-test")
	cliCtx := newTestCliContext()
	provider := openRpcRecording(mockIter, &utils.Config{}, testLog, cliCtx, []string{"f1.rpc"}).(*rpcRequestProvider)
	defer provider.Close()

	req1 := &rpc.RequestAndResults{Query: &rpc.Body{}, Response: &rpc.Response{BlockID: 5}}

	mockIter.EXPECT().Next().Return(true)
	mockIter.EXPECT().Error().Return(nil)
	mockIter.EXPECT().Value().Return(req1)
	mockConsumer.EXPECT().Consume((5), (0), req1).Return(nil)

	mockIter.EXPECT().Next().Return(true)
	mockIter.EXPECT().Error().Return(nil)
	mockIter.EXPECT().Value().Return(nil)

	mockIter.EXPECT().Close()

	err := provider.Run(0, 10, toRPCConsumer(mockConsumer))
	assert.Error(t, err)
	assert.EqualError(t, err, "iterator returned nil request")
}

func TestRpcRequestProvider_Run_GetLogsSkippedInLoop(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockIter := rpc.NewMockIterator(ctrl)
	mockConsumer := NewMockRPCReqConsumer(ctrl)
	testLog := logger.NewLogger("info", "rpc-provider-test")
	cliCtx := newTestCliContext()
	provider := openRpcRecording(mockIter, &utils.Config{}, testLog, cliCtx, []string{"f1.rpc"}).(*rpcRequestProvider)
	defer provider.Close()

	req1 := &rpc.RequestAndResults{Query: &rpc.Body{MethodBase: "otherMethod"}, Response: &rpc.Response{BlockID: 4}}
	getLogsReq := &rpc.RequestAndResults{Query: &rpc.Body{MethodBase: "getLogs"}, Response: &rpc.Response{BlockID: 5}}
	normalReq := &rpc.RequestAndResults{Query: &rpc.Body{MethodBase: "anotherMethod"}, Response: &rpc.Response{BlockID: 6}}

	mockIter.EXPECT().Next().Return(true)
	mockIter.EXPECT().Error().Return(nil)
	mockIter.EXPECT().Value().Return(req1)
	mockConsumer.EXPECT().Consume(4, 0, req1).Return(nil)

	mockIter.EXPECT().Next().Return(true)
	mockIter.EXPECT().Error().Return(nil)
	mockIter.EXPECT().Value().Return(getLogsReq)

	mockIter.EXPECT().Next().Return(true)
	mockIter.EXPECT().Error().Return(nil)
	mockIter.EXPECT().Value().Return(normalReq)
	mockConsumer.EXPECT().Consume(6, 0, normalReq).Return(nil)

	mockIter.EXPECT().Next().Return(false)
	mockIter.EXPECT().Close()

	err := provider.Run(0, 10, toRPCConsumer(mockConsumer))
	assert.NoError(t, err)
}

func TestRpcRequestProvider_Run_ProcessFirst_IteratorError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockIter := rpc.NewMockIterator(ctrl)

	testLog := logger.NewLogger("info", "rpc-provider-test")
	cliCtx := newTestCliContext()
	provider := openRpcRecording(mockIter, &utils.Config{}, testLog, cliCtx, []string{"f1.rpc"}).(*rpcRequestProvider)
	defer provider.Close()

	iterErr := errors.New("processFirst iter failed")

	mockIter.EXPECT().Next().Return(true)
	mockIter.EXPECT().Error().Return(iterErr).AnyTimes()

	mockIter.EXPECT().Close()

	err := provider.Run(0, 10, toRPCConsumer(NewMockRPCReqConsumer(ctrl)))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "iterator returned error; processFirst iter failed")
}

func TestRpcRequestProvider_Run_ProcessFirst_NilValue(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockIter := rpc.NewMockIterator(ctrl)
	testLog := logger.NewLogger("info", "rpc-provider-test")
	cliCtx := newTestCliContext()
	provider := openRpcRecording(mockIter, &utils.Config{}, testLog, cliCtx, []string{"f1.rpc"}).(*rpcRequestProvider)
	defer provider.Close()

	mockIter.EXPECT().Next().Return(true)
	mockIter.EXPECT().Error().Return(nil)
	mockIter.EXPECT().Value().Return(nil)

	mockIter.EXPECT().Close()

	err := provider.Run(0, 10, toRPCConsumer(NewMockRPCReqConsumer(ctrl)))
	assert.Error(t, err)
	assert.EqualError(t, err, "iterator returned nil request")
}

func TestRpcRequestProvider_Run_ProcessFirst_NoItems_LogsCritical(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockIter := rpc.NewMockIterator(ctrl)
	testLog := logger.NewLogger("info", "rpc-provider-test")
	cliCtx := newTestCliContext()
	provider := openRpcRecording(mockIter, &utils.Config{}, testLog, cliCtx, []string{"f1.rpc"}).(*rpcRequestProvider)
	defer provider.Close()

	mockIter.EXPECT().Next().Return(false).AnyTimes()
	mockIter.EXPECT().Close()

	err := provider.Run(0, 10, toRPCConsumer(NewMockRPCReqConsumer(ctrl)))
	assert.NoError(t, err)
}
