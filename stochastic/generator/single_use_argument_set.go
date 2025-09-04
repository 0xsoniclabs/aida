// Copyright 2024 Fantom Foundation
// This file is part of Aida Testing Infrastructure for Sonic
//
// Aida is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software
// (at your option) any later version.
//
// Aida is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with Aida. If not, see <http://www.gnu.org/licenses/>.

package generator

import (
	"fmt"

	"github.com/0xsoniclabs/aida/stochastic/statistics"
)

// SingleUseArgumentSet data structure for generating non-reusable arguments.
// This is an extension of the ArgumentSet data structure.
// Deleted arguments are not reused and cannot be chosen anymore in the future.
// This type of argument set is needed for self-destructing account addresses.
type SingleUseArgumentSet struct {
	argset      ArgumentSet
	ctr         ArgumentType   // argument counter for new arguments
	translation []ArgumentType // translation table for arguments
	ArgumentSet
}

// NewSingleUseArgumentSet creates a new argument set whose arguments,
// when deleted, will not be reused in future.
func NewSingleUseArgumentSet(argset ArgumentSet) *SingleUseArgumentSet {
	t := make([]ArgumentType, argset.Size())
	for i := range argset.Size() {
		t[i] = i
	}
	return &SingleUseArgumentSet{
		argset:      argset,
		ctr:         argset.Size(),
		translation: t,
	}
}

// Choose an argument from the argument set according to its kind.
func (a *SingleUseArgumentSet) Choose(kind int) (int64, error) {
	v, err := a.argset.Choose(kind)
	if err != nil {
		return 0, err
	}
	switch kind {
	case statistics.ZeroArgID:
		return 0, nil
	case statistics.NewArgID:
		a.ctr++
		v := a.ctr
		a.translation = append(a.translation, v)
		return v, nil
	default:
		if v <= 0 || int(v) > len(a.translation) {
			return 0, fmt.Errorf("Choose: argument %v out of range [0,%v]", v, len(a.translation))
		}
		return a.translation[v], nil
	}
}

// Remove argument k from the argument set.
func (a *SingleUseArgumentSet) Remove(k ArgumentType) error {
	if k == 0 { // zero cannot be removed
		return nil
	}
	i := a.find(k)
	if i < 0 {
		return fmt.Errorf("Remove: argument %v not found", k)
	}
	a.translation = append(a.translation[:i], a.translation[i+1:]...)
	if err := a.argset.Remove(i); err != nil {
		return err
	}
	return nil
}

// Size returns the number of arguments in the argument set.
func (a *SingleUseArgumentSet) Size() ArgumentType {
	return a.argset.Size()
}

// find the argument in the translation table for a given argument k.
func (a *SingleUseArgumentSet) find(k ArgumentType) ArgumentType {
	for i := range int64(len(a.translation)) {
		if a.translation[i] == k {
			return i
		}
	}
	return -1
}
