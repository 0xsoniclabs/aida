package arguments

import (
	"math/rand"
	"testing"

	"github.com/0xsoniclabs/aida/stochastic"
	"github.com/0xsoniclabs/aida/stochastic/statistics/exponential"
)

func TestEmpiricalRandomizer(t *testing.T) {
	rg := rand.New(rand.NewSource(1))
	qpdf := make([]float64, stochastic.QueueLen)
	qpdf[0] = 0.5
	k := 5
	qpdf[k+1] = 0.5
	n := int64(100)
	ecdf := exponential.ToECDF(5.0, stochastic.NumECDFPoints)
	r, err := NewEmpiricalRandomizer(rg, qpdf, ecdf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r == nil {
		t.Fatalf("unexpected nil EmpiricalRandomizer")
	}
	for range 1000 {
		v := r.SampleArg(n)
		if v < 0 || v >= n {
			t.Fatalf("sampled value out of range: %d", v)
		}
	}
	for range 20 {
		if v := r.SampleQueue(); v != 1+k {
			t.Fatalf("expected %d, got %d", 1+k, v)
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
