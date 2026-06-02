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

package executor

//go:generate mockgen -source substate_provider.go -destination substate_provider_mocks.go -package executor

import (
	"errors"

	"github.com/0xsoniclabs/aida/txcontext"
	substatecontext "github.com/0xsoniclabs/aida/txcontext/substate"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/substate/db"
	"github.com/urfave/cli/v2"
)

// ----------------------------------------------------------------------------
//                              Implementation
// ----------------------------------------------------------------------------

// OpenSubstateProvider opens a substate database as configured in the given parameters.
func OpenSubstateProvider(cfg *utils.Config, ctxt *cli.Context, aidaDb db.BaseDB) (Provider[txcontext.TxContext], error) {
	substateDb, err := db.MakeDefaultSubstateDBFromBaseDB(aidaDb)
	if err != nil {
		return nil, err
	}
	return &substateProvider{
		db:                  substateDb,
		ctxt:                ctxt,
		numParallelDecoders: cfg.Workers,
	}, nil
}

// substateProvider is an adapter of Aida's SubstateProvider interface defined above to the
// current substate implementation offered by github.com/0xsoniclabs/substate.
type substateProvider struct {
	db                  db.SubstateDB
	ctxt                *cli.Context
	numParallelDecoders int
}

func (s substateProvider) Run(from int, to int, consumer Consumer[txcontext.TxContext]) error {
	iter := s.db.NewSubstateIterator(from, s.numParallelDecoders)
	release := func(err error) error {
		iter.Release()
		return errors.Join(err, iter.Error())
	}

	for iter.Next() {
		tx := iter.Value()
		if tx.Block >= uint64(to) {
			return release(nil)
		}
		if err := consumer(TransactionInfo[txcontext.TxContext]{int(tx.Block), tx.Transaction, substatecontext.NewTxContext(tx)}); err != nil {
			return release(err)
		}
	}
	return release(nil)
}

func (s substateProvider) Close() {
	// ignored, database is opened it top-most level
}
