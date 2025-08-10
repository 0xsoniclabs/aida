package updateset

//
//import (
//	"path/filepath"
//	"testing"
//
//	"github.com/0xsoniclabs/aida/utils"
//	"github.com/stretchr/testify/assert"
//	"github.com/stretchr/testify/require"
//	"github.com/urfave/cli/v2"
//)
//
//func TestCmd_RunGenUpdateSetCommand(t *testing.T) {
//	// given - basic priming test with default settings
//	tempDir := t.TempDir()
//	aidaDbPath := filepath.Join(tempDir, "aida-db")
//	require.NoError(t, utils.CopyDir("../../dataset/sample-rlp-db", aidaDbPath))
//	app := cli.NewApp()
//	app.Commands = []*cli.Command{&GenUpdateSetCommand}
//
//	args := utils.NewArgs("test").
//		Arg(GenUpdateSetCommand.Name).
//		Flag(utils.AidaDbFlag.Name, aidaDbPath).
//		Arg("100").
//		Arg("100").
//		Build()
//
//	// when
//	err := app.Run(args)
//
//	// then
//	assert.NoError(t, err)
//}
