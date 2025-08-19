package update

import (
	"github.com/urfave/cli/v2"
)

var Command = cli.Command{
	Name:  "update",
	Usage: "Update the database.",
	Subcommands: []*cli.Command{
		&updateDbCommand,
		&updateHashesCommand,
	},
}
