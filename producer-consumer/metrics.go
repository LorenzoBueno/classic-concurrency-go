package main

import (
	"sync"
	"time"
)

type Metrics struct {
	mu                   sync.Mutex
	produced             int
	consumed             int
	perConsumer          []int
	timeouts             int
	occupancySum         int64
	occupancySampleCount int64
}

type Result struct {
	Produced         int
	Consumed         int
	PerConsumer      []int
	Timeouts         int
	Duration         time.Duration
	Throughput       float64
	AverageOccupancy float64
}

func newMetrics(consumers int) *Metrics {
	return &Metrics{perConsumer: make([]int, consumers)}
}

func (m *Metrics) recordProduced(occupancy int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.produced++
	m.recordOccupancy(occupancy)
}

func (m *Metrics) recordConsumed(consumerID, occupancy int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.consumed++
	m.perConsumer[consumerID]++
	m.recordOccupancy(occupancy)
}

func (m *Metrics) recordTimeout() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.timeouts++
}

func (m *Metrics) recordOccupancy(occupancy int) {
	m.occupancySum += int64(occupancy)
	m.occupancySampleCount++
}

func (m *Metrics) result(duration time.Duration) Result {
	m.mu.Lock()
	defer m.mu.Unlock()

	averageOccupancy := 0.0
	if m.occupancySampleCount > 0 {
		averageOccupancy = float64(m.occupancySum) / float64(m.occupancySampleCount)
	}

	throughput := 0.0
	if duration > 0 {
		throughput = float64(m.consumed) / duration.Seconds()
	}

	return Result{
		Produced:         m.produced,
		Consumed:         m.consumed,
		PerConsumer:      append([]int(nil), m.perConsumer...),
		Timeouts:         m.timeouts,
		Duration:         duration,
		Throughput:       throughput,
		AverageOccupancy: averageOccupancy,
	}
}
