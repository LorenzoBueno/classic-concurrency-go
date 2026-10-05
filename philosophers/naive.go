package main

import (
	"sync"
	"sync/atomic"
	"time"
)

// runNaive is the base version: every philosopher picks the LEFT fork first
// and then the right one. All four Coffman conditions hold, so it deadlocks.
// To see the deadlock, use R >= N (every philosopher must get a meal slot).
func runNaive(n int, r int64, m *Metrics) {
	forks := newForks(n)
	remaining := r // meals still to be claimed (shared, atomic)

	var wg sync.WaitGroup
	for id := 0; id < n; id++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			left, right := id, (id+1)%n

			for atomic.AddInt64(&remaining, -1) >= 0 {
				time.Sleep(randDuration(maxThink)) // thinking

				hungry := time.Now()
				<-forks[left]
				time.Sleep(pauseBetweenForks)
				<-forks[right]
				m.Record(id, time.Since(hungry))

				time.Sleep(randDuration(maxEat)) // eating

				forks[right] <- struct{}{}
				forks[left] <- struct{}{}
			}
		}(id)
	}
	wg.Wait()
}