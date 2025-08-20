package generate

import "github.com/urfave/cli/v2"

// Command is a set of subcommands for generating various database-related stuff.
var Command = cli.Command{
	Name:  "generate",
	Usage: `Used for generating data.`,
	Subcommands: []*cli.Command{
		&generateDbHashCommand,
		&generateDeletedAccountsCommand,
		&generateEthereumGenesisCommand,
	},
}
