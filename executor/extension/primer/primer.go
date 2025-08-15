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

package primer

import (
	"github.com/0xsoniclabs/aida/executor"
	"github.com/0xsoniclabs/aida/executor/extension"
	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/prime"
	"github.com/0xsoniclabs/aida/utils"
)

func MakeStateDbPrimer[T any](cfg *utils.Config) executor.Extension[T] {
	if cfg.SkipPriming {
		return extension.NilExtension[T]{}
	}

	return makeStateDbPrimer[T](cfg, logger.NewLogger(cfg.LogLevel, "StateDb-Primer"))
}

func makeStateDbPrimer[T any](cfg *utils.Config, log logger.Logger) *stateDbPrimer[T] {
	return &stateDbPrimer[T]{
		cfg: cfg,
		log: log,
	}
}

type stateDbPrimer[T any] struct {
	extension.NilExtension[T]
	cfg *utils.Config
	log logger.Logger
}

// PreRun primes StateDb to given block.
func (p *stateDbPrimer[T]) PreRun(_ executor.State[T], ctx *executor.Context) (err error) {
	if p.cfg.SkipPriming {
		p.log.Warning("Skipping priming (disabled by user)...")
		return nil
	}

	if p.cfg.PrimeRandom {
		p.log.Infof("Randomized Priming enabled; Seed: %v, threshold: %v", p.cfg.RandomSeed, p.cfg.PrimeThreshold)
	}

	p.log.Infof("Update buffer size: %v bytes", p.cfg.UpdateBufferSize)
	primer := prime.NewPrimer(p.cfg, ctx.State, ctx.AidaDb, p.log)
	return primer.Prime()
}
