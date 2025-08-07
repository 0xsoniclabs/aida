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

package main

import (
	"github.com/0xsoniclabs/aida/cmd/util-db/clone"
	"github.com/0xsoniclabs/aida/cmd/util-db/compact"
	"github.com/0xsoniclabs/aida/cmd/util-db/generate"
	"github.com/0xsoniclabs/aida/cmd/util-db/info"
	"github.com/0xsoniclabs/aida/cmd/util-db/merge"
	"log"
	"os"

	"github.com/0xsoniclabs/aida/cmd/util-db/db"
	"github.com/0xsoniclabs/aida/cmd/util-db/primer"

	"github.com/urfave/cli/v2"
)

// UtilDbApp data structure
var UtilDbApp = cli.App{
	Name:      "Aida Database",
	HelpName:  "util-db",
	Usage:     "merge source data into profiling database",
	Copyright: "(c) 2022 Fantom Foundation",
	Commands: []*cli.Command{
		&clone.Command,
		&compact.Command,
		&db.UpdateCommand,
		&info.Command,
		&db.ValidateCommand,
		&generate.EthereumGenesisCommand,
		&generate.DeletedAccountsCommand,
		&generate.DbHashCommand,
		&merge.Command,
		&db.UpdateCommand,
		&db.PrintDbHashCommand,
		&db.ScrapeCommand,
		&db.MetadataCommand,

		//Priming only
		&primer.RunPrimerCmd,
	},
}

// main implements aida-db functions
func main() {
	if err := UtilDbApp.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
