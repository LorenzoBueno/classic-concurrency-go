package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"time"
)

const (
	maxThink = 10 * time.Millisecond
	maxEat   = 10 * time.Millisecond

	// Pause between picking the first and the second fork. It is the same for
	// every strategy, so the comparison is fair. In the naive version it makes
	// the deadlock deterministic: it is longer than maxThink, so every
	// philosopher already holds the left fork before anyone asks for the right.
	pauseBetweenForks = 50 * time.Millisecond
)

// strategy is the signature shared by every version of the problem.
// It must block until all philosophers have finished.
type strategy func(n int, r int64, m *Metrics)

func main() {
	version := flag.String("version", "naive", "naive | hierarchy")
	n := flag.Int("n", 5, "number of philosophers (N >= 2)")
	r := flag.Int64("r", 100, "total number of meals to serve (R)")
	flag.Parse()

	if *n < 2 || *r < 1 {
		fmt.Fprintln(os.Stderr, "n must be >= 2 and r must be >= 1")
		os.Exit(2)
	}

	strategies := map[string]strategy{
		"naive":     runNaive,
		"hierarchy": runHierarchy,
	}
	run, ok := strategies[*version]
	if !ok {
		fmt.Fprintf(os.Stderr, "unknown version %q\n", *version)
		os.Exit(2)
	}

	m := NewMetrics(*n)
	done := make(chan struct{})
	start := time.Now()
	go func() {
		run(*n, *r, m)
		close(done)
	}()

	// Watchdog: if no meal is served during a full tick, the table is stuck.
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	last := int64(-1)
	for {
		select {
		case <-done:
			fmt.Printf("OK: version=%s n=%d r=%d finished in %v\n",
				*version, *n, *r, time.Since(start).Round(time.Millisecond))
			m.Print()
			return
		case <-ticker.C:
			cur := m.TotalMeals()
			if cur == last {
				fmt.Printf("DEADLOCK: version=%s n=%d r=%d, no progress for 1s (%d meals served)\n",
					*version, *n, *r, cur)
				m.Print()
				os.Exit(1)
			}
			last = cur
		}
	}
}

// newForks creates n forks. A fork is a channel with capacity 1:
// holding a token in the channel means the fork is on the table.
// Picking up = receive, putting down = send.
func newForks(n int) []chan struct{} {
	forks := make([]chan struct{}, n)
	for i := range forks {
		forks[i] = make(chan struct{}, 1)
		forks[i] <- struct{}{}
	}
	return forks
}

// randDuration returns a random duration in [0, upTo).
func randDuration(upTo time.Duration) time.Duration {
	return time.Duration(rand.Int63n(int64(upTo)))
}