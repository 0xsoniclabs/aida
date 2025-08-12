package info

import (
	"fmt"
	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/substate/db"
	"github.com/urfave/cli/v2"
	"strconv"
)

var printExceptionsCommand = cli.Command{
	Action:    printExceptionsAction,
	Name:      "exception",
	Usage:     "Prints exceptions for given block number.",
	ArgsUsage: "<BlockNum>",
	Flags: []cli.Flag{
		&utils.AidaDbFlag,
	},
}

func printExceptionsAction(ctx *cli.Context) error {
	cfg, argErr := utils.NewConfig(ctx, utils.OneToNArgs)
	if argErr != nil {
		return argErr
	}

	log := logger.NewLogger(cfg.LogLevel, "AidaDb-PrintException")

	if ctx.Args().Len() != 1 {
		return fmt.Errorf("printExceptionsAction command requires exactly 1 argument")
	}
	blockNumStr := ctx.Args().Slice()[0]
	blockNum, err := strconv.ParseUint(ctx.Args().Slice()[0], 10, 64)
	if err != nil {
		return fmt.Errorf("cannot parse block number %s; %v", blockNumStr, err)
	}

	return printExceptionForBlock(cfg, log, blockNum)
}

func printExceptionForBlock(cfg *utils.Config, log logger.Logger, blockNum uint64) error {
	exceptionDb, err := db.NewReadOnlyExceptionDB(cfg.AidaDb)
	if err != nil {
		return fmt.Errorf("cannot open aida-db; %w", err)
	}

	defer func() {
		err = exceptionDb.Close()
		if err != nil {
			log.Warningf("Error closing aida db: %v", err)
		}
	}()

	exception, err := exceptionDb.GetException(blockNum)
	if err != nil {
		return fmt.Errorf("cannot get exception for block %d; %v", blockNum, err)
	}

	log.Noticef("Exception for block %v: %v", blockNum, exception)
	return nil
}
