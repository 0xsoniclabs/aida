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

import (
	"errors"
	"fmt"
	"os"

	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/rpc"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/urfave/cli/v2"
)

func OpenRpcRecording(cfg *utils.Config, ctx *cli.Context) (Provider[*rpc.RequestAndResults], error) {
	fileInfo, err := os.Stat(cfg.RpcRecordingPath)
	if err != nil {
		return nil, fmt.Errorf("cannot stat the rpc path; %w", err)
	}

	log := logger.NewLogger(cfg.LogLevel, "rpc-provider")
	if !fileInfo.IsDir() {
		iter, err := rpc.NewFileReader(ctx.Context, cfg.RpcRecordingPath)
		if err != nil {
			return nil, fmt.Errorf("cannot open rpc recording file; %v", err)
		}
		return openRpcRecording(iter, cfg, log, ctx, []string{cfg.RpcRecordingPath}), nil
	}

	files, err := utils.GetFilesWithinDirectories("", []string{cfg.RpcRecordingPath})
	if err != nil {
		return nil, fmt.Errorf("cannot get files from dir %v; %w", cfg.RpcRecordingPath, err)
	}

	iter, err := rpc.NewFileReader(ctx.Context, files[0])
	if err != nil {
		return nil, fmt.Errorf("cannot open rpc recording file; %v", err)
	}
	return openRpcRecording(iter, cfg, log, ctx, files), nil

}

func openRpcRecording(iter rpc.Iterator, cfg *utils.Config, log logger.Logger, ctxt *cli.Context, files []string) Provider[*rpc.RequestAndResults] {
	return &rpcRequestProvider{
		ctxt:     ctxt,
		fileName: cfg.RpcRecordingPath,
		iter:     iter,
		files:    files,
		log:      log,
	}
}

type rpcRequestProvider struct {
	ctxt     *cli.Context
	fileName string
	iter     rpc.Iterator
	log      logger.Logger
	files    []string
	nextFile int
}

func (r *rpcRequestProvider) Run(from int, to int, consumer Consumer[*rpc.RequestAndResults]) (err error) {
	r.nextFile++

	defer func() {
		if err != nil {
			r.log.Infof("Last iterated file: %v", r.files[r.nextFile-1])
		}
	}()

	if err = r.processFirst(from, consumer); err != nil {
		return err
	}

	for r.iter.Next() {
		if r.iter.Error() != nil {
			return fmt.Errorf("iterator returned error; %v", r.iter.Error())
		}

		req := r.iter.Value()

		if req == nil {
			return errors.New("iterator returned nil request")
		}

		// get logs is not yet implemented, skip these for now
		if req.Query.MethodBase == "getLogs" {
			continue
		}

		req.DecodeInfo()
		// are we skipping requests?
		if req.RecordedBlock < from {
			continue
		}

		if req.RecordedBlock >= to {
			return nil
		}

		if err = consumer(TransactionInfo[*rpc.RequestAndResults]{req.RecordedBlock, 0, req}); err != nil {
			return err
		}
	}

	if r.nextFile < len(r.files) {
		var err error
		r.iter, err = rpc.NewFileReader(r.ctxt.Context, r.files[r.nextFile])
		if err != nil {
			return fmt.Errorf("cannot open rpc recording file %v; %w", r.files[r.nextFile], err)
		}
		return r.Run(from, to, consumer)
	}

	return nil
}

func (r *rpcRequestProvider) Close() {
	r.iter.Close()
}

// processFirst takes first request and logs information about run.
func (r *rpcRequestProvider) processFirst(from int, consumer Consumer[*rpc.RequestAndResults]) error {
	if r.iter.Next() {
		if r.iter.Error() != nil {
			return fmt.Errorf("iterator returned error; %v", r.iter.Error())
		}

		req := r.iter.Value()
		if req == nil {
			return errors.New("iterator returned nil request")
		}

		req.DecodeInfo()
		r.log.Noticef("Iterating file %v/%v path: %v", r.nextFile, len(r.files), r.files[r.nextFile-1])
		r.log.Noticef("First block of recording: %v", req.RecordedBlock)

		// are we skipping requests?
		if req.RecordedBlock < from {
			r.log.Noticef("Skipping %v blocks. This might take a while, skip rate is ~50k Req/s "+
				"and there is up to 2500 Requests in a block.", from-req.RecordedBlock)
			return nil
		}

		// get logs is not yet implemented, skip these for now
		if req.Query.MethodBase == "getLogs" {
			return nil
		}

		if err := consumer(TransactionInfo[*rpc.RequestAndResults]{req.RecordedBlock, 0, req}); err != nil {
			return err
		}
	} else {
		r.log.Critical("Iterator returned no requests.")
	}

	return nil
}
