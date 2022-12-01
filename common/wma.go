package common

// MovingAverage implements a simplistic non-threadsafe windowed moving average.
// The implementation will give erroneous values until the window is filled.
type MovingAverage struct {
	values []int
	idx    int
}

func NewMovingAverage(windowSize int) *MovingAverage {
	return &MovingAverage{
		values: make([]int, windowSize),
	}
}

func (ma *MovingAverage) Add(value int) {
	ma.values[ma.idx] = value
	ma.idx++
	ma.idx %= len(ma.values)
}

func (ma *MovingAverage) Avg() float64 {
	var sum int
	for _, v := range ma.values {
		sum += v
	}
	return float64(sum) / float64(len(ma.values))
}

func (ma *MovingAverage) Max() int {
	max := ma.values[0]
	for _, v := range ma.values {
		if v > max {
			max = v
		}
	}
	return max
}
func (ma *MovingAverage) Min() int {
	min := ma.values[0]
	for _, v := range ma.values {
		if v < min {
			min = v
		}
	}
	return min
}

type SlidingAverage float64

func NewSlidingAverage() *SlidingAverage {
	var x SlidingAverage
	return &x
}

func (sa *SlidingAverage) Add(value int) {
	*sa = SlidingAverage(0.95*float64(*sa) + 0.05*float64(value))
}

func (sa *SlidingAverage) Avg() float64 {
	return float64(*sa)
}
