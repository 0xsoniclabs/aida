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

package register

import (
	"testing"

	"github.com/0xsoniclabs/aida/config"
	"github.com/0xsoniclabs/aida/executor"
	"github.com/0xsoniclabs/aida/logger"
	rr "github.com/0xsoniclabs/aida/register"
	"github.com/0xsoniclabs/aida/rpc"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestRegisterRequestProgress_PreRun(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	dir := t.TempDir()
	r := &registerRequestProgress{
		log: logger.NewLogger("info", "test"),
		id:  rr.MakeRunIdentity(int64(0), &config.Config{}),
		cfg: &config.Config{
			RegisterRun: dir,
		},
		ps: utils.NewPrinters(),
	}

	err := r.PreRun(executor.State[*rpc.RequestAndResults]{}, &executor.Context{})
	assert.Nil(t, err)
}

func TestRegisterRequestProgress_PostTransaction(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	dir := t.TempDir()
	r := &registerRequestProgress{
		log: logger.NewLogger("info", "test"),
		id:  rr.MakeRunIdentity(int64(0), &config.Config{}),
		cfg: &config.Config{
			RegisterRun: dir,
		},
		ps:              utils.NewPrinters(),
		reportFrequency: 1,
	}

	err := r.PostTransaction(executor.State[*rpc.RequestAndResults]{}, &executor.Context{})
	assert.Nil(t, err)
}

func TestRegisterRequestProgress_PostRun(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	dir := t.TempDir()
	cfg := &config.Config{
		RegisterRun: dir,
	}
	mockPrinter := utils.NewMockPrinter(ctrl)
	ps := utils.NewCustomPrinters([]utils.Printer{mockPrinter})
	id := rr.MakeRunIdentity(int64(0), cfg)
	meta, _ := rr.MakeRunMetadata("", id, rr.FetchUnixInfo)
	meta.Ps = ps
	r := &registerRequestProgress{
		log:             logger.NewLogger("info", "test"),
		id:              id,
		cfg:             cfg,
		ps:              ps,
		reportFrequency: 1,
		meta:            meta,
	}
	mockPrinter.EXPECT().Print().Times(2)
	mockPrinter.EXPECT().Close().Times(2)

	err := r.PostRun(executor.State[*rpc.RequestAndResults]{}, &executor.Context{}, nil)
	assert.Nil(t, err)
}

func TestMakeRegisterRequestProgress(t *testing.T) {
	cfg := &config.Config{
		RegisterRun: t.TempDir(),
	}
	r := MakeRegisterRequestProgress(cfg, 1, 1)
	_, ok := r.(*registerRequestProgress)
	assert.True(t, ok)
}
