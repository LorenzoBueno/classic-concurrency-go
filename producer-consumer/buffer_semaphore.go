package main

import (
	"sync"
	"sync/atomic"
	"time"
)

var endOfStream = Item{ProducerID: -1}

func (i Item) isEndOfStream() bool { return i.ProducerID < 0 }

type SemaphoreBuffer struct {
	slots    []Item
	in, out  int
	inMu     sync.Mutex
	outMu    sync.Mutex
	notFull  *Semaphore
	notEmpty *Semaphore
	count    atomic.Int64
}

func NewSemaphoreBuffer(capacity int) *SemaphoreBuffer {
	return &SemaphoreBuffer{
		slots:    make([]Item, capacity),
		notFull:  NewSemaphore(capacity, capacity),
		notEmpty: NewSemaphore(0, capacity),
	}
}

func (b *SemaphoreBuffer) Put(item Item) int {
	b.notFull.Acquire()

	b.inMu.Lock()
	b.slots[b.in] = item
	b.in = (b.in + 1) % len(b.slots)
	b.inMu.Unlock()

	occupancy := int(b.count.Add(1))
	b.notEmpty.Release()
	return occupancy
}

func (b *SemaphoreBuffer) take() (Item, int) {
	b.outMu.Lock()
	item := b.slots[b.out]
	b.slots[b.out] = Item{}
	b.out = (b.out + 1) % len(b.slots)
	b.outMu.Unlock()

	occupancy := int(b.count.Add(-1))
	b.notFull.Release()
	return item, occupancy
}

func (b *SemaphoreBuffer) Get() (Item, int) {
	b.notEmpty.Acquire()
	return b.take()
}

func (b *SemaphoreBuffer) GetTimeout(timeout time.Duration) (item Item, occupancy int, ok bool) {
	select {
	case <-b.notEmpty.Ready():
		item, occupancy = b.take()
		return item, occupancy, true
	case <-time.After(timeout):
		return Item{}, 0, false
	}
}

func produceSemaphore(id, amount int, buffer *SemaphoreBuffer, metrics *Metrics, wg *sync.WaitGroup) {
	defer wg.Done()

	for sequence := 0; sequence < amount; sequence++ {
		occupancy := buffer.Put(Item{ProducerID: id, Sequence: sequence})
		metrics.recordProduced(occupancy)
	}
}

func consumeSemaphore(
	id int,
	buffer *SemaphoreBuffer,
	timeout time.Duration,
	withTimeout bool,
	metrics *Metrics,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	for {
		var (
			item      Item
			occupancy int
		)
		if withTimeout {
			var ok bool
			item, occupancy, ok = buffer.GetTimeout(timeout)
			if !ok {
				metrics.recordTimeout()
				continue
			}
		} else {
			item, occupancy = buffer.Get()
		}

		if item.isEndOfStream() {
			return
		}
		metrics.recordConsumed(id, occupancy)
	}
}

func runSemaphore(config Config) Result {
	buffer := NewSemaphoreBuffer(config.BufferCapacity)
	metrics := newMetrics(config.Consumers)
	start := time.Now()

	var consumerWG sync.WaitGroup
	consumerWG.Add(config.Consumers)
	for id := 0; id < config.Consumers; id++ {
		go consumeSemaphore(id, buffer, config.Timeout, id == 0, metrics, &consumerWG)
	}

	var producerWG sync.WaitGroup
	producerWG.Add(config.Producers)
	for id := 0; id < config.Producers; id++ {
		go produceSemaphore(id, config.ItemsPerProducer, buffer, metrics, &producerWG)
	}

	producerWG.Wait()
	for i := 0; i < config.Consumers; i++ {
		buffer.Put(endOfStream)
	}
	consumerWG.Wait()

	return metrics.result(time.Since(start))
}