package evms

import (
	"sync/atomic"
	"time"

	"github.com/holiman/goevmlab/utils"
)

const (
	slowTestWarmupRuns = 500
)

type VMStat struct {
	// Some metrics
	tracingSpeedWMA     utils.SlidingAverage
	lastTracingTime     atomic.Uint64
	shortestTracingTime atomic.Uint64
	longestTracingTime  atomic.Uint64
	numExecs            atomic.Uint64
}

// TraceDone marks the tracing speed metric, and returns 'true' if the test is
// 'slow'.
func (stat *VMStat) TraceDone(start time.Time) (time.Duration, bool) {
	numexecs := stat.numExecs.Add(1)
	duration := time.Since(start)
	stat.tracingSpeedWMA.Add(int(duration))

	durationNs := uint64(duration)
	stat.lastTracingTime.Store(durationNs)
	updateMinDuration(&stat.shortestTracingTime, durationNs)
	if updateMaxDuration(&stat.longestTracingTime, durationNs) {
		// Don't count the first runs, let the VM warm up.
		if numexecs > slowTestWarmupRuns {
			return duration, true
		}
	}
	return duration, false
}

func (stat *VMStat) Stats() []any {
	return []any{
		"execSpeed", time.Duration(stat.tracingSpeedWMA.Avg()).Round(100 * time.Microsecond),
		"last", time.Duration(stat.lastTracingTime.Load()).Round(100 * time.Microsecond),
		"shortest", time.Duration(stat.shortestTracingTime.Load()).Round(100 * time.Microsecond),
		"longest", time.Duration(stat.longestTracingTime.Load()),
		"count", stat.numExecs.Load(),
	}
}

func updateMinDuration(field *atomic.Uint64, duration uint64) {
	for {
		current := field.Load()
		if current != 0 && current <= duration {
			return
		}
		if field.CompareAndSwap(current, duration) {
			return
		}
	}
}

func updateMaxDuration(field *atomic.Uint64, duration uint64) bool {
	for {
		current := field.Load()
		if current >= duration {
			return false
		}
		if field.CompareAndSwap(current, duration) {
			return true
		}
	}
}

type tracingResult struct {
	Slow     bool
	ExecTime time.Duration
	Cmd      string
}
