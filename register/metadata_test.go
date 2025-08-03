package register

import (
	"testing"

	"github.com/0xsoniclabs/aida/utils"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestRegister_MakeRunMetadata(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	meta := map[string]string{
		"AppName": "testApp",
	}
	rm, err := MakeRunMetadata(":memory:", MakeRunIdentity(0, &utils.Config{
		ArchiveMode:      true,
		ArchiveQueryRate: 99,
	}), func() (map[string]string, error) {
		return meta, nil
	})
	assert.NoError(t, err)
	assert.NotNil(t, rm)
	assert.Equal(t, meta["AppName"], rm.Meta["AppName"])
	assert.Equal(t, "true", rm.Meta["ArchiveMode"])
	assert.Equal(t, "99", rm.Meta["ArchiveQueryRate"])

}

func TestRunMetadata_Print(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPrinter := utils.NewMockPrinter(ctrl)
	meta := &RunMetadata{
		Meta: map[string]string{},
		Ps:   utils.NewCustomPrinters([]utils.Printer{mockPrinter}),
	}
	mockPrinter.EXPECT().Print()
	meta.Print()
}

func TestRunMetadata_Close(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPrinter := utils.NewMockPrinter(ctrl)
	meta := &RunMetadata{
		Meta: map[string]string{},
		Ps:   utils.NewCustomPrinters([]utils.Printer{mockPrinter}),
	}
	mockPrinter.EXPECT().Close()
	meta.Close()
}

func TestRunMetadata_sqlite3(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	meta := map[string]string{
		"AppName": "testApp",
	}
	rm := &RunMetadata{
		Meta: meta,
		Ps:   utils.NewPrinters(),
	}
	a, b, c, d := rm.sqlite3(":memory:")
	assert.Equal(t, ":memory:", a)
	assert.NotNil(t, b)
	assert.NotNil(t, c)
	assert.NotNil(t, d)
	out := d()
	assert.Equal(t, len(meta), len(out))
}

func TestRegister_FetchUnixInfo(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	info, err := FetchUnixInfo()
	assert.NoError(t, err)
	assert.NotNil(t, info)
	assert.Equal(t, 10, len(info))
}
