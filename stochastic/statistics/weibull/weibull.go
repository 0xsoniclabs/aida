// Copyright 2025 Fantom Foundation
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
package weibull

import (
    "math"
    "math/rand"
)

func CDF(lambda, k, x float64) float64 {
    if x <= 0 {
        return 0
    }
    if x >= 1 {
        return 1
    }
    num := 1 - math.Exp(-lambda*math.Pow(x, k))
    den := 1 - math.Exp(-lambda)
    return num / den
}

func PiecewiseLinearCDF(lambda, k float64, n int) [][2]float64 {
    fn := make([][2]float64, 0, n+1)
    for i := 0; i <= n; i++ {
        x := float64(i) / float64(n)
        y := CDF(lambda, k, x)
        fn = append(fn, [2]float64{x, y})
    }
    return fn
}

func Quantile(lambda, k, p float64) float64 {
    if p <= 0 {
        return 0
    }
    if p >= 1 {
        return 1
    }
    t := 1 - p + p*math.Exp(-lambda)
    v := -math.Log(t) / lambda
    if v <= 0 {
        return 0
    }
    return math.Pow(v, 1.0/k)
}

func Sample(rg *rand.Rand, lambda, k float64, n int64) int64 {
    if n <= 0 {
        return 0
    }
    x := Quantile(lambda, k, rg.Float64())
    v := int64(float64(n) * x)
    if v < 0 {
        return 0
    }
    if v >= n {
        return n - 1
    }
    return v
}