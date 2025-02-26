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

package logger

import (
	"bufio"
	"fmt"
	"os"
	"sync"

	"github.com/0xsoniclabs/aida/executor"
	"github.com/0xsoniclabs/aida/executor/extension"
	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/state/proxy"
	"github.com/0xsoniclabs/aida/utils"
)

const inputSize = 100

type dbLogger[T any] struct {
	extension.NilExtension[T]
	cfg    *utils.Config
	log    logger.Logger
	file   *os.File
	writer *bufio.Writer
	input  chan string
	wg     *sync.WaitGroup
}

// MakeDbLogger creates an extensions which logs any Db transaction into a file and log level DEBUG
func MakeDbLogger[T any](cfg *utils.Config) executor.Extension[T] {
	if cfg.DbLogging == "" {
		return extension.NilExtension[T]{}
	}

	return makeDbLogger[T](cfg, logger.NewLogger(cfg.LogLevel, "Db-Logger"))
}

func makeDbLogger[T any](cfg *utils.Config, log logger.Logger) *dbLogger[T] {
	return &dbLogger[T]{
		cfg:   cfg,
		log:   log,
		input: make(chan string, inputSize),
		wg:    new(sync.WaitGroup),
	}
}

// PreRun creates a logging file
func (l *dbLogger[T]) PreRun(_ executor.State[T], ctx *executor.Context) error {
	var err error
	l.file, err = os.Create(l.cfg.DbLogging)
	if err != nil {
		return fmt.Errorf("cannot create db-logging file; %v", err)
	}
	// create buffered logging
	l.writer = bufio.NewWriter(l.file)

	l.wg.Add(1)
	go l.doLogging()

	// in some cases, StateDb does not have to be initialized yet
	if ctx.State != nil {
		ctx.State = proxy.NewLoggerProxy(ctx.State, l.log, l.input, l.wg)
	}

	return nil
}

// PreTransaction checks whether ctx.State has not been overwritten by temporary prepper,
// if so it creates new NewLoggerProxy. This is mainly used by the aida-vm tool.
func (l *dbLogger[T]) PreTransaction(_ executor.State[T], ctx *executor.Context) error {
	// if ctx.State has not been change, no need to slow down the app by creating new Proxy
	if _, ok := ctx.State.(*proxy.LoggingStateDb); ok {
		return nil
	}

	ctx.State = proxy.NewLoggerProxy(ctx.State, l.log, l.input, l.wg)
	return nil
}

func (l *dbLogger[T]) doLogging() {
	defer func() {
		err := l.writer.Flush()
		if err != nil {
			l.log.Errorf("cannot flush db-logging writer; %v", err)
		}

		err = l.file.Close()
		if err != nil {
			l.log.Errorf("cannot close db-logging file; %v", err)
		}
		l.wg.Done()
	}()
	for {
		in, ok := <-l.input
		if !ok {
			return
		}
		_, err := l.writer.WriteString(in + "\n")
		if err != nil {
			l.log.Errorf("cannot write into db-log-file; %v", err)
		}
	}
}
