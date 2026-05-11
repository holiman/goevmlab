package evms

import (
	"sync"
	"time"

	"github.com/holiman/goevmlab/utils"
)

const (
	// Give each VM enough runs to build a useful baseline before preserving
	// generated tests as slow-test artefacts.
	slowTestWarmupRuns  = 500
	slowTestMinDuration = 250 * time.Millisecond
	slowTestMultiplier  = 5
)

type VMStat struct {
	// Batch VMs can share one VMStat across per-thread instances.
	lock sync.Mutex

	// Some metrics
	tracingSpeedWMA     utils.SlidingAverage
	lastTracingTime     time.Duration
	shortestTracingTime time.Duration
	longestTracingTime  time.Duration
	numExecs            uint64
}

// TraceDone marks the tracing speed metric, and returns 'true' if the test is
// 'slow'.
func (stat *VMStat) TraceDone(start time.Time) (time.Duration, bool) {
	duration := time.Since(start)
	stat.lock.Lock()
	defer stat.lock.Unlock()

	stat.numExecs++
	numexecs := stat.numExecs
	// Compare against the previous average. Including this run first would
	// let a large outlier raise its own slow-test threshold.
	avg := time.Duration(stat.tracingSpeedWMA.Avg())

	stat.lastTracingTime = duration
	if stat.shortestTracingTime == 0 || duration < stat.shortestTracingTime {
		stat.shortestTracingTime = duration
	}
	if duration > stat.longestTracingTime {
		stat.longestTracingTime = duration
	}
	stat.tracingSpeedWMA.Add(int(duration))

	// Don't count the first runs, let the moving average accumulate.
	if numexecs <= slowTestWarmupRuns || avg == 0 {
		return duration, false
	}
	threshold := max(avg * slowTestMultiplier, slowTestMinDuration)
	return duration, duration > threshold
}

func (stat *VMStat) Stats() []any {
	stat.lock.Lock()
	defer stat.lock.Unlock()

	threshold := max(time.Duration(stat.tracingSpeedWMA.Avg()) * slowTestMultiplier, slowTestMinDuration)
	return []any{
		"execSpeed", time.Duration(stat.tracingSpeedWMA.Avg()).Round(100 * time.Microsecond),
		"last", stat.lastTracingTime.Round(100 * time.Microsecond),
		"shortest", stat.shortestTracingTime.Round(100 * time.Microsecond),
		"longest", stat.longestTracingTime,
		"slowLimit", threshold.Round(100 * time.Microsecond),
		"count", stat.numExecs,
	}
}

type tracingResult struct {
	Slow     bool
	ExecTime time.Duration
	Cmd      string
}
