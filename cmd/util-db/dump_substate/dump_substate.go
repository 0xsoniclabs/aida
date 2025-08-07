package dump_substate

import (
	"fmt"
	"github.com/0xsoniclabs/aida/utildb"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/substate/db"
	"github.com/urfave/cli/v2"
)

// Command returns content in substates in json format
var Command = cli.Command{
	Action:    dumpSubstateAction,
	Name:      "dump-substate",
	Usage:     "returns content in substates in json format",
	ArgsUsage: "<blockNumFirst> <blockNumLast>",
	Flags: []cli.Flag{
		&utils.WorkersFlag,
		&utils.AidaDbFlag,
		&utils.SubstateEncodingFlag,
	},
	Description: `
The aida-vm dump command requires two arguments:
<blockNumFirst> <blockNumLast>

<blockNumFirst> and <blockNumLast> are the first and
last block of the inclusive range of blocks to replay transactions.`,
}

// dumpSubstateAction prepares config and arguments before SubstateDumpAction
func dumpSubstateAction(ctx *cli.Context) error {
	var err error

	cfg, err := utils.NewConfig(ctx, utils.BlockRangeArgs)
	if err != nil {
		return err
	}

	// prepare substate database
	sdb, err := db.NewReadOnlySubstateDB(cfg.AidaDb)
	if err != nil {
		return fmt.Errorf("cannot open aida-db; %w", err)
	}
	defer sdb.Close()

	// set encoding type
	err = sdb.SetSubstateEncoding(cfg.SubstateEncoding)
	if err != nil {
		return fmt.Errorf("cannot set substate encoding; %w", err)
	}

	// run substate dump task
	taskPool := sdb.NewSubstateTaskPool("aida-vm dump", utildb.SubstateDumpTask, cfg.First, cfg.Last, ctx)
	err = taskPool.Execute()

	return err
}
