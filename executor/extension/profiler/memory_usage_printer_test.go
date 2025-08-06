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
	"github.com/0xsoniclabs/aida/config"
	"testing"

	"github.com/0xsoniclabs/aida/executor"
	"github.com/0xsoniclabs/aida/executor/extension"
	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/state"
	"github.com/gogo/protobuf/plugin/stringer"
	"go.uber.org/mock/gomock"
)

func TestMemoryUsagePrinter_MemoryBreakdownIsNotPrintedWhenBreakdownIsNil(t *testing.T) {
	cfg := &config.Config{}
	cfg.MemoryBreakdown = true

	ctrl := gomock.NewController(t)

	log := logger.NewMockLogger(ctrl)
	db := state.NewMockStateDB(ctrl)
	ext := makeMemoryUsagePrinter[any](cfg, log)

	usage := &state.MemoryUsage{
		Breakdown: nil,
	}

	gomock.InOrder(
		// Prerun
		db.EXPECT().GetMemoryUsage().Return(usage),
		log.EXPECT().Noticef(gomock.Any(), gomock.Any(), gomock.Any()),

		// Postrun
		db.EXPECT().GetMemoryUsage().Return(usage),
		log.EXPECT().Noticef(gomock.Any(), gomock.Any(), gomock.Any()),
	)

	ext.PreRun(executor.State[any]{}, &executor.Context{State: db})
	ext.PostRun(executor.State[any]{}, &executor.Context{State: db}, nil)

}

func TestMemoryUsagePrinter_MemoryBreakdownIsNotPrintedWhenDatabaseIsNil(t *testing.T) {
	cfg := &config.Config{}
	cfg.MemoryBreakdown = true

	ctrl := gomock.NewController(t)

	log := logger.NewMockLogger(ctrl)
	ext := makeMemoryUsagePrinter[any](cfg, log)

	gomock.InOrder(
		log.EXPECT().Notice(gomock.Any()).Times(0),
	)

	ext.PreRun(executor.State[any]{}, &executor.Context{State: nil})
	ext.PostRun(executor.State[any]{}, &executor.Context{State: nil}, nil)

}

func TestMemoryUsagePrinter_MemoryBreakdownIsPrintedWhenEnabled(t *testing.T) {
	cfg := &config.Config{}
	cfg.MemoryBreakdown = true

	ctrl := gomock.NewController(t)

	log := logger.NewMockLogger(ctrl)
	db := state.NewMockStateDB(ctrl)
	ext := makeMemoryUsagePrinter[any](cfg, log)

	usage := &state.MemoryUsage{
		UsedBytes: 1,
		Breakdown: stringer.NewStringer(),
	}

	gomock.InOrder(
		// Prerun
		db.EXPECT().GetMemoryUsage().Return(usage),
		log.EXPECT().Noticef(gomock.Any(), uint64(1), gomock.Any()),

		// Postrun
		db.EXPECT().GetMemoryUsage().Return(usage),
		log.EXPECT().Noticef(gomock.Any(), uint64(1), gomock.Any()),
	)

	ext.PreRun(executor.State[any]{}, &executor.Context{State: db})
	ext.PostRun(executor.State[any]{}, &executor.Context{State: db}, nil)

}

func TestMemoryUsagePrinter_NoPrinterIsCreatedIfNotEnabled(t *testing.T) {
	cfg := &config.Config{}
	ext := MakeMemoryUsagePrinter[any](cfg)

	if _, ok := ext.(extension.NilExtension[any]); !ok {
		t.Errorf("profiler is enabled although not set in configuration")
	}
}
