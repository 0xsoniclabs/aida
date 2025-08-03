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

package profiler

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/0xsoniclabs/aida/executor"
	"github.com/0xsoniclabs/aida/executor/extension"
	"github.com/0xsoniclabs/aida/profile/blockprofile"
	"github.com/0xsoniclabs/aida/txcontext"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestBlockProfilerExtension_NoProfileIsCollectedIfDisabled(t *testing.T) {
	config := &utils.Config{}
	ext := MakeBlockRuntimeAndGasCollector(config)

	if _, ok := ext.(extension.NilExtension[txcontext.TxContext]); !ok {
		t.Errorf("profiler is enabled although not set in configuration")
	}
}

func TestBlockProfilerExtension_ProfileDbIsCreated(t *testing.T) {
	path := t.TempDir() + "/profile.db"
	config := &utils.Config{}
	config.ProfileBlocks = true
	config.ProfileDB = path

	ext := MakeBlockRuntimeAndGasCollector(config)

	if err := ext.PreRun(executor.State[txcontext.TxContext]{}, nil); err != nil {
		t.Fatalf("unexpected error during pre-run; %v", err)
	}

	if _, err := os.Stat(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			t.Fatal("db was not created")
		}
		t.Fatalf("unexpected error; %v", err)
	}
}

func TestBlockRuntimeAndGasCollector_PreTransaction(t *testing.T) {
	b := &BlockRuntimeAndGasCollector{}
	s := executor.State[txcontext.TxContext]{}
	err := b.PreTransaction(s, nil)
	assert.Nil(t, err)
	assert.InDelta(t, time.Now().Second(), b.txTimer.Second(), 5)
}

func TestBlockRuntimeAndGasCollector_PostTransaction(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockProfileDB := blockprofile.NewMockProfileDB(ctrl)
	mockContext := blockprofile.NewMockContext(ctrl)
	b := &BlockRuntimeAndGasCollector{
		profileDb: mockProfileDB,
		ctx:       mockContext,
	}
	s := executor.State[txcontext.TxContext]{}
	mockContext.EXPECT().RecordTransaction(gomock.Any(), gomock.Any()).Return(nil)
	err := b.PostTransaction(s, nil)
	assert.Nil(t, err)
	assert.InDelta(t, time.Now().Second(), b.txTimer.Second(), 60)
}

func TestBlockRuntimeAndGasCollector_PreBlock(t *testing.T) {
	b := &BlockRuntimeAndGasCollector{}
	s := executor.State[txcontext.TxContext]{}
	err := b.PreBlock(s, nil)
	assert.Nil(t, err)
	assert.InDelta(t, time.Now().Second(), b.blockTimer.Second(), 5)
}

func TestBlockRuntimeAndGasCollector_PostBlock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("case success", func(t *testing.T) {
		mockProfileDB := blockprofile.NewMockProfileDB(ctrl)
		mockContext := blockprofile.NewMockContext(ctrl)
		b := &BlockRuntimeAndGasCollector{
			profileDb: mockProfileDB,
			ctx:       mockContext,
		}
		mockContext.EXPECT().GetProfileData(gomock.Any(), gomock.Any()).Return(&blockprofile.ProfileData{}, nil)
		mockProfileDB.EXPECT().Add(gomock.Any()).Return(nil)
		s := executor.State[txcontext.TxContext]{}
		err := b.PostBlock(s, nil)
		assert.Nil(t, err)
		assert.InDelta(t, time.Now().Second(), b.blockTimer.Second(), 60)
	})
	t.Run("case error", func(t *testing.T) {
		mockProfileDB := blockprofile.NewMockProfileDB(ctrl)
		mockContext := blockprofile.NewMockContext(ctrl)
		b := &BlockRuntimeAndGasCollector{
			profileDb: mockProfileDB,
			ctx:       mockContext,
		}
		mockError := errors.New("mock error")
		mockContext.EXPECT().GetProfileData(gomock.Any(), gomock.Any()).Return(nil, mockError)
		s := executor.State[txcontext.TxContext]{}
		err := b.PostBlock(s, nil)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), mockError.Error())
		assert.InDelta(t, time.Now().Second(), b.blockTimer.Second(), 60)
	})
	t.Run("case error 2", func(t *testing.T) {
		mockProfileDB := blockprofile.NewMockProfileDB(ctrl)
		mockContext := blockprofile.NewMockContext(ctrl)
		b := &BlockRuntimeAndGasCollector{
			profileDb: mockProfileDB,
			ctx:       mockContext,
		}
		mockError := errors.New("mock error")
		mockContext.EXPECT().GetProfileData(gomock.Any(), gomock.Any()).Return(&blockprofile.ProfileData{}, nil)
		mockProfileDB.EXPECT().Add(gomock.Any()).Return(mockError)
		s := executor.State[txcontext.TxContext]{}
		err := b.PostBlock(s, nil)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), mockError.Error())
		assert.InDelta(t, time.Now().Second(), b.blockTimer.Second(), 60)
	})
}

func TestBlockRuntimeAndGasCollector_PostRun(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockProfileDB := blockprofile.NewMockProfileDB(ctrl)
	mockContext := blockprofile.NewMockContext(ctrl)
	b := &BlockRuntimeAndGasCollector{
		profileDb: mockProfileDB,
		ctx:       mockContext,
	}
	mockProfileDB.EXPECT().Close().Return(nil)
	s := executor.State[txcontext.TxContext]{}
	err := b.PostRun(s, nil, nil)
	assert.Nil(t, err)
}
