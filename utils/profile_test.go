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

package utils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/state"
	"github.com/0xsoniclabs/tosca/go/tosca"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestStartCPUProfile(t *testing.T) {
	tempDir := t.TempDir()
	profilePath := filepath.Join(tempDir, "cpu.prof")

	t.Run("WithValidPath", func(t *testing.T) {
		cfg := &Config{CPUProfile: profilePath}
		err := StartCPUProfile(cfg)
		assert.NoError(t, err)

		// Verify file was created
		_, err = os.Stat(profilePath)
		assert.NoError(t, err)

		// Clean up
		StopCPUProfile(cfg)
	})

	t.Run("WithInvalidPath", func(t *testing.T) {
		// Set profile to a path that is inaccessible
		cfg := &Config{CPUProfile: "/nonexistent/directory/cpu.prof"}
		err := StartCPUProfile(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "could not create CPU profile")
	})

	t.Run("WithEmptyPath", func(t *testing.T) {
		cfg := &Config{CPUProfile: ""}
		err := StartCPUProfile(cfg)
		assert.NoError(t, err)
	})
}

func TestStopCPUProfile(t *testing.T) {
	t.Run("WithEmptyPath", func(t *testing.T) {
		cfg := &Config{CPUProfile: ""}
		// This should not panic
		StopCPUProfile(cfg)
	})

	t.Run("AfterStarting", func(t *testing.T) {
		tempDir := t.TempDir()
		profilePath := filepath.Join(tempDir, "cpu_to_stop.prof")

		cfg := &Config{CPUProfile: profilePath}
		err := StartCPUProfile(cfg)
		assert.NoError(t, err)

		// This should not panic
		StopCPUProfile(cfg)
	})
}

func TestStartMemoryProfile(t *testing.T) {
	tempDir := t.TempDir()
	profilePath := filepath.Join(tempDir, "mem.prof")

	t.Run("WithValidPath", func(t *testing.T) {
		cfg := &Config{MemoryProfile: profilePath}
		err := StartMemoryProfile(cfg)
		assert.NoError(t, err)

		// Verify file was created
		_, err = os.Stat(profilePath)
		assert.NoError(t, err)
	})

	t.Run("WithInvalidPath", func(t *testing.T) {
		// Set profile to a path that is inaccessible
		cfg := &Config{MemoryProfile: "/nonexistent/directory/mem.prof"}
		err := StartMemoryProfile(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "could not create memory profile")
	})

	t.Run("WithEmptyPath", func(t *testing.T) {
		cfg := &Config{MemoryProfile: ""}
		err := StartMemoryProfile(cfg)
		assert.NoError(t, err)
	})
}

type temp struct {
	logger logger.Logger
}

func (p *temp) Run(parameters tosca.Parameters) (tosca.Result, error) {
	//TODO implement me
	panic("implement me")
}

func (p *temp) ResetProfile() {
	//TODO implement me
	panic("implement me")
}

func (p *temp) DumpProfile() {
	p.logger.Noticef("DumpProfile")
}

func (p *temp) String() string {
	return "temp"
}

func TestMemoryBreakdown(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := state.NewMockStateDB(ctrl)
	mockLogger := logger.NewMockLogger(ctrl)

	t.Run("WithBreakdownEnabled", func(t *testing.T) {
		cfg := &Config{MemoryBreakdown: true}

		// Mock memory usage with breakdown
		mockUsage := &state.MemoryUsage{
			UsedBytes: 1000,
			Breakdown: &temp{},
		}
		mockDB.EXPECT().GetMemoryUsage().Return(mockUsage)
		mockLogger.EXPECT().Noticef("State DB memory usage: %d byte\n%s", mockUsage.UsedBytes, mockUsage.Breakdown)

		MemoryBreakdown(mockDB, cfg, mockLogger)
	})

	t.Run("WithBreakdownEnabledButUnavailable", func(t *testing.T) {
		cfg := &Config{
			MemoryBreakdown: true,
			DbImpl:          "somedb",
			DbVariant:       "somevariant",
		}

		// Mock memory usage without breakdown
		mockUsage := &state.MemoryUsage{
			UsedBytes: 1000,
			Breakdown: nil,
		}
		mockDB.EXPECT().GetMemoryUsage().Return(mockUsage)
		mockLogger.EXPECT().Noticef("Memory usage summary is unavailable. The selected storage solution: %v variant: %v, may not support memory breakdowns.",
			cfg.DbImpl, cfg.DbVariant)

		MemoryBreakdown(mockDB, cfg, mockLogger)
	})

	t.Run("WithBreakdownDisabled", func(t *testing.T) {
		cfg := &Config{MemoryBreakdown: false}

		// No method calls expected

		MemoryBreakdown(mockDB, cfg, mockLogger)
	})
}

func TestPrintEvmStatistics(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := logger.NewMockLogger(ctrl)

	err := tosca.RegisterInterpreterFactory("test-vm", func(config any) (tosca.Interpreter, error) {
		return &temp{
			logger: mockLogger,
		}, nil
	})
	if err != nil {
		t.Fatalf("Failed to register interpreter factory: %v", err)
	}
	cfg := &Config{
		VmImpl: "test-vm",
	}
	mockLogger.EXPECT().Noticef("DumpProfile").Return()
	PrintEvmStatistics(cfg)
}
