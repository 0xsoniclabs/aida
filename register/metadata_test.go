// Copyright 2025 Sonic Labs
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
	err := meta.Print()
	assert.NoError(t, err)
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
	err := meta.Close()
	assert.NoError(t, err)
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

	info, _ := FetchUnixInfo()
	// error can happen based on running environment
	assert.NotNil(t, info)
	assert.Equal(t, 10, len(info))
}
