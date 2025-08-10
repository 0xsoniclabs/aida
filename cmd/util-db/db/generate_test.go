package db

//
//import (
//	"os"
//	"path/filepath"
//	"testing"
//
//	"github.com/0xsoniclabs/aida/utils"
//	"github.com/stretchr/testify/assert"
//	"github.com/stretchr/testify/require"
//	"github.com/urfave/cli/v2"
//)
//
//func TestCmd_GenerateCommand(t *testing.T) {
//	// given
//	tempDir := t.TempDir()
//	aidaDbPath := filepath.Join(tempDir, "aida-db")
//	require.NoError(t, utils.CopyDir("../../dataset/sample-rlp-db", aidaDbPath))
//	outputPath := filepath.Join(tempDir, "output")
//
//	app := cli.NewApp()
//	app.Commands = []*cli.Command{&GenerateCommand}
//
//	args := utils.NewArgs("test").
//		Arg(GenerateCommand.Name).
//		Flag(utils.AidaDbFlag.Name, aidaDbPath).
//		Flag(utils.ChainIDFlag.Name, int(utils.MainnetChainID)).
//		Flag(utils.OutputFlag.Name, outputPath).
//		Flag(utils.WorkersFlag.Name, 1).
//		Arg("1"). // firstBlock
//		Arg("1"). // lastBlock
//		Arg("1"). // firstEpoch
//		Arg("1"). // lastEpoch
//		Build()
//
//	// when
//	err := app.Run(args)
//
//	// then
//	assert.NoError(t, err)
//	_, err = os.Stat(outputPath)
//	assert.NoError(t, err)
//
//}
