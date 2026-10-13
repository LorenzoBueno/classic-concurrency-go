package main

import "sync"

func produce(id, amount int, buffer chan<- Item, metrics *Metrics, wg *sync.WaitGroup) {
	defer wg.Done()

	for sequence := 0; sequence < amount; sequence++ {
		buffer <- Item{ProducerID: id, Sequence: sequence}
		metrics.recordProduced(len(buffer))
	}
}
