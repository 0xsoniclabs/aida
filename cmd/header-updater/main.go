package main

import (
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

var UpdateHeaderApp = cli.App{
	Name:      "Update Headers",
	HelpName:  "update-header",
	Usage:     "Commands to update headers in workspace.",
	Copyright: "(c) 2025 Sonic Labs",
	Commands: []*cli.Command{
		&updateYearCommand,
	},
}

// main implements aida-db functions
func main() {
	if err := UpdateHeaderApp.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
