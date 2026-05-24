package evms

import (
	"time"

	"github.com/holiman/goevmlab/utils"
	"sync/atomic"
)

type VMStat struct {
	// Some metrics
	tracingSpeedWMA    utils.SlidingAverage
	longestTracingTime atomic.Int64
	numExecs           atomic.Uint64
}

// TraceDone marks the tracing speed metric, and returns 'true' if the test is
// 'slow'.
func (stat *VMStat) TraceDone(start time.Time) (time.Duration, bool) {
	numexecs := stat.numExecs.Add(1)
	duration := time.Since(start)
	stat.tracingSpeedWMA.Add(int(duration))
	// This is not strictly correct - two racing updates may interleave and
	// one may overwrite the other. That's fine, it's just a bit of stats.
	if duration > time.Duration(stat.longestTracingTime.Load()) {
		stat.longestTracingTime.Store(int64(duration))
		// Don't count the first 500 runs, let it accumulate.
		if numexecs > 500 {
			return duration, true
		}
	}
	return duration, false
}

func (stat *VMStat) Stats() []any {
	return []any{
		"execSpeed", time.Duration(stat.tracingSpeedWMA.Avg()).Round(100 * time.Microsecond),
		"longest", time.Duration(stat.longestTracingTime.Load()),
		"count", stat.numExecs.Load(),
	}
}

type tracingResult struct {
	Slow     bool
	ExecTime time.Duration
	Cmd      string
}
