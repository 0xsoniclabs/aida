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

package utils

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
)

var StateHashQueue StateHashQueueProvider

func MakeStateHashQueueProvider() StateHashProvider {
	StateHashQueue = StateHashQueueProvider{stateHashQueue: make(chan string, 10)}
	return &StateHashQueue
}

type StateHashQueueProvider struct {
	stateHashQueue chan string
}

func (p *StateHashQueueProvider) GetStateHash(number int) (common.Hash, error) {
	stateRoot, ok := <-p.stateHashQueue
	if !ok {
		return common.Hash{}, fmt.Errorf("unexpected end of state hash queue")
	}

	return common.HexToHash(stateRoot), nil
}

func (p *StateHashQueueProvider) AddStateHash(stateHash string) {
	p.stateHashQueue <- stateHash[:]
}
