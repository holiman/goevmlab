package evms

import (
	"sync"
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
}

func TestVMStatTracksConcurrentDurations(t *testing.T) {
	stat := new(VMStat)
	const workers = 8
	const runs = 100

	var wg sync.WaitGroup
	for worker := range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for run := range runs {
				markTraceDone(stat, time.Duration(worker+run+1)*time.Millisecond)
			}
		}()
	}
	wg.Wait()

	fields := statFields(stat)
	if have, want := fields["count"], uint64(workers*runs); have != want {
		t.Fatalf("wrong execution count: have %v, want %v", have, want)
	}
	shortest := fields["shortest"].(time.Duration)
	longest := fields["longest"].(time.Duration)
	if shortest <= 0 {
		t.Fatalf("expected non-zero shortest duration, got %v", shortest)
	}
	if longest < shortest {
		t.Fatalf("longest duration %v below shortest duration %v", longest, shortest)
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

func TestVMStatSlowDetectionTracksNewLongestAfterWarmup(t *testing.T) {
	stat := new(VMStat)

	for range slowTestWarmupRuns {
		markTraceDone(stat, 100*time.Millisecond)
	}
	if _, slow := markTraceDone(stat, 50*time.Millisecond); slow {
		t.Fatalf("duration below the previous longest was marked slow")
	}
	if _, slow := markTraceDone(stat, 200*time.Millisecond); !slow {
		t.Fatalf("new longest duration after warmup was not marked slow")
	}
}
