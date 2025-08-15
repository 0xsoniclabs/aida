package clone

import (
	"testing"
	"time"

	"github.com/0xsoniclabs/substate/db"
	"go.uber.org/mock/gomock"
)

func TestCloner_stop(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockDbSrc := db.NewMockBaseDB(ctrl)
	mockDbTarget := db.NewMockBaseDB(ctrl)

	mockDbSrc.EXPECT().Close()
	mockDbTarget.EXPECT().Close()

	ch := make(chan any)
	c := &cloner{
		aidaDb:  mockDbSrc,
		cloneDb: mockDbTarget,
		stopCh:  ch,
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
