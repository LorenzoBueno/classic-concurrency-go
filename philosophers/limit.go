package main

import "time"

func runLimit(s *System) {
	seats := make(chan struct{}, s.N-1)
	for id := 0; id < s.N; id++ {
		s.WG.Add(1)
		go limitPhilosopher(id, s, seats)
	}
}

func limitPhilosopher(id int, s *System, seats chan struct{}) {
	defer s.WG.Done()

	for i := 0; i < s.R; i++ {
		think()

		select {
		case seats <- struct{}{}:
		case <-s.Done:
			return
		}
		if !limitMeal(id, s) {
			<-seats
			return
		}
		<-seats
	}
}

func limitMeal(id int, s *System) bool {
	left := id
	right := (id + 1) % s.N

	start := time.Now()
	if !s.acquire(left) {
		return false
	}
	waited := time.Since(start)
	defer s.release(left)

	start = time.Now()
	if !s.acquire(right) {
		return false
	}
	waited += time.Since(start)
	defer s.release(right)

	s.Metrics.recordMeal(id, waited)
	eat()
	return true
}
