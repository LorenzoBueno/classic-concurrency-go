package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"time"
)

type PhilStats struct {
	Meals     int
	TotalWait time.Duration
	MaxWait   time.Duration
}

func (s PhilStats) avgWait() time.Duration {
	if s.Meals == 0 {
		return 0
	}
	return s.TotalWait / time.Duration(s.Meals)
}

type Metrics struct {
	stats []PhilStats
}

func newMetrics(n int) *Metrics {
	return &Metrics{stats: make([]PhilStats, n)}
}

func (m *Metrics) recordMeal(id int, wait time.Duration) {
	s := &m.stats[id]
	s.Meals++
	s.TotalWait += wait
	if wait > s.MaxWait {
		s.MaxWait = wait
	}
}

func jain(xs []float64) float64 {
	var sum, sumSq float64
	for _, x := range xs {
		sum += x
		sumSq += x * x
	}
	if sumSq == 0 {
		return 1
	}
	return (sum * sum) / (float64(len(xs)) * sumSq)
}

func (m *Metrics) avgWaitsAsFloat() []float64 {
	out := make([]float64, 0, len(m.stats))
	for _, s := range m.stats {
		out = append(out, float64(s.avgWait()))
	}
	return out
}

func ms(d time.Duration) float64 {
	return float64(d) / float64(time.Millisecond)
}

func (m *Metrics) report(version string, n, r int, elapsed time.Duration, deadlock bool) {
	fmt.Printf("\n=== version=%s N=%d R=%d | tempo total: %v ===\n",
		version, n, r, elapsed.Round(time.Millisecond))
	if deadlock {
		fmt.Println(">>> DEADLOCK DETECTADO: nenhum progresso; execução encerrada de forma ordenada.")
		fmt.Println(">>> Índices de fairness não mensuráveis: a execução não completou as iterações previstas.")
	} else {
		fmt.Println(">>> Execução concluída sem deadlock.")
	}

	fmt.Printf("%-10s %-10s %-16s %-16s\n", "Filósofo", "Refeições", "Espera média", "Espera máxima")
	meals := make([]float64, 0, n)
	for i, s := range m.stats {
		meals = append(meals, float64(s.Meals))
		fmt.Printf("%-10d %-10d %-16v %-16v\n", i, s.Meals,
			s.avgWait().Round(time.Microsecond), s.MaxWait.Round(time.Microsecond))
	}
	if deadlock {
		fmt.Println("\nÍndice de Jain (refeições):    n/a")
		fmt.Println("Índice de Jain (espera média): n/a")
	} else {
		fmt.Printf("\nÍndice de Jain (refeições):    %.4f\n", jain(meals))
		fmt.Printf("Índice de Jain (espera média): %.4f\n", jain(m.avgWaitsAsFloat()))
	}
	fmt.Println("(1.0 = perfeitamente justo; n/a = fairness não mensurável)")
}

func (m *Metrics) writeCSV(path, version string, n, r, rep int, elapsed time.Duration, deadlock bool) error {
	info, err := os.Stat(path)
	needHeader := err != nil || info.Size() == 0

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	if needHeader {
		header := []string{
			"version", "n", "r", "repetition", "deadlock", "duration_ms",
			"philosopher", "meals", "avg_wait_ms", "max_wait_ms", "jain_wait",
		}
		if err := w.Write(header); err != nil {
			return err
		}
	}

	jw := strconv.FormatFloat(jain(m.avgWaitsAsFloat()), 'f', 4, 64)
	if deadlock {
		jw = "n/a"
	}
	for i, s := range m.stats {
		avgWait := "n/a"
		maxWait := "n/a"
		if s.Meals > 0 {
			avgWait = strconv.FormatFloat(ms(s.avgWait()), 'f', 3, 64)
			maxWait = strconv.FormatFloat(ms(s.MaxWait), 'f', 3, 64)
		}
		row := []string{
			version,
			strconv.Itoa(n),
			strconv.Itoa(r),
			strconv.Itoa(rep),
			strconv.FormatBool(deadlock),
			strconv.FormatFloat(ms(elapsed), 'f', 3, 64),
			strconv.Itoa(i),
			strconv.Itoa(s.Meals),
			avgWait,
			maxWait,
			jw,
		}
		if err := w.Write(row); err != nil {
			return err
		}
	}

	w.Flush()
	return w.Error()
}
