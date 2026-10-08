package main

type Semaphore struct {
	permits chan struct{}
}

func NewSemaphore(initial, max int) *Semaphore {
	s := &Semaphore{permits: make(chan struct{}, max)}
	for i := 0; i < initial; i++ {
		s.permits <- struct{}{}
	}
	return s
}

func (s *Semaphore) Acquire() {
	<-s.permits
}

func (s *Semaphore) Release() {
	s.permits <- struct{}{}
}

func (s *Semaphore) Ready() <-chan struct{} {
	return s.permits
}