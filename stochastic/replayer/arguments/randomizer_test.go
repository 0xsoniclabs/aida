package arguments

import (
    "math/rand"
    "testing"

    "github.com/0xsoniclabs/aida/stochastic"
)

// TestEmpiricalRandomizer_SampleArgRange ensures SampleArg stays within [0,n].
func TestEmpiricalRandomizer_SampleArgRange(t *testing.T) {
    rg := rand.New(rand.NewSource(42))

    // Simple ECDF representing a uniform distribution.
    ecdf := [][2]float64{{0.0, 0.0}, {1.0, 1.0}}

    // Construct a valid qpdf for queue sampling: pdf[0] in (0,1), others positive and <1.
    qpdf := make([]float64, stochastic.QueueLen)
    qpdf[0] = 0.2
    rest := 0.8 / float64(stochastic.QueueLen-1)
    for i := 1; i < len(qpdf); i++ {
        qpdf[i] = rest
    }

    r, err := NewEmpiricalRandomizer(rg, qpdf, ecdf)
    if err != nil {
        t.Fatalf("unexpected error constructing EmpiricalRandomizer: %v", err)
    }

    n := int64(100)
    for i := 0; i < 1000; i++ {
        v := r.SampleArg(n)
        if v < 0 || v > n {
            t.Fatalf("sampled value out of range: %d (n=%d)", v, n)
        }
    }
}

// TestEmpiricalRandomizer_SampleQueueRange ensures SampleQueue stays within [1,QueueLen-1].
func TestEmpiricalRandomizer_SampleQueueRange(t *testing.T) {
    rg := rand.New(rand.NewSource(1337))

    // Valid qpdf: pdf[0] in (0,1), others positive and <1; shape doesn’t matter for range.
    qpdf := make([]float64, stochastic.QueueLen)
    qpdf[0] = 0.1
    rest := 0.9 / float64(stochastic.QueueLen-1)
    for i := 1; i < len(qpdf); i++ {
        qpdf[i] = rest
    }

    // Simple ECDF; not used by SampleQueue but required by constructor.
    ecdf := [][2]float64{{0.0, 0.0}, {1.0, 1.0}}

    r, err := NewEmpiricalRandomizer(rg, qpdf, ecdf)
    if err != nil {
        t.Fatalf("unexpected error constructing EmpiricalRandomizer: %v", err)
    }

    for i := 0; i < 1000; i++ {
        v := r.SampleQueue()
        if v < 1 || v >= stochastic.QueueLen {
            t.Fatalf("queue index out of range: %d", v)
        }
    }
}
