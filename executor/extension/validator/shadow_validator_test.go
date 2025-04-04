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

package validator

import (
	"errors"
	"strings"
	"testing"

	"github.com/0xsoniclabs/aida/ethtest"
	"github.com/0xsoniclabs/aida/executor"
	"github.com/0xsoniclabs/aida/state"
	"github.com/0xsoniclabs/aida/txcontext"
	"github.com/0xsoniclabs/aida/utils"
	"go.uber.org/mock/gomock"
)

func TestShadowDbValidator_PostBlockPass(t *testing.T) {
	cfg := &utils.Config{}

	ctrl := gomock.NewController(t)
	db := state.NewMockStateDB(ctrl)

	data := ethtest.CreateTestTransaction(t)
	ctx := new(executor.Context)
	ctx.State = db
	st := executor.State[txcontext.TxContext]{Block: 1, Transaction: 1, Data: data}

	gomock.InOrder(
		db.EXPECT().GetHash(),
		db.EXPECT().Error().Return(nil),
	)

	ext := makeShadowDbValidator(cfg)

	err := ext.PostBlock(st, ctx)
	if err != nil {
		t.Fatalf("post-block cannot return error; %v", err)
	}
}

func TestShadowDbValidator_PostBlockReturnsError(t *testing.T) {
	cfg := &utils.Config{}

	ctrl := gomock.NewController(t)
	db := state.NewMockStateDB(ctrl)

	data := ethtest.CreateTestTransaction(t)
	ctx := new(executor.Context)
	ctx.State = db
	st := executor.State[txcontext.TxContext]{Block: 1, Transaction: 1, Data: data}

	expectedErr := errors.New("FAIL")

	gomock.InOrder(
		db.EXPECT().GetHash(),
		db.EXPECT().Error().Return(expectedErr),
	)

	ext := makeShadowDbValidator(cfg)

	err := ext.PostBlock(st, ctx)
	if err == nil {
		t.Fatalf("post-block must return error; %v", err)
	}

	if strings.Compare(err.Error(), expectedErr.Error()) != 0 {
		t.Fatalf("unexpected error\ngot:%v\nwant:%v", err.Error(), expectedErr.Error())
	}
}
