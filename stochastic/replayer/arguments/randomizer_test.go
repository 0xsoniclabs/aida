package arguments

//type constSource64 struct{ v int64 }

//func (c *constSource64) Int63() int64    { return c.v }
//func (c *constSource64) Uint64() uint64  { return uint64(c.v) }
//func (c *constSource64) Seed(seed int64) {}


func TestExponentialArgRandomizerSampleRange(t *testing.T) {
	rg := rand.New(rand.NewSource(1))
	r := NewExponentialArgRandomizer(rg, 1.5)
	n := 100
	for range 1000 {
		v := r.SampleArg(n)
		if v < 0 || v >= n {
			t.Fatalf("sampled value out of range: %d", v)
		}
	}
}

//func TestExponentialQueueRandomizerSampleRange(t *testing.T) {
//	rg := rand.New(rand.NewSource(2))
//	r := NewExponentialQueueRandomizer(rg, 1.2)
//	for range 1000 {
//		v := r.SampleQueue()
//		if v < 1 || v >= stochastic.QueueLen {
//			t.Fatalf("queue index out of range: %d", v)
//		}
//	}
//}

//func TestEmpiricalQueueRandomizerCtorInvalidLen(t *testing.T) {
//	rg := rand.New(rand.NewSource(3))
//	bad := make([]float64, stochastic.QueueLen-1)
//	if r := NewEmpiricalQueueRandomizer(rg, bad); r != nil {
//		t.Fatalf("expected nil for invalid length")
//	}
//}

//func TestEmpiricalQueueRandomizerCtorInvalidFactor(t *testing.T) {
//	rg := rand.New(rand.NewSource(4))
//	bad := make([]float64, stochastic.QueueLen)
//	bad[0] = 1.0
//	if r := NewEmpiricalQueueRandomizer(rg, bad); r != nil {
//		t.Fatalf("expected nil for qpdf[0]>=1.0")
//	}
//}

//func TestEmpiricalQueueRandomizerSampleDeterministic(t *testing.T) {
//	rg := rand.New(rand.NewSource(5))
//	qpdf := make([]float64, stochastic.QueueLen)
//	qpdf[0] = 0.0
//	k := 5
//	qpdf[k+1] = 1.0
//	r := NewEmpiricalQueueRandomizer(rg, qpdf)
//	if r == nil {
//		t.Fatalf("unexpected nil EmpiricalQueueRandomizer")
//	}
//	for range 20 {
//		if v := r.SampleQueue(); v != 1+k {
//			t.Fatalf("expected %d, got %d", 1+k, v)
//		}
//	}
//}

//func TestEmpiricalQueueRandomizerSampleLastPositiveFallback(t *testing.T) {
//	rg := rand.New(&constSource64{v: (1<<63 - 1) - (1 << 20)})
//	qpdf := make([]float64, stochastic.QueueLen)
//	qpdf[0] = 0.5
//	qpdf[1] = 0.2
//	qpdf[2] = 0.2
//	r := NewEmpiricalQueueRandomizer(rg, qpdf)
//	if r == nil {
//		t.Fatalf("unexpected nil EmpiricalQueueRandomizer")
//	}
//	if v := r.SampleQueue(); v != 2 {
//		t.Fatalf("expected lastPositive fallback return 2, got %d", v)
//	}
//}

//func TestEmpiricalQueueRandomizerSampleAllZeroDefault(t *testing.T) {
//	rg := rand.New(&constSource64{v: (1<<63 - 1) - (1 << 20)})
//	qpdf := make([]float64, stochastic.QueueLen)
//	r := NewEmpiricalQueueRandomizer(rg, qpdf)
//	if r == nil {
//		t.Fatalf("unexpected nil EmpiricalQueueRandomizer")
//	}
//	if v := r.SampleQueue(); v != 1 {
//		t.Fatalf("expected default return 1, got %d", v)
//	}
//}
