// Copyright 2024 Fantom Foundation
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

package statedb

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/0xsoniclabs/aida/executor"
	"github.com/0xsoniclabs/aida/state"
	"go.uber.org/mock/gomock"
)

func TestArchivePrepper_ArchiveGetsReleasedInPostBlock(t *testing.T) {
	ext := MakeArchivePrepper[any]()

	ctrl := gomock.NewController(t)
	db := state.NewMockStateDB(ctrl)
	archive := state.NewMockNonCommittableStateDB(ctrl)

	gomock.InOrder(
		db.EXPECT().GetArchiveState(uint64(1)).Return(archive, nil),
		archive.EXPECT().Release(),
	)

	state := executor.State[any]{
		Block: 2,
	}
	ctx := &executor.Context{
		State: db,
	}
	if err := ext.PreBlock(state, ctx); err != nil {
		t.Fatalf("failed to to run pre-block: %v", err)
	}
	if err := ext.PostBlock(state, ctx); err != nil {
		t.Fatalf("failed to to run post-block: %v", err)
	}
}

func TestMakeArchivePrepper(t *testing.T) {
	ext := MakeArchivePrepper[any]()
	assert.NotNil(t, ext)
}

func TestArchivePrepper_PreBlock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockStateDB(ctrl)
	mockDb.EXPECT().GetArchiveState(gomock.Any()).Return(nil, nil)
	a := archivePrepper[string]{}
	err := a.PreBlock(executor.State[string]{}, &executor.Context{
		State: mockDb,
	})
	assert.Nil(t, err)
}

func TestArchivePrepper_PreTransaction(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockNonCommittableStateDB(ctrl)
	mockDb.EXPECT().BeginTransaction(gomock.Any()).Return(nil)
	a := archivePrepper[string]{}
	err := a.PreTransaction(executor.State[string]{}, &executor.Context{
		Archive: mockDb,
	})
	assert.Nil(t, err)
}

func TestArchivePrepper_PostTransaction(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockNonCommittableStateDB(ctrl)
	mockDb.EXPECT().EndTransaction().Return(nil)
	a := archivePrepper[string]{}
	err := a.PostTransaction(executor.State[string]{}, &executor.Context{
		Archive: mockDb,
	})
	assert.Nil(t, err)
}
func TestArchivePrepper_PostBlock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDb := state.NewMockNonCommittableStateDB(ctrl)
	mockDb.EXPECT().Release().Return(nil)
	a := archivePrepper[string]{}
	err := a.PostBlock(executor.State[string]{}, &executor.Context{
		Archive: mockDb,
	})
	assert.Nil(t, err)
}
