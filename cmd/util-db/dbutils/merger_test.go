package dbutils

import (
	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/substate/db"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestMerger(t *testing.T) {
	path1 := t.TempDir() + "/aida-db"
	sdb1, err := db.NewDefaultSubstateDB(path1)
	require.NoError(t, err)
	require.NoError(t, sdb1.SetSubstateEncoding(db.ProtobufEncodingSchema))

	// Put some substates
	putSubstate(t, sdb1, 1)
	putSubstate(t, sdb1, 2)
	putSubstate(t, sdb1, 3)

	path2 := t.TempDir() + "/merge-db"
	sdb2, err := db.NewDefaultSubstateDB(path2)
	require.NoError(t, err)
	require.NoError(t, sdb2.SetSubstateEncoding(db.ProtobufEncodingSchema))

	// Put some substates
	putSubstate(t, sdb2, 4)
	putSubstate(t, sdb2, 5)
	putSubstate(t, sdb2, 6)

	cfg := &utils.Config{
		AidaDb:          path1,
		LogLevel:        "CRITICAL",
		SkipMetadata:    false,
		DeleteSourceDbs: true,
		CompactDb:       true,
	}

	md := &utils.AidaDbMetadata{}
	md.SetLogger(logger.NewLogger("CRITICAL", "MergerTest"))

	merger := NewMerger(cfg, sdb1, []db.BaseDB{sdb2}, []string{path2}, md)
	require.NoError(t, merger.Merge())

	lastSS, err := sdb1.GetLastSubstate()
	require.NoError(t, err)
	require.Equal(t, uint64(6), lastSS.Env.Number)

	require.NoError(t, merger.FinishMerge())
	merger.CloseSourceDbs()
}

func putSubstate(t *testing.T, sdb db.SubstateDB, block uint64) {
	t.Helper()
	ss := CreateEmptySubstate()
	ss.Env.Number = block
	ss.Block = block
	err := sdb.PutSubstate(ss)
	require.NoError(t, err)
}
