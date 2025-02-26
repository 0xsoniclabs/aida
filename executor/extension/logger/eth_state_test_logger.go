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
	"github.com/0xsoniclabs/aida/executor"
	"github.com/0xsoniclabs/aida/executor/extension"
	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/txcontext"
	"github.com/0xsoniclabs/aida/utils"
)

const defaultReportFrequency = 1000

type ethStateTestLogger struct {
	extension.NilExtension[txcontext.TxContext]
	log             logger.Logger
	overall         int
	reportFrequency int
}

func MakeEthStateTestLogger(cfg *utils.Config, testReportFrequency int) executor.Extension[txcontext.TxContext] {
	if testReportFrequency <= 0 {
		testReportFrequency = defaultReportFrequency
	}
	return makeEthStateTestLogger(logger.NewLogger(cfg.LogLevel, "EthStateTestLogger"), testReportFrequency)
}

func makeEthStateTestLogger(log logger.Logger, frequency int) *ethStateTestLogger {
	return &ethStateTestLogger{
		reportFrequency: frequency,
		log:             log,
		overall:         0,
	}
}

// PreBlock reports test name and fork.
func (l *ethStateTestLogger) PreBlock(s executor.State[txcontext.TxContext], _ *executor.Context) error {
	l.log.Infof("Currently running:\n%s", s.Data)
	l.overall++
	if l.overall%l.reportFrequency == 0 {
		l.log.Noticef("%v tests has been processed so far...", l.overall)
	}
	return nil
}

// PostRun reports total tests run.
func (l *ethStateTestLogger) PostRun(executor.State[txcontext.TxContext], *executor.Context, error) error {
	l.log.Noticef("Total %v tests processed.", l.overall)
	return nil
}
