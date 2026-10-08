package main

import (
	"sync"
	"time"
)

func consume(
	id int,
	buffer <-chan Item,
	timeout time.Duration,
	withTimeout bool,
	metrics *Metrics,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	if !withTimeout {
		for range buffer {
			metrics.recordConsumed(id, len(buffer))
		}
		return
	}

	for {
		select {
		case _, ok := <-buffer:
			if !ok {
				return
			}
			metrics.recordConsumed(id, len(buffer))
		case <-time.After(timeout):
			metrics.recordTimeout()
		}
	}
}
