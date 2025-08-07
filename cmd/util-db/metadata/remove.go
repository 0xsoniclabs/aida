package metadata

import (
	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/substate/db"
	"github.com/urfave/cli/v2"
)

// removeCommand is a command used for creating testing environment without metadata
var removeCommand = cli.Command{
	Action: removeAction,
	Name:   "remove",
	Usage:  "remove metadata from aidaDb",
	Flags: []cli.Flag{
		&utils.AidaDbFlag,
	},
	Description: `
Removes block and epoch range and ChainID from metadata for given AidaDb.
`,
}

// removeAction command is used for testing scenario where AidaDb does not have metadata and a patch
// is applied onto it
func removeAction(ctx *cli.Context) error {
	aidaDbPath := ctx.String(utils.AidaDbFlag.Name)

	// open db
	base, err := db.NewDefaultBaseDB(aidaDbPath)
	if err != nil {
		return err
	}

	defer base.Close()
	md := utils.NewAidaDbMetadata(base, "DEBUG")
	md.DeleteMetadata()

	return nil
}
