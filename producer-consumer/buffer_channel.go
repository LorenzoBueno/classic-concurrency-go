package main

import (
	"fmt"
	"sync"
	"time"
)

type Item struct {
	ProducerID int
	Sequence   int
}

func runChannel(config Config) Result {
	buffer := make(chan Item, config.BufferCapacity)
	metrics := newMetrics(config.Consumers)
	start := time.Now()

	var consumerWG sync.WaitGroup
	consumerWG.Add(config.Consumers)
	for id := 0; id < config.Consumers; id++ {
		go consume(id, buffer, config.Timeout, id == 0, metrics, &consumerWG)
	}

	var producerWG sync.WaitGroup
	producerWG.Add(config.Producers)
	for id := 0; id < config.Producers; id++ {
		go produce(id, config.ItemsPerProducer, buffer, metrics, &producerWG)
	}

	producerWG.Wait()
	close(buffer)
	consumerWG.Wait()

	return metrics.result(time.Since(start))
}

func printResult(result Result, config Config, implementation string) {
	fmt.Println("Implementacao:", implementation)
	fmt.Printf("Produtores: %d | Consumidores: %d | Capacidade: %d\n", config.Producers, config.Consumers, config.BufferCapacity)
	fmt.Printf("Produzidos: %d\n", result.Produced)
	fmt.Printf("Consumidos: %d\n", result.Consumed)
	fmt.Printf("Itens perdidos: %d\n", result.Produced-result.Consumed)
	fmt.Printf("Duracao: %s\n", result.Duration)
	fmt.Printf("Throughput: %.2f itens/s\n", result.Throughput)
	fmt.Printf("Ocupacao media (amostrada): %.2f/%d\n", result.AverageOccupancy, config.BufferCapacity)
	fmt.Printf("Timeouts: %d\n", result.Timeouts)
	for id, count := range result.PerConsumer {
		fmt.Printf("Consumidor %d: %d itens\n", id, count)
	}
}
