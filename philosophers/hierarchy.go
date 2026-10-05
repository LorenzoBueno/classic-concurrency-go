package main

import (
	"sync"
	"sync/atomic"
	"time"
)

// runHierarchy breaks the circular wait condition: forks are numbered and
// every philosopher picks the LOWER-numbered fork first. Only the last
// philosopher (N-1) behaves differently from the naive version: its forks are
// N-1 and 0, so it picks fork 0 first.
func runHierarchy(n int, r int64, m *Metrics) {
	forks := newForks(n)
	remaining := r // meals still to be claimed (shared, atomic)

	var wg sync.WaitGroup
	for id := 0; id < n; id++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			first, second := id, (id+1)%n
			if first > second {
				first, second = second, first
			}

			for atomic.AddInt64(&remaining, -1) >= 0 {
				time.Sleep(randDuration(maxThink)) // thinking

				hungry := time.Now()
				<-forks[first]
				time.Sleep(pauseBetweenForks) // same pause as naive, for a fair comparison
				<-forks[second]
				m.Record(id, time.Since(hungry))

				time.Sleep(randDuration(maxEat)) // eating

				forks[second] <- struct{}{}
				forks[first] <- struct{}{}
			}
		}(id)
	}
	wg.Wait()
}