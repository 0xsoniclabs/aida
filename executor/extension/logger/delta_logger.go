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

package logger

import (
	"bufio"
	"fmt"
	"os"

	"github.com/0xsoniclabs/aida/executor"
	"github.com/0xsoniclabs/aida/executor/extension"
	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/state/proxy"
	"github.com/0xsoniclabs/aida/utils"
)

type deltaLogger[T any] struct {
	extension.NilExtension[T]
	cfg  *utils.Config
	log  logger.Logger
	sink *proxy.DeltaLogSink
}

// MakeDeltaLogger creates an extension that produces delta-debugger compatible traces.
func MakeDeltaLogger[T any](cfg *utils.Config) executor.Extension[T] {
	if cfg.DeltaLogging == "" {
		return extension.NilExtension[T]{}
	}

	return makeDeltaLogger[T](cfg, logger.NewLogger(cfg.LogLevel, "Delta-Logger"))
}

func makeDeltaLogger[T any](cfg *utils.Config, log logger.Logger) *deltaLogger[T] {
	return &deltaLogger[T]{cfg: cfg, log: log}
}

// PreRun prepares the sink and wraps an already initialized StateDB.
func (l *deltaLogger[T]) PreRun(_ executor.State[T], ctx *executor.Context) error {
	file, err := os.Create(l.cfg.DeltaLogging)
	if err != nil {
		return fmt.Errorf("cannot create delta-log file; %w", err)
	}

	l.sink = proxy.NewDeltaLogSink(l.log, bufio.NewWriter(file), file)

	if ctx.State != nil {
		ctx.State = proxy.NewDeltaLoggerProxy(ctx.State, l.sink)
	}
	return nil
}

// PreTransaction wraps any freshly created StateDB so per-transaction state is logged.
func (l *deltaLogger[T]) PreTransaction(_ executor.State[T], ctx *executor.Context) error {
	if l.sink == nil || ctx.State == nil {
		return nil
	}

	if _, ok := ctx.State.(*proxy.DeltaLoggingStateDB); ok {
		return nil
	}

	ctx.State = proxy.NewDeltaLoggerProxy(ctx.State, l.sink)
	return nil
}

// PostRun closes the sink to flush and fsync the trace.
func (l *deltaLogger[T]) PostRun(_ executor.State[T], _ *executor.Context, _ error) error {
	if l.sink == nil {
		return nil
	}
	return l.sink.Close()
}
