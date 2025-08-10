package db

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/0xsoniclabs/aida/cmd/util-db/flags"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

// TestCmd_RunMergeCommand tests the MergeCommand
func TestCmd_RunMergeCommand(t *testing.T) {
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "output-db")
	sourceDb1 := filepath.Join(tempDir, "aida-db-1")
	sourceDb2 := filepath.Join(tempDir, "aida-db-2")
	require.NoError(t, utils.CopyDir("../../dataset/aida-db-0-1k-protobuf", sourceDb1))
	require.NoError(t, utils.CopyDir("../../dataset/aida-db-0-1k-protobuf", sourceDb2))

	// given
	app := cli.NewApp()
	app.Commands = []*cli.Command{&MergeCommand}
	args := utils.NewArgs("test").
		Arg(MergeCommand.Name).
		Flag(utils.SubstateEncodingFlag.Name, "protobuf").
		Flag(utils.AidaDbFlag.Name, outputPath).
		Flag(flags.SkipMetadata.Name, true).
		Arg(sourceDb1).
		Arg(sourceDb2).
		Build()

	// when
	err := app.Run(args)

	// then
	assert.NoError(t, err)
	_, err = os.Stat(outputPath)
	assert.NoError(t, err)
}
