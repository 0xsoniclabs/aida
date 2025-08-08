package dbutils

import (
	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/substate/db"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"testing"
)

func TestFindDbHashOnline(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockLogger := logger.NewMockLogger(ctrl)
	mockDb, _ := GenerateTestAidaDb(t)
	utils.AidaDbRepositoryUrl = utils.AidaDbRepositorySonicUrl

	mockLogger.EXPECT().Noticef("looking for db-hash online on %v", utils.AidaDbRepositorySonicUrl)
	mockLogger.EXPECT().Info("METADATA: First block saved successfully")
	mockLogger.EXPECT().Info("METADATA: Last block saved successfully")

	md := &utils.AidaDbMetadata{
		Db:         mockDb,
		FirstBlock: 0,
		LastBlock:  0,
		FirstEpoch: 0,
		LastEpoch:  0,
		ChainId:    utils.SonicMainnetChainID,
		DbType:     utils.CloneType,
	}
	md.SetLogger(mockLogger)

	_, err := FindDbHashOnline(utils.SonicMainnetChainID, mockLogger, md)
	require.ErrorContains(t, err, "could not find db-hash for your db range")
}

func TestGenerateDbHash(t *testing.T) {
	mockDb, _ := GenerateTestAidaDb(t)
	dbHash, err := GenerateDbHash(mockDb, "WARNING")
	require.NoError(t, err)
	require.NotEmpty(t, dbHash)
}

func TestGeneratePrefixHash(t *testing.T) {
	mockDb, _ := GenerateTestAidaDb(t)
	dbHash, err := GeneratePrefixHash(mockDb, db.SubstateDBPrefix, "WARNING")
	require.NoError(t, err)
	require.NotEmpty(t, dbHash)
}
