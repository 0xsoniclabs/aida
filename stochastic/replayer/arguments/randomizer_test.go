package arguments

import (
	"math/rand"
	"testing"

	"github.com/0xsoniclabs/aida/stochastic"
)

// TestRandomizer_FailNewRandomizer tests the failing NewRandomizer
func TestRandomizer_FailNewRandomizer(t *testing.T) {
	rg := rand.New(rand.NewSource(1))
	q_pmf := make([]float64, 1)
	a_cdf := [][2]float64{{0.0, 0.0}, {1.0, 1.0}}
	_, err := NewRandomizer(rg, q_pmf, a_cdf)
	if err == nil {
		t.Fatalf("expected to fail")
	}
}

// TestRandomizer_Simple tests the NewRandomizer, SampleArg, and SampleQueue functions.
func TestRandomizer_Simple(t *testing.T) {
	rg := rand.New(rand.NewSource(1))
	q_pmf := make([]float64, stochastic.QueueLen)
	x := 1.0 / float64(stochastic.QueueLen)
	for i := range stochastic.QueueLen {
		q_pmf[i] = x
	}
	n := int64(100)
	a_cdf := [][2]float64{{0.0, 0.0}, {1.0, 1.0}}
	r, err := NewRandomizer(rg, q_pmf, a_cdf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r == nil {
		t.Fatalf("unexpected nil EmpiricalRandomizer")
	}
	for range 10000 {
		v := r.SampleArg(n)
		if v < 0 || v >= n {
			t.Fatalf("sampled argument value out of range: %d", v)
		}
	}
	for range 10000 {
		if v := r.SampleQueue(); v < 1 || v >= stochastic.QueueLen {
			t.Fatalf("sampled queue value out of range [1,%d): %d", stochastic.QueueLen, v)
		}
	}
}

// TestRandomizer_SampleQueueRange ensures SampleQueue stays within [1,QueueLen-1].
func TestRandomizer_SampleQueueRange(t *testing.T) {
	rg := rand.New(rand.NewSource(1337))

	// Valid q_pmf: pdf[0] in (0,1), others positive and <1; shape doesn't matter for range.
	q_pmf := make([]float64, stochastic.QueueLen)
	q_pmf[0] = 0.1
	rest := 0.9 / float64(stochastic.QueueLen-1)
	for i := 1; i < len(q_pmf); i++ {
		q_pmf[i] = rest
	}

	// Simple a_cdf; not used by SampleQueue but required by constructor.
	a_cdf := [][2]float64{{0.0, 0.0}, {1.0, 1.0}}

	r, err := NewRandomizer(rg, q_pmf, a_cdf)
	if err != nil {
		t.Fatalf("unexpected error constructing EmpiricalRandomizer: %v", err)
	}

	for range 1000 {
		v := r.SampleQueue()
		if v < 1 || v >= stochastic.QueueLen {
			t.Fatalf("queue index out of range: %d", v)
		}
	}
}
