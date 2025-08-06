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

package config

import (
	"github.com/0xsoniclabs/aida/cmd/util-db/flags"
	"github.com/0xsoniclabs/aida/config/chainid"
	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/substate/db"
	"github.com/urfave/cli/v2"
)

// createConfigFromFlags returns Config instance with user specified values or the default ones
func createConfigFromFlags(ctx *cli.Context) *Config {
	cfg := &Config{
		AppName:     ctx.App.HelpName,
		CommandName: ctx.Command.Name,

		AidaDb:                   getFlagValue(ctx, utils.AidaDbFlag).(string),
		ArchiveMaxQueryAge:       getFlagValue(ctx, utils.ArchiveMaxQueryAgeFlag).(int),
		ArchiveMode:              getFlagValue(ctx, utils.ArchiveModeFlag).(bool),
		ArchiveQueryRate:         getFlagValue(ctx, utils.ArchiveQueryRateFlag).(int),
		ArchiveVariant:           getFlagValue(ctx, utils.ArchiveVariantFlag).(string),
		BalanceRange:             getFlagValue(ctx, utils.BalanceRangeFlag).(int64),
		BasicBlockProfiling:      getFlagValue(ctx, utils.BasicBlockProfilingFlag).(bool),
		BlockLength:              getFlagValue(ctx, utils.BlockLengthFlag).(uint64),
		CPUProfile:               getFlagValue(ctx, utils.CpuProfileFlag).(string),
		CPUProfilePerInterval:    getFlagValue(ctx, utils.CpuProfilePerIntervalFlag).(bool),
		Cache:                    getFlagValue(ctx, utils.CacheFlag).(int),
		CarmenCheckpointInterval: getFlagValue(ctx, utils.CarmenCheckpointInterval).(int),
		CarmenCheckpointPeriod:   getFlagValue(ctx, utils.CarmenCheckpointPeriod).(int),
		CarmenSchema:             getFlagValue(ctx, utils.CarmenSchemaFlag).(int),
		ChainID:                  chainid.ChainID(getFlagValue(ctx, utils.ChainIDFlag).(int)),
		ChannelBufferSize:        getFlagValue(ctx, utils.ChannelBufferSizeFlag).(int),
		CompactDb:                getFlagValue(ctx, utils.CompactDbFlag).(bool),
		ContinueOnFailure:        getFlagValue(ctx, utils.ContinueOnFailureFlag).(bool),
		ContractNumber:           getFlagValue(ctx, utils.ContractNumberFlag).(int64),
		CustomDbName:             getFlagValue(ctx, utils.CustomDbNameFlag).(string),
		DbComponent:              getFlagValue(ctx, utils.DbComponentFlag).(string),
		DbImpl:                   getFlagValue(ctx, utils.StateDbImplementationFlag).(string),
		DbLogging:                getFlagValue(ctx, utils.StateDbLoggingFlag).(string),
		DbTmp:                    getFlagValue(ctx, utils.DbTmpFlag).(string),
		DbVariant:                getFlagValue(ctx, utils.StateDbVariantFlag).(string),
		Debug:                    getFlagValue(ctx, utils.TraceDebugFlag).(bool),
		DebugFrom:                getFlagValue(ctx, utils.DebugFromFlag).(uint64),
		DeleteSourceDbs:          getFlagValue(ctx, utils.DeleteSourceDbsFlag).(bool),
		DeletionDb:               getFlagValue(ctx, utils.DeletionDbFlag).(string),
		DiagnosticServer:         getFlagValue(ctx, utils.DiagnosticServerFlag).(int64),
		ErrorLogging:             getFlagValue(ctx, utils.ErrorLoggingFlag).(string),
		EvmImpl:                  getFlagValue(ctx, utils.EvmImplementation).(string),
		Fork:                     getFlagValue(ctx, utils.ForkFlag).(string),
		Genesis:                  getFlagValue(ctx, utils.GenesisFlag).(string),
		EthTestType:              chainid.EthTestType(getFlagValue(ctx, utils.EthTestTypeFlag).(int)),
		IncludeStorage:           getFlagValue(ctx, utils.IncludeStorageFlag).(bool),
		KeepDb:                   getFlagValue(ctx, utils.KeepDbFlag).(bool),
		KeysNumber:               getFlagValue(ctx, utils.KeysNumberFlag).(int64),
		LogLevel:                 getFlagValue(ctx, logger.LogLevelFlag).(string),
		MaxNumErrors:             getFlagValue(ctx, utils.MaxNumErrorsFlag).(int),
		MaxNumTransactions:       getFlagValue(ctx, utils.MaxNumTransactionsFlag).(int),
		MemoryBreakdown:          getFlagValue(ctx, utils.MemoryBreakdownFlag).(bool),
		MemoryProfile:            getFlagValue(ctx, utils.MemoryProfileFlag).(string),
		MicroProfiling:           getFlagValue(ctx, utils.MicroProfilingFlag).(bool),
		NoHeartbeatLogging:       getFlagValue(ctx, utils.NoHeartbeatLoggingFlag).(bool),
		NonceRange:               getFlagValue(ctx, utils.NonceRangeFlag).(int),
		OnlySuccessful:           getFlagValue(ctx, utils.OnlySuccessfulFlag).(bool),
		OperaBinary:              getFlagValue(ctx, utils.OperaBinaryFlag).(string),
		ClientDb:                 getFlagValue(ctx, utils.ClientDbFlag).(string),
		Output:                   getFlagValue(ctx, utils.OutputFlag).(string),
		OverwriteRunId:           getFlagValue(ctx, utils.OverwriteRunIdFlag).(string),
		PrimeRandom:              getFlagValue(ctx, utils.RandomizePrimingFlag).(bool),
		PrimeThreshold:           getFlagValue(ctx, utils.PrimeThresholdFlag).(int),
		Profile:                  getFlagValue(ctx, utils.ProfileFlag).(bool),
		ProfileBlocks:            getFlagValue(ctx, utils.ProfileBlocksFlag).(bool),
		ProfileDB:                getFlagValue(ctx, utils.ProfileDBFlag).(string),
		ProfileDepth:             getFlagValue(ctx, utils.ProfileDepthFlag).(int),
		ProfileEVMCall:           getFlagValue(ctx, utils.ProfileEVMCallFlag).(bool),
		ProfileFile:              getFlagValue(ctx, utils.ProfileFileFlag).(string),
		ProfileInterval:          getFlagValue(ctx, utils.ProfileIntervalFlag).(uint64),
		ProfileSqlite3:           getFlagValue(ctx, utils.ProfileSqlite3Flag).(string),
		ProfilingDbName:          getFlagValue(ctx, utils.ProfilingDbNameFlag).(string),
		RandomSeed:               getFlagValue(ctx, utils.RandomSeedFlag).(int64),
		RegisterRun:              getFlagValue(ctx, utils.RegisterRunFlag).(string),
		RpcRecordingPath:         getFlagValue(ctx, utils.RpcRecordingFileFlag).(string),
		ShadowDb:                 getFlagValue(ctx, utils.ShadowDb).(bool),
		ShadowImpl:               getFlagValue(ctx, utils.ShadowDbImplementationFlag).(string),
		ShadowVariant:            getFlagValue(ctx, utils.ShadowDbVariantFlag).(string),
		SkipMetadata:             getFlagValue(ctx, flags.SkipMetadata).(bool),
		SkipPriming:              getFlagValue(ctx, utils.SkipPrimingFlag).(bool),
		SkipStateHashScrapping:   getFlagValue(ctx, utils.SkipStateHashScrappingFlag).(bool),
		SnapshotDepth:            getFlagValue(ctx, utils.SnapshotDepthFlag).(int),
		StateDbSrc:               getFlagValue(ctx, utils.StateDbSrcFlag).(string),
		StateDbSrcDirectAccess:   getFlagValue(ctx, utils.StateDbSrcOverwriteFlag).(bool),
		StateDbSrcReadOnly:       false,
		// TODO re-enable equality check once supported in Carmen
		StateValidationMode:    SubsetCheck,
		SubstateDb:             getFlagValue(ctx, utils.AidaDbFlag).(string),
		SubstateEncoding:       db.SubstateEncodingSchema(getFlagValue(ctx, utils.SubstateEncodingFlag).(string)),
		SyncPeriodLength:       getFlagValue(ctx, utils.SyncPeriodLengthFlag).(uint64),
		TargetDb:               getFlagValue(ctx, utils.TargetDbFlag).(string),
		TargetEpoch:            getFlagValue(ctx, utils.TargetEpochFlag).(uint64),
		Trace:                  getFlagValue(ctx, utils.TraceFlag).(bool),
		TraceDirectory:         getFlagValue(ctx, utils.TraceDirectoryFlag).(string),
		TraceFile:              getFlagValue(ctx, utils.TraceFileFlag).(string),
		TrackProgress:          getFlagValue(ctx, utils.TrackProgressFlag).(bool),
		TrackerGranularity:     getFlagValue(ctx, utils.TrackerGranularityFlag).(int),
		TransactionLength:      getFlagValue(ctx, utils.TransactionLengthFlag).(uint64),
		UpdateBufferSize:       getFlagValue(ctx, utils.UpdateBufferSizeFlag).(uint64),
		UpdateDb:               getFlagValue(ctx, utils.UpdateDbFlag).(string),
		OverwritePreWorldState: getFlagValue(ctx, utils.OverwritePreWorldStateFlag).(bool),
		UpdateType:             getFlagValue(ctx, utils.UpdateTypeFlag).(string),
		Validate:               getFlagValue(ctx, utils.ValidateFlag).(bool),
		ValidateStateHashes:    getFlagValue(ctx, utils.ValidateStateHashesFlag).(bool),
		ValidateTxState:        getFlagValue(ctx, utils.ValidateTxStateFlag).(bool),
		ValuesNumber:           getFlagValue(ctx, utils.ValuesNumberFlag).(int64),
		VmImpl:                 getFlagValue(ctx, utils.VmImplementation).(string),
		Workers:                getFlagValue(ctx, utils.WorkersFlag).(int),
		TxGeneratorType:        getFlagValue(ctx, utils.TxGeneratorTypeFlag).([]string),
	}

	return cfg
}

