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

package delta

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/0xsoniclabs/aida/utils"
)

// StateTesterConfig describes how the delta debugger should replay traces.
type StateTesterConfig struct {
	DbImpl       string
	Variant      string
	TmpDir       string
	CarmenSchema int
	LogLevel     string
	ChainID      int
}

// NewStateTester prepares a testFunc that replays operations against a StateDB backend.
func NewStateTester(cfg StateTesterConfig) (testFunc, error) {
	dbImpl := strings.TrimSpace(cfg.DbImpl)
	if dbImpl == "" {
		dbImpl = utils.StateDbImplementationFlag.Value
	}

	tmpDir := cfg.TmpDir
	if tmpDir == "" {
		tmpDir = os.TempDir()
	}

	logLevel := cfg.LogLevel
	if logLevel == "" {
		logLevel = "INFO"
	}

	schema := cfg.CarmenSchema
	if schema == 0 {
		schema = utils.CarmenSchemaFlag.Value
	}

	chainID := cfg.ChainID
	if chainID == 0 {
		chainID = 250
	}

	base := utils.Config{
		DbImpl:       dbImpl,
		DbVariant:    cfg.Variant,
		CarmenSchema: schema,
		DbTmp:        tmpDir,
		LogLevel:     logLevel,
		ChainID:      utils.ChainID(chainID),
	}

	return func(ctx context.Context, ops []TraceOp) (outcome, error) {
		localCfg := base
		db, dbPath, err := utils.PrepareStateDB(&localCfg)
		if err != nil {
			return outcomeUnresolved, fmt.Errorf("delta: prepare state-db: %w", err)
		}

		replayer := newStateReplayer(db)
		var (
			panicValue any
			replayErr  error
			cleanupErr error
		)

		func() {
			defer func() {
				if closeErr := db.Close(); closeErr != nil {
					cleanupErr = errors.Join(cleanupErr, fmt.Errorf("delta: close state-db: %w", closeErr))
				}
				if rmErr := os.RemoveAll(dbPath); rmErr != nil {
					cleanupErr = errors.Join(cleanupErr, fmt.Errorf("delta: remove temp state-db: %w", rmErr))
				}
				if r := recover(); r != nil {
					panicValue = r
				}
			}()
			replayErr = replayer.Execute(ctx, ops)
		}()

		logFailure := func(err error) (outcome, error) {
			if err != nil {
				fmt.Fprintf(os.Stderr, "aida-delta-debugger: backend run failed: %v\n", err)
			}
			return outcomeFail, nil
		}

		if panicValue != nil {
			err := fmt.Errorf("panic: %v", panicValue)
			if replayErr != nil {
				err = errors.Join(replayErr, err)
			}
			if cleanupErr != nil {
				err = errors.Join(err, cleanupErr)
			}
			return logFailure(err)
		}

		if replayErr != nil {
			if errors.Is(replayErr, context.Canceled) || errors.Is(replayErr, context.DeadlineExceeded) {
				return outcomeUnresolved, errors.Join(replayErr, cleanupErr)
			}
			return logFailure(errors.Join(replayErr, cleanupErr))
		}

		if cleanupErr != nil {
			return logFailure(cleanupErr)
		}

		return outcomePass, nil
	}, nil
}
