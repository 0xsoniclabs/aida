package generate

import "github.com/urfave/cli/v2"

// Command is a set of subcommands for generating various database-related stuff.
var Command = cli.Command{
	Name:  "clone",
	Usage: `Used for creation of standalone subset of aida-db or patch`,
	Subcommands: []*cli.Command{
		&generateDbHashCommand,
		&generateDeletedAccountsCommand,
		&extractEthereumGenesisCommand,
	},
}
