package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

type System struct {
	N          int
	R          int
	Forks      []chan struct{}
	Done       chan struct{}
	Progress   atomic.Int64
	WG         sync.WaitGroup
	Metrics    *Metrics
	NaiveReady chan struct{}
	NaiveStart chan struct{}
}

func newSystem(n, r int) *System {
	s := &System{
		N:       n,
		R:       r,
		Forks:   make([]chan struct{}, n),
		Done:    make(chan struct{}),
		Metrics: newMetrics(n),
	}
	for i := range s.Forks {
		s.Forks[i] = make(chan struct{}, 1)
		s.Forks[i] <- struct{}{}
	}
	return s
}

func (s *System) acquire(fork int) bool {
	select {
	case <-s.Forks[fork]:
		s.Progress.Add(1)
		return true
	case <-s.Done:
		return false
	}
}

func (s *System) release(fork int) {
	s.Forks[fork] <- struct{}{}
}

func think() { time.Sleep(time.Duration(1+rand.Intn(5)) * time.Millisecond) }
func eat()   { time.Sleep(time.Duration(1+rand.Intn(5)) * time.Millisecond) }

func main() {
	version := flag.String("version", "hierarchy", "naive, hierarchy ou limit")
	n := flag.Int("n", 5, "número de filósofos")
	r := flag.Int("r", 100, "número de iterações (refeições) por filósofo")
	timeout := flag.Duration("timeout", 2*time.Second, "tempo sem progresso para declarar deadlock")
	csvPath := flag.String("csv", "", "se informado, acrescenta os resultados a este arquivo CSV")
	rep := flag.Int("rep", 1, "número da repetição (gravado no CSV)")
	flag.Parse()

	if *n < 2 || *r < 1 {
		fmt.Fprintln(os.Stderr, "n deve ser >= 2 e r deve ser >= 1")
		os.Exit(2)
	}

	sys := newSystem(*n, *r)

	switch *version {
	case "naive":
		runNaive(sys)
	case "hierarchy":
		runHierarchy(sys)
	case "limit":
		runLimit(sys)
	default:
		fmt.Fprintf(os.Stderr, "versão %q inválida ou ainda não implementada\n", *version)
		os.Exit(2)
	}

	begin := time.Now()

	finished := make(chan struct{})
	go func() {
		sys.WG.Wait()
		close(finished)
	}()

	deadlock := false
	last := sys.Progress.Load()
	lastChange := time.Now()
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

loop:
	for {
		select {
		case <-finished:
			break loop
		case <-ticker.C:
			cur := sys.Progress.Load()
			if cur != last {
				last, lastChange = cur, time.Now()
			} else if time.Since(lastChange) > *timeout {
				deadlock = true
				close(sys.Done)
				<-finished
				break loop
			}
		}
	}

	elapsed := time.Since(begin)
	sys.Metrics.report(*version, *n, *r, elapsed, deadlock)

	if *csvPath != "" {
		if err := sys.Metrics.writeCSV(*csvPath, *version, *n, *r, *rep, elapsed, deadlock); err != nil {
			fmt.Fprintf(os.Stderr, "erro ao gravar CSV: %v\n", err)
			os.Exit(1)
		}
	}
}
