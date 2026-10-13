package main

import (
	"sync"
	"testing"
	"time"
)

func TestLimitCompletesAllMeals(t *testing.T) {
	const n, r = 5, 20
	system := newSystem(n, r)
	runLimit(system)

	finished := make(chan struct{})
	go func() {
		system.WG.Wait()
		close(finished)
	}()

	select {
	case <-finished:
	case <-time.After(5 * time.Second):
		close(system.Done)
		select {
		case <-finished:
		case <-time.After(time.Second):
		}
		t.Fatal("limit strategy did not finish")
	}

	for id, stats := range system.Metrics.stats {
		if stats.Meals != r {
			t.Errorf("philosopher %d ate %d meals; want %d", id, stats.Meals, r)
		}
	}
}

func TestNaiveDeadlocksAndStopsCleanly(t *testing.T) {
	const n, r = 5, 10
	system := newSystem(n, r)
	runNaive(system)

	var stopOnce sync.Once
	stop := func() {
		stopOnce.Do(func() { close(system.Done) })
	}
	finished := make(chan struct{})
	go func() {
		system.WG.Wait()
		close(finished)
	}()
	t.Cleanup(func() {
		stop()
		select {
		case <-finished:
		case <-time.After(time.Second):
			t.Error("naive strategy left goroutines blocked after cancellation")
		}
	})

	deadline := time.After(time.Second)
	ticker := time.NewTicker(time.Millisecond)
	defer ticker.Stop()
	for system.Progress.Load() < int64(n) {
		select {
		case <-ticker.C:
		case <-deadline:
			t.Fatal("all philosophers did not acquire their left fork")
		case <-finished:
			t.Fatal("naive strategy completed instead of deadlocking")
		}
	}

	select {
	case <-finished:
		t.Fatal("naive strategy completed instead of deadlocking")
	default:
	}
	stop()
	select {
	case <-finished:
	case <-time.After(time.Second):
		t.Fatal("naive strategy left goroutines blocked after cancellation")
	}
}
