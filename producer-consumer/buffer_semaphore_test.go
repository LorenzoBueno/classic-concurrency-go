package main

import (
	"sync"
	"testing"
	"time"
)

func TestRunSemaphoreConsumesEveryProducedItem(t *testing.T) {
	for _, k := range []int{1, 10, 100} {
		config := Config{
			Producers:        4,
			Consumers:        3,
			BufferCapacity:   k,
			ItemsPerProducer: 250,
			Timeout:          time.Millisecond,
		}

		result := runSemaphore(config)
		want := config.Producers * config.ItemsPerProducer
		if result.Produced != want || result.Consumed != want {
			t.Fatalf("k=%d: produzidos=%d consumidos=%d; esperado %d", k, result.Produced, result.Consumed, want)
		}
		sum := 0
		for _, c := range result.PerConsumer {
			sum += c
		}
		if sum != want {
			t.Fatalf("k=%d: soma por consumidor = %d; esperado %d", k, sum, want)
		}
	}
}

func TestSemaphoreBufferExactlyOnce(t *testing.T) {
	const producers, consumers, perProducer, capacity = 5, 5, 400, 3
	buffer := NewSemaphoreBuffer(capacity)

	var seen [producers][perProducer]int
	var seenMu sync.Mutex

	var cwg sync.WaitGroup
	cwg.Add(consumers)
	for c := 0; c < consumers; c++ {
		go func() {
			defer cwg.Done()
			for {
				item, _ := buffer.Get()
				if item.isEndOfStream() {
					return
				}
				seenMu.Lock()
				seen[item.ProducerID][item.Sequence]++
				seenMu.Unlock()
			}
		}()
	}

	var pwg sync.WaitGroup
	pwg.Add(producers)
	for p := 0; p < producers; p++ {
		go func(id int) {
			defer pwg.Done()
			for s := 0; s < perProducer; s++ {
				buffer.Put(Item{ProducerID: id, Sequence: s})
			}
		}(p)
	}
	pwg.Wait()
	for c := 0; c < consumers; c++ {
		buffer.Put(endOfStream)
	}
	cwg.Wait()

	for p := range seen {
		for s, n := range seen[p] {
			if n != 1 {
				t.Fatalf("item (%d,%d) retirado %d vezes; esperado 1", p, s, n)
			}
		}
	}
}

func TestSemaphoreConsumerRecordsTimeoutAndThenStops(t *testing.T) {
	buffer := NewSemaphoreBuffer(2)
	metrics := newMetrics(1)
	var wg sync.WaitGroup
	wg.Add(1)
	go consumeSemaphore(0, buffer, 500*time.Microsecond, true, metrics, &wg)

	time.Sleep(3 * time.Millisecond)
	buffer.Put(endOfStream)
	wg.Wait()

	if metrics.result(time.Millisecond).Timeouts == 0 {
		t.Fatal("consumidor deveria registrar pelo menos um timeout")
	}
}