package evms

import (
	"time"

	"github.com/holiman/goevmlab/utils"
	"sync/atomic"
)

type VMStat struct {
	// Some metrics
	tracingSpeedWMA    utils.SlidingAverage
	longestTracingTime time.Duration
	numExecs           atomic.Uint64
}

// TraceDone marks the tracing speed metric, and returns 'true' if the test is
// 'slow'.
func (stat *VMStat) TraceDone(start time.Time) (time.Duration, bool) {
	numexecs := stat.numExecs.Add(1)
	duration := time.Since(start)
	stat.tracingSpeedWMA.Add(int(duration))
	if duration > stat.longestTracingTime {
		stat.longestTracingTime = duration
		// Don't count the first 500 runs, let it accumulate.
		if numexecs > 500 {
			return duration, true
		}
	}
	return duration, false
}

func (stat *VMStat) Stats() []any {
	return []interface{}{
		"execSpeed", time.Duration(stat.tracingSpeedWMA.Avg()).Round(100 * time.Microsecond),
		"longest", stat.longestTracingTime,
		"count", stat.numExecs.Load(),
	}
}

type tracingResult struct {
	Slow     bool
	ExecTime time.Duration
	Cmd      string
}
