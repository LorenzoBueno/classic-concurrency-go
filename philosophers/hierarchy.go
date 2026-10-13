package main

import "time"

func runHierarchy(s *System) {
	for id := 0; id < s.N; id++ {
		s.WG.Add(1)
		go hierarchyPhilosopher(id, s)
	}
}

func hierarchyPhilosopher(id int, s *System) {
	defer s.WG.Done()

	first := id
	second := (id + 1) % s.N
	if second < first {
		first, second = second, first
	}

	for i := 0; i < s.R; i++ {
		think()

		start := time.Now()
		if !s.acquire(first) {
			return
		}
		if !s.acquire(second) {
			return
		}
		s.Metrics.recordMeal(id, time.Since(start))

		eat()

		s.release(second)
		s.release(first)
	}
}