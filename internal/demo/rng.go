package demo

import "math"

// rng is a seeded pseudo-random number generator (SplitMix64): small, fast,
// and the same sequence for the same seed on every platform. It is not for
// anything secret.
type rng struct {
	state uint64
}

// SplitMix64's constants.
const (
	splitMixGamma  = 0x9e3779b97f4a7c15
	splitMixMul1   = 0xbf58476d1ce4e5b9
	splitMixMul2   = 0x94d049bb133111eb
	splitMixShift1 = 30
	splitMixShift2 = 27
	splitMixShift3 = 31
)

// newRNG returns a generator seeded with seed.
func newRNG(seed uint64) *rng { return &rng{state: seed} }

// next returns the next 64 random bits.
func (r *rng) next() uint64 {
	r.state += splitMixGamma
	z := r.state
	z = (z ^ (z >> splitMixShift1)) * splitMixMul1
	z = (z ^ (z >> splitMixShift2)) * splitMixMul2
	return z ^ (z >> splitMixShift3)
}

// intn returns a number in [0, n); 0 when n is not positive. The modulo
// bias is negligible for the small n used here.
func (r *rng) intn(n int64) int64 {
	if n <= 0 {
		return 0
	}
	v := r.next() % uint64(n)
	if v > math.MaxInt64 {
		return 0 // unreachable: v < n
	}
	return int64(v)
}

// between returns a number in [lo, hi], both inclusive.
func (r *rng) between(lo, hi int64) int64 { return lo + r.intn(hi-lo+1) }
