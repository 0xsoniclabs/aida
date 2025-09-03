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

package profiler

import (
	"github.com/0xsoniclabs/aida/config"
	"github.com/0xsoniclabs/aida/executor"
	"github.com/0xsoniclabs/aida/executor/extension"
	"github.com/0xsoniclabs/aida/profile"
)

// MakeMemoryProfiler creates an executor.Extension that records memory profiling data if enabled in the configuration.
func MakeMemoryProfiler[T any](cfg *config.Config) executor.Extension[T] {
	if cfg.MemoryProfile == "" {
		return extension.NilExtension[T]{}
	}
	return &memoryProfiler[T]{cfg: cfg}
}

type memoryProfiler[T any] struct {
	extension.NilExtension[T]
	cfg *config.Config
}

func (p *memoryProfiler[T]) PostRun(executor.State[T], *executor.Context, error) error {
	return profile.StartMemoryProfile(p.cfg)
}
