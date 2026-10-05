package main

import (
	"fmt"
	"sync"
	"time"
)

// Metrics collects, per philosopher, the number of meals and the time spent
// waiting for forks (from getting hungry until holding both forks).
type Metrics struct {
	mu        sync.Mutex
	meals     []int64
	totalWait []time.Duration
	maxWait   []time.Duration
}

func NewMetrics(n int) *Metrics {
	return &Metrics{
		meals:     make([]int64, n),
		totalWait: make([]time.Duration, n),
		maxWait:   make([]time.Duration, n),
	}
}

// Record is called when philosopher id gets both forks and starts eating.
func (m *Metrics) Record(id int, wait time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.meals[id]++
	m.totalWait[id] += wait
	if wait > m.maxWait[id] {
		m.maxWait[id] = wait
	}
}

func (m *Metrics) TotalMeals() int64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	var total int64
	for _, v := range m.meals {
		total += v
	}
	return total
}

func (m *Metrics) Print() {
	m.mu.Lock()
	defer m.mu.Unlock()
	fmt.Println("philosopher | meals | avg wait    | max wait")
	for i := range m.meals {
		var avg time.Duration
		if m.meals[i] > 0 {
			avg = m.totalWait[i] / time.Duration(m.meals[i])
		}
		fmt.Printf("%11d | %5d | %-11v | %v\n",
			i, m.meals[i], avg.Round(time.Microsecond), m.maxWait[i].Round(time.Microsecond))
	}
}