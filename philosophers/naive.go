package main

import "time"

func runNaive(s *System) {
	s.NaiveReady = make(chan struct{}, s.N)
	s.NaiveStart = make(chan struct{})
	s.WG.Add(1)
	go func() {
		defer s.WG.Done()
		for i := 0; i < s.N; i++ {
			select {
			case <-s.NaiveReady:
			case <-s.Done:
				close(s.NaiveStart)
				return
			}
		}
		close(s.NaiveStart)
	}()

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
		waited := time.Since(start)
		s.NaiveReady <- struct{}{}
		select {
		case <-s.NaiveStart:
		case <-s.Done:
			s.release(left)
			return
		}

		start = time.Now()
		if !s.acquire(right) {
			s.release(left)
			return
		}
		waited += time.Since(start)
		s.Metrics.recordMeal(id, waited)

		eat()

		s.release(right)
		s.release(left)
	}
}
