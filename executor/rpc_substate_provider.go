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

package executor

//go:generate mockgen -source rpc_substate_provider.go -destination rpc_substate_provider_mocks.go -package executor

import (
	"fmt"

	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/txcontext"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/urfave/cli/v2"
)

// ----------------------------------------------------------------------------
//                              Implementation
// ----------------------------------------------------------------------------

// OpenRPCSubstateProvider opens a substate database as configured in the given parameters.
func OpenRPCSubstateProvider(cfg *utils.Config, ctxt *cli.Context) (Provider[txcontext.TxContext], error) {
	ipcPath := cfg.OperaDb + "/sonic.ipc"

	log := logger.NewLogger("info", "RPCSubstateProvider")
	client, err := utils.GetRpcOrIpcClient(ctxt.Context, cfg.ChainID, ipcPath, log)
	if err != nil {
		return nil, err
	}
	return &rpcSubstateProvider{
		client:              client,
		ctxt:                ctxt,
		numParallelDecoders: cfg.Workers,
	}, nil
}

// rpcSubstateProvider is an adapter of Aida's RPCRpcsubstateProvider interface defined above to the
// current substate implementation offered by github.com/0xsoniclabs/substate.
type rpcSubstateProvider struct {
	client              *rpc.Client
	ctxt                *cli.Context
	numParallelDecoders int
}

func (s rpcSubstateProvider) Run(from int, to int, consumer Consumer[txcontext.TxContext]) error {
	for i := from; i < to; i++ {
		res, err := utils.RetrieveBlock(s.client, fmt.Sprintf("0x%x", i))
		if err != nil {
			return fmt.Errorf("failed to retrieve block %d; %w", i, err)
		}
		fmt.Printf("Block %d: %s\n", i, res)
	}
	return nil
	//iter := s.db.NewSubstateIterator(from, s.numParallelDecoders)
	//for iter.Next() {
	//	tx := iter.Value()
	//	if tx.Block >= uint64(to) {
	//		return nil
	//	}
	//	if err := consumer(TransactionInfo[txcontext.TxContext]{int(tx.Block), tx.Transaction, substatecontext.NewTxContext(tx)}); err != nil {
	//		return err
	//	}
	//}
	//// this cannot be used in defer because Release() has a WaitGroup.Wait() call
	//// so if called after iter.Error() there is a change the error does not get distributed.
	//iter.Release()
	//return iter.Error()
}

func (s rpcSubstateProvider) Close() {
	s.client.Close()
}
