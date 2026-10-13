package main

import (
	"sync"
	"testing"
	"time"
)

func TestRunChannelConsumesEveryProducedItem(t *testing.T) {
	config := Config{
		Producers:        4,
		Consumers:        3,
		BufferCapacity:   10,
		ItemsPerProducer: 250,
		Timeout:          time.Millisecond,
	}

	result := runChannel(config)
	want := config.Producers * config.ItemsPerProducer
	if result.Produced != want {
		t.Fatalf("produzidos = %d; esperado %d", result.Produced, want)
	}
	if result.Consumed != want {
		t.Fatalf("consumidos = %d; esperado %d", result.Consumed, want)
	}

	consumedByWorkers := 0
	for _, count := range result.PerConsumer {
		consumedByWorkers += count
	}
	if consumedByWorkers != want {
		t.Fatalf("soma por consumidor = %d; esperado %d", consumedByWorkers, want)
	}
}

func TestConfigValidation(t *testing.T) {
	valid := Config{Producers: 1, Consumers: 1, BufferCapacity: 1, ItemsPerProducer: 1, Timeout: time.Millisecond}
	if err := valid.validate(); err != nil {
		t.Fatalf("configuracao valida retornou erro: %v", err)
	}

	invalid := valid
	invalid.BufferCapacity = 0
	if err := invalid.validate(); err == nil {
		t.Fatal("configuracao com buffer zero deveria falhar")
	}
}

func TestConsumerRecordsTimeoutAndThenStops(t *testing.T) {
	buffer := make(chan Item)
	metrics := newMetrics(1)
	var wg sync.WaitGroup
	wg.Add(1)
	go consume(0, buffer, 500*time.Microsecond, true, metrics, &wg)

	time.Sleep(3 * time.Millisecond)
	close(buffer)
	wg.Wait()

	result := metrics.result(time.Millisecond)
	if result.Timeouts == 0 {
		t.Fatal("consumidor deveria registrar pelo menos um timeout")
	}
}
