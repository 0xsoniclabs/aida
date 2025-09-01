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
	"errors"

	"github.com/0xsoniclabs/aida/stochastic/statistics"
)

// SingleUseArgumentSet data structure for generating non-reusable arguments.
// This is an extension of the ArgumentSet data structure.
// Deleted arguments are not reused and cannot be chosen anymore.
// This type of argument set is needed for self-destructing account addresses.
type SingleUseArgumentSet struct {
	argset      *ArgumentSet   // underlying random argument set
	ctr         ArgumentType   // counter
	translation []ArgumentType // translation table for argument
}

// NewSingleUseArgumentSet creates a new argument set whose arguments,
// when deleted, cannot be reused.
func NewSingleUseArgumentSet(argset *ArgumentSet) *SingleUseArgumentSet {
	t := make([]ArgumentType, argset.n)
	for i := range argset.n {
		t[i] = i + 1 // shifted by one because of zero value
	}
	return &SingleUseArgumentSet{
		argset:      argset,
		ctr:         argset.n,
		translation: t,
	}
}

// Choose an argument from the argument set according to the kind of argument.
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
			return 0, errors.New("translation index out of range")
		}
		return a.translation[v-1], nil
	}
}

// Remove deletes an indirect index.
func (a *SingleUseArgumentSet) Remove(k ArgumentType) error {
	if k == 0 {
		return nil
	}

	// find argument in translation table
	i := a.find(k)
	if i < 0 {
		return errors.New("index not found")
	}

	// delete index i from the translation table and the random access generator.
	a.translation = append(a.translation[:i], a.translation[i+1:]...)
	if err := a.argset.Remove(i); err != nil {
		return err
	}

	return nil
}

// find finds the index in the translation table for a given index k.
func (a *SingleUseArgumentSet) find(k ArgumentType) ArgumentType {
	for i := int64(0); i < int64(len(a.translation)); i++ {
		if a.translation[i] == k {
			return i
		}
	}
	return -1
}

// Size returns the number of arguments in the argument set.
func (a *SingleUseArgumentSet) Size() ArgumentType {
	return a.argset.n
}
