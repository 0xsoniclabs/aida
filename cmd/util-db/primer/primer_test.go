package primer

import (
	"path/filepath"
	"testing"

	"github.com/0xsoniclabs/aida/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func TestCmd_RunPrimerCmd(t *testing.T) {
	// given - basic priming test with default settings
	tempDir := t.TempDir()
	aidaDbPath := filepath.Join(tempDir, "aida-db")
	require.NoError(t, utils.CopyDir("../../dataset/sample-pb-db", aidaDbPath))
	app := cli.NewApp()
	app.Commands = []*cli.Command{&RunPrimerCmd}

	args := utils.NewArgs("test").
		Arg(RunPrimerCmd.Name).
		Flag(utils.AidaDbFlag.Name, aidaDbPath).
		Flag(utils.StateDbImplementationFlag.Name, "carmen").
		Flag(utils.StateDbVariantFlag.Name, "go-file").
		Arg("100"). // block number to prime to
		Build()

	// when
	err := app.Run(args)

	// then
	assert.NoError(t, err)
}
