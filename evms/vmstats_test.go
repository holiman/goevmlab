package evms

import (
	"testing"
	"time"
)

func markTraceDone(stat *VMStat, duration time.Duration) (time.Duration, bool) {
	return stat.TraceDone(time.Now().Add(-duration))
}

func statFields(stat *VMStat) map[string]any {
	fields := make(map[string]any)
	stats := stat.Stats()
	for i := 0; i+1 < len(stats); i += 2 {
		fields[stats[i].(string)] = stats[i+1]
	}
	return fields
}

func TestVMStatTracksDurations(t *testing.T) {
	stat := new(VMStat)

	markTraceDone(stat, 20*time.Millisecond)
	markTraceDone(stat, 10*time.Millisecond)

	fields := statFields(stat)
	if have, want := fields["count"], uint64(2); have != want {
		t.Fatalf("wrong execution count: have %v, want %v", have, want)
	}
	if fields["execSpeed"].(time.Duration) <= 0 {
		t.Fatalf("expected non-zero execution speed, got %v", fields["execSpeed"])
	}
	if fields["last"].(time.Duration) <= 0 {
		t.Fatalf("expected non-zero last duration, got %v", fields["last"])
	}
	shortest := fields["shortest"].(time.Duration)
	longest := fields["longest"].(time.Duration)
	if shortest <= 0 {
		t.Fatalf("expected non-zero shortest duration, got %v", shortest)
	}
	if longest < shortest {
		t.Fatalf("longest duration %v below shortest duration %v", longest, shortest)
	}
	if fields["slowLimit"].(time.Duration) < slowTestMinDuration {
		t.Fatalf("slow limit below minimum: %v", fields["slowLimit"])
	}
}

func TestVMStatSlowDetectionNeedsWarmup(t *testing.T) {
	stat := new(VMStat)

	for i := range slowTestWarmupRuns {
		if _, slow := markTraceDone(stat, 2*time.Second); slow {
			t.Fatalf("run %d was marked slow during warmup", i)
		}
	}
}

func TestVMStatSlowDetectionUsesMovingAverage(t *testing.T) {
	stat := new(VMStat)

	for range slowTestWarmupRuns {
		markTraceDone(stat, 10*time.Millisecond)
	}
	if _, slow := markTraceDone(stat, 200*time.Millisecond); slow {
		t.Fatalf("duration below the minimum slow threshold was marked slow")
	}
	if _, slow := markTraceDone(stat, 2*time.Second); !slow {
		t.Fatalf("large outlier was not marked slow")
	}
}
