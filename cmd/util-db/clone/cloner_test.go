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

package clone

import (
	"testing"
	"time"

	"github.com/0xsoniclabs/substate/db"
	"go.uber.org/mock/gomock"
)

func TestCloner_stop(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockDbSrc := db.NewMockSubstateDB(ctrl)
	mockDbTarget := db.NewMockSubstateDB(ctrl)

	mockDbSrc.EXPECT().Close()
	mockDbTarget.EXPECT().Close()

	ch := make(chan any)
	c := &cloner{
		sourceDb: mockDbSrc,
		cloneDb:  mockDbTarget,
		stopCh:   ch,
	}
	c.stop()
	ticker := time.NewTicker(time.Second)
	select {
	case <-c.stopCh:
		return
	case <-ticker.C:
		t.Fatal("stop channel was not closed")
	}
}
