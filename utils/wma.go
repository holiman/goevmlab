package utils

import (
	"math"
	"sync/atomic"
)

// SlidingAverage implements a decaying moving average.
type SlidingAverage struct {
	val atomic.Uint64
}

func NewSlidingAverage() *SlidingAverage {
	var x SlidingAverage
	return &x
}

func (sa *SlidingAverage) Add(sample int) {
	current := math.Float64frombits(sa.val.Load())
	updated := 0.95*current + 0.05*float64(sample)
	sa.val.Store(math.Float64bits(updated))
}

func (sa *SlidingAverage) Avg() float64 {
	return math.Float64frombits(sa.val.Load())
}
