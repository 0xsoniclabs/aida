package stochastic

import (
	"log"

	"github.com/Fantom-foundation/Aida/stochastic"
	"github.com/Fantom-foundation/Aida/utils"
	"github.com/urfave/cli/v2"
)

// StochasticGenerateCommand data structure for the record app
var StochasticGenerateCommand = cli.Command{
	Action:    stochasticGenerateAction,
	Name:      "generate",
	Usage:     "generate uniform events file",
	ArgsUsage: "",
	Flags: []cli.Flag{
		&utils.BlockLengthFlag,
		&utils.SyncPeriodLengthFlag,
		&utils.OperationFrequency,
		&utils.ContractNumberFlag,
		&utils.KeysNumberFlag,
		&utils.ValuesNumberFLag,
		&utils.SnapshotDepthFlag,
	},
	Description: "The stochastic produces an events.json file with uniform parameters",
}

// stochasticGenerateAction generates the uniform simulation data and writes the JSON file.
func stochasticGenerateAction(ctx *cli.Context) error {
	var err error

	cfg, err := utils.NewConfig(ctx, utils.NoArgs)
	if err != nil {
		return err
	}

	// create a new uniformly distributed event registry
	eventRegistry := stochastic.GenerateUniformRegistry(cfg)

	// writing event registry
	log.Printf("write events file ...\n")
	outputFileName := ctx.String(utils.OutputFlag.Name)
	if outputFileName == "" {
		outputFileName = "./events.json"
	}
	WriteEvents(eventRegistry, outputFileName)

	return nil
}
