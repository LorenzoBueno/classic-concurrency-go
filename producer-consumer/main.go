package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)

type Config struct {
	Producers        int
	Consumers        int
	BufferCapacity   int
	ItemsPerProducer int
	Timeout          time.Duration
}

func (c Config) validate() error {
	if c.Producers <= 0 {
		return fmt.Errorf("p deve ser maior que zero")
	}
	if c.Consumers <= 0 {
		return fmt.Errorf("c deve ser maior que zero")
	}
	if c.BufferCapacity <= 0 {
		return fmt.Errorf("k deve ser maior que zero")
	}
	if c.ItemsPerProducer <= 0 {
		return fmt.Errorf("items deve ser maior que zero")
	}
	if c.Timeout <= 0 {
		return fmt.Errorf("timeout deve ser maior que zero")
	}
	return nil
}

func main() {
	implementation := flag.String("impl", "channel", "implementacao do buffer: channel")
	producers := flag.Int("p", 4, "numero de produtores")
	consumers := flag.Int("c", 4, "numero de consumidores")
	capacity := flag.Int("k", 10, "capacidade do buffer")
	items := flag.Int("items", 1000, "itens produzidos por produtor")
	timeout := flag.Duration("timeout", 10*time.Millisecond, "timeout do primeiro consumidor")
	flag.Parse()

	if *implementation != "channel" {
		fmt.Fprintf(os.Stderr, "implementacao %q ainda nao esta disponivel; use -impl=channel\n", *implementation)
		os.Exit(2)
	}

	config := Config{
		Producers:        *producers,
		Consumers:        *consumers,
		BufferCapacity:   *capacity,
		ItemsPerProducer: *items,
		Timeout:          *timeout,
	}
	if err := config.validate(); err != nil {
		fmt.Fprintf(os.Stderr, "configuracao invalida: %v\n", err)
		os.Exit(2)
	}

	result := runChannel(config)
	printResult(result, config)
	if result.Produced != result.Consumed {
		fmt.Fprintln(os.Stderr, "erro: o total produzido difere do total consumido")
		os.Exit(1)
	}
}
