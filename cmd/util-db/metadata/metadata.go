package metadata

import (
	"github.com/urfave/cli/v2"
)

var Command = cli.Command{
	Name:  "metadata",
	Usage: "Does action with AidaDb metadata",
	Subcommands: []*cli.Command{
		&printCommand,
		&generateCommand,
		&insertCommand,
		&removeCommand,
	},
}
