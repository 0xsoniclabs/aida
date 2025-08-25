package validate

import (
	"fmt"
	"testing"

	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/substate/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func TestCmd_ValidateCommand(t *testing.T) {
	// given
	_, aidaDbPath := utils.CreateTestSubstateDb(t)
	app := cli.NewApp()
	app.Commands = []*cli.Command{&Command}

	args := utils.NewArgs("test").
		Arg(Command.Name).
		Flag(utils.AidaDbFlag.Name, aidaDbPath).
		Build()

	// when
	err := app.Run(args)

	// then
	assert.NoError(t, err)
}

func TestCmd_ValidateCommandError(t *testing.T) {
	tests := []struct {
		name        string
		argsBuilder *utils.ArgsBuilder
		wantErr     string
		setup       func(aidaDbPath string)
	}{
		{
			name: "CannotParseCfg",
			argsBuilder: utils.NewArgs("test").
				Arg(Command.Name).
				Flag(utils.ChainIDFlag.Name, 9990099).
				Flag(utils.AidaDbFlag.Name, ""),
			wantErr: "cannot parse config",
			setup:   func(aidaDbPath string) {},
		},
		{
			name: "WrongAidaDbType",
			argsBuilder: utils.NewArgs("test").
				Arg(Command.Name),
			wantErr: fmt.Sprintf("your db type (%v) cannot be validated", utils.NoType),
			setup: func(aidaDbPath string) {
				aidaDb, err := db.NewDefaultBaseDB(aidaDbPath)
				require.NoError(t, err)
				md := utils.NewAidaDbMetadata(aidaDb, "CRITICAL")
				err = md.SetDbType(utils.NoType)
				require.NoError(t, err)
				err = aidaDb.Close()
				require.NoError(t, err)
			},
		},
		{
			name: "NoDbHashFound",
			argsBuilder: utils.NewArgs("test").
				Arg(Command.Name),
			wantErr: "could not find expected db hash",
			setup: func(aidaDbPath string) {
				aidaDb, err := db.NewDefaultBaseDB(aidaDbPath)
				require.NoError(t, err)
				md := utils.NewAidaDbMetadata(aidaDb, "CRITICAL")
				err = md.SetDbHash([]byte{})
				require.NoError(t, err)
				err = aidaDb.Close()
				require.NoError(t, err)
			},
		},
		{
			name: "WrongDbHash",
			argsBuilder: utils.NewArgs("test").
				Arg(Command.Name),
			wantErr: "hashes are different",
			setup: func(aidaDbPath string) {
				aidaDb, err := db.NewDefaultBaseDB(aidaDbPath)
				require.NoError(t, err)
				md := utils.NewAidaDbMetadata(aidaDb, "CRITICAL")
				err = md.SetDbHash([]byte("wrong-hash"))
				require.NoError(t, err)
				err = aidaDb.Close()
				require.NoError(t, err)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, aidaDbPath := utils.CreateTestSubstateDb(t)
			test.setup(aidaDbPath)
			app := cli.NewApp()
			app.Commands = []*cli.Command{&Command}
			// when
			test.argsBuilder.Flag(utils.AidaDbFlag.Name, aidaDbPath)
			err := app.Run(test.argsBuilder.Build())

			// then
			require.ErrorContains(t, err, test.wantErr)
		})
	}

}