// getFlagValue returns value specified by user if flag is present in cli context, otherwise return default flag value
func getFlagValue(ctx *cli.Context, flag interface{}) interface{} {
	cmdFlags := ctx.Command.Flags
	for _, cmdFlag := range cmdFlags {
		switch f := flag.(type) {
		case cli.IntFlag:
			if cmdFlag.Names()[0] == f.Name {
				return ctx.Int(f.Name)
			}

		case cli.Uint64Flag:
			if cmdFlag.Names()[0] == utils.UpdateBufferSizeFlag.Name {
				return ctx.Uint64(f.Name) * 1_000_000
			} else if cmdFlag.Names()[0] == f.Name {
				return ctx.Uint64(f.Name)
			}

		case cli.Int64Flag:
			if cmdFlag.Names()[0] == f.Name {
				return ctx.Int64(f.Name)
			}

		case cli.StringFlag:
			if cmdFlag.Names()[0] == f.Name {
				return ctx.String(f.Name)
			}

		case cli.PathFlag:
			if cmdFlag.Names()[0] == f.Name {
				return ctx.Path(f.Name)
			}

		case cli.BoolFlag:
			if cmdFlag.Names()[0] == f.Name {
				return ctx.Bool(f.Name)
			}
		case cli.StringSliceFlag:
			if cmdFlag.Names()[0] == f.Name {
				return ctx.StringSlice(f.Name)
			}
		}
	}

	// If flag not found, return the default value of the flag
	switch f := flag.(type) {
	case cli.IntFlag:
		return f.Value
	case cli.Uint64Flag:
		return f.Value
	case cli.Int64Flag:
		return f.Value
	case cli.StringFlag:
		return f.Value
	case cli.PathFlag:
		return f.Value
	case cli.BoolFlag:
		return f.Value
	case cli.StringSliceFlag:
		if f.Value == nil {
			return []string{}
		}
		return f.Value.Value()
	}

	return nil
}
