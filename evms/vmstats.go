package evms

import (
	"time"

	"github.com/holiman/goevmlab/utils"
)

type VmStat struct {
	// Some metrics
	tracingSpeedWMA    utils.SlidingAverage
	longestTracingTime time.Duration
	numExecs           int
}

// TraceDone marks the tracing speed metric, and returns 'true' if the test is
// 'slow'.
func (stat *VmStat) TraceDone(start time.Time) (time.Duration, bool) {
	stat.numExecs++
	duration := time.Since(start)
	stat.tracingSpeedWMA.Add(int(duration))
	if duration > stat.longestTracingTime {
		stat.longestTracingTime = duration
		// Don't count the first 500 runs, let it accumulate.
		if stat.numExecs > 500 {
			return duration, true
		}
	}
	return duration, false
}

type tracingResult struct {
	Slow     bool
	ExecTime time.Duration
	Cmd      string
}
