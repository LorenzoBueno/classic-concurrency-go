package main

import "time"

func runNaive(s *System) {
	for id := 0; id < s.N; id++ {
		s.WG.Add(1)
		go naivePhilosopher(id, s)
	}
}

func naivePhilosopher(id int, s *System) {
	defer s.WG.Done()

	left := id
	right := (id + 1) % s.N

	for i := 0; i < s.R; i++ {
		think()

		start := time.Now()
		if !s.acquire(left) {
			return
		}
		time.Sleep(20 * time.Millisecond)
		if !s.acquire(right) {
			return
		}
		s.Metrics.recordMeal(id, time.Since(start))

		eat()

		s.release(right)
		s.release(left)
	}
}