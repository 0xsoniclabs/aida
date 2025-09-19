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

package markov

import (
	"fmt"
	"math"

	"github.com/0xsoniclabs/aida/stochastic/statistics/discrete"
	"gonum.org/v1/gonum/mat"
)

const (
	estimationEps = 1e-9 // epsilon for stationary distribution
)

// Chain
type Chain struct {
	n int         // number of states
	a [][]float64 // stochastic matrix
	l []string    // labels of states
}

// New creates a new MarkovChain from a stochastic matrix and a list of labels.
// The matrix must be square and the number of labels must match the number of rows/columns.
// Each row must sum to one and contain only values in the interval [0,1].
// Each label must be unique.
func New(a [][]float64, labels []string) (*Chain, error) {
	// check uniqueness of labels
	n := len(labels)
	labelCount := map[string]int{}
	for i := range n {
		labelCount[labels[i]]++
	}
	for k, c := range labelCount {
		if c > 1 {
			return nil, fmt.Errorf("New: the state (%v) occurs more than once (%v)", k, c)
		}
	}

	// check markov property of matrix: (1) nxn matrix, (2) rows sum to one, (3) all values in [0,1].
	if len(a) != n {
		return nil, fmt.Errorf("New: number of labels (%v) mismatches number of rows (%v)", n, len(a))
	}
	for i := range n {
		if len(a[i]) != n {
			return nil, fmt.Errorf("New: number of columns (%v) in row (%v) is not equal to the number of labels (%v)", len(a[i]), i, n)
		}
		total := 0.0
		for j := range n {
			if !(a[i][j] >= 0.0 && a[i][j] <= 1.0) {
				return nil, fmt.Errorf("New: invalid probability (%v) in row %v, column %v", a[i][j], i, j)
			}
			total += a[i][j]
		}
		if math.Abs(total-1.0) > 1e-9 {
			return nil, fmt.Errorf("New: row %v does not sum to one (%v)", i, total)
		}
	}
	return &Chain{a: a, l: labels, n: n}, nil
}

// Sample the next state in a markov chain for a given state i.
func (mc Chain) Sample(i int, x float64) (int, error) {
	if x < 0 || x >= 1.0 {
		return 0, fmt.Errorf("probabilistic argument (%v) is not in interval [0,1]", x)
	}
	if i < 0 || i >= mc.n {
		return 0, fmt.Errorf("Sample: state index (%v) out of range", i)
	}
	return discrete.Quantile(mc.a[i], x), nil
}

// Stationary computes the stationary distribution of a Markov Chain.
func (mc Chain) Stationary() ([]float64, error) {
	// flatten matrix for gonum package
	elements := []float64{}
	for i := range mc.n {
		for j := range mc.n {
			elements = append(elements, mc.a[i][j])
		}
	}
	a := mat.NewDense(mc.n, mc.n, elements)

	// perform eigenvalue decomposition
	var eig mat.Eigen
	ok := eig.Factorize(a, mat.EigenLeft)
	if !ok {
		return nil, fmt.Errorf("eigen-value decomposition failed")
	}

	// find index for eigenvalue of one
	// (note that it is not necessarily the first index)
	v := eig.Values(nil)
	k := -1
	for i, eigenValue := range v {
		if math.Abs(real(eigenValue)-1.0) < estimationEps && math.Abs(imag(eigenValue)) < estimationEps {
			k = i
		}
	}
	if k == -1 {
		return nil, fmt.Errorf("eigen-decomposition failed; no eigenvalue of one found")
	}

	// find left eigenvectors of decomposition
	var ev mat.CDense
	eig.LeftVectorsTo(&ev)

	// compute total for eigenvector with eigenvalue of one.
	total := complex128(0)
	for i := range mc.n {
		total += ev.At(i, k)
	}
	if imag(total) > estimationEps {
		return nil, fmt.Errorf("eigen-decomposition failed; eigen-vector is a complex number")
	}

	// normalize eigenvector by total
	stationary := []float64{}
	for i := range mc.n {
		stationary = append(stationary, math.Abs(real(ev.At(i, k))/real(total)))
	}
	return stationary, nil
}

func (mc Chain) Label(i int) (string, error) {
	if i < 0 || i >= mc.n {
		return "", fmt.Errorf("Label: state is out of range (%v)", i)
	}
	return mc.l[i], nil
}

// Find the state index for a given label.
func (mc Chain) Find(label string) (int, error) {
	for i := range mc.l {
		if mc.l[i] == label {
			return i, nil
		}
	}
	return -1, nil
}
