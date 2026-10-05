# Classic Concurrency Problems in Go

Assignment T1 for **FPPD – Fundamentals of Parallel and Distributed Processing** (98713-04).

This repository contains two independent Go programs that solve classic synchronization problems using goroutines, channels, `select`, and counting semaphores, together with an analysis of deadlock, starvation, and fairness based on execution data.

## Authors

- Antonio Augusto Fell Dal Bem
- Arthur de Oliveira Ferreira
- Gabriel Michealsen Borges
- Lorenzo Santos de Souza de Moraes Bueno

## Repository structure

```
.
├── go.mod
├── README.md
├── philosophers/        # Problem 1: Dining Philosophers
├── producer-consumer/   # Problem 2: Bounded-Buffer Producer/Consumer
├── results/             # Collected execution data (CSV)
└── report/              # PDF report (3-5 pages)
```

## Requirements

- Go 1.21 or later (adjust to the version in `go.mod`) 
    - Install in cmd with 'winget install GoLang.Go'
- No external dependencies

## Problem 1: Dining Philosophers

N philosophers sit at a circular table, with forks modeled as channels. The program takes N (number of philosophers) and R (number of iterations) as parameters and terminates in an orderly way, with no blocked goroutines.

### Versions

| Version | Strategy | Deadlock-free |
|---------|----------|---------------|
| `naive` | Everyone picks the left fork first | No (demonstrates deadlock) |
| `hierarchy` | Resource hierarchy (symmetry breaking) | Yes |
| `limit` | At most N-1 philosophers at the table | Yes |
| `waiter` | Centralized arbitration by a waiter process | Yes |

### Running

```bash
cd philosophers
go run -race . -version=hierarchy -n=5 -r=100
```

| Flag | Description |
|------|-------------|
| `-version` | `naive`, `hierarchy`, `limit` or `waiter` |
| `-n` | Number of philosophers |
| `-r` | Number of iterations per philosopher |

### Collected metrics

- Number of meals per philosopher
- Average fork waiting time per philosopher

These are used in the report to compare the fairness of each strategy.

> Note: the `naive` version is expected to deadlock. It exists to demonstrate the problem.

## Problem 2: Producer/Consumer with Bounded Buffer

P producers and C consumers share a buffer of capacity K. Producers block when the buffer is full; consumers block when it is empty.

### Versions

- **Channel version:** the buffer is a Go channel with capacity K, relying on Go's native blocking semantics.
- **Semaphore version:** the buffer is a shared data structure protected by counting semaphores (`notEmpty`, `notFull`) and mutual exclusion on the read and write positions, so that no two producers write to the same slot and no two consumers take the same item.

### Features

- **Orderly shutdown:** once producers finish, consumers drain the remaining items and exit with no deadlock and no lost items. The program checks that total produced equals total consumed.
- **Consumption with timeout:** at least one consumer uses `select` with `time.After` to run an alternative behavior when no item arrives within a given interval.

### Running

```bash
cd producer-consumer
go run -race . -impl=channel -p=4 -c=4 -k=10 -items=1000
```

| Flag | Description |
|------|-------------|
| `-impl` | `channel` or `semaphore` |
| `-p` | Number of producers |
| `-c` | Number of consumers |
| `-k` | Buffer capacity |
| `-items` | Items produced per producer |

### Collected metrics

Varying K (e.g., 1, 10, 100) with P and C fixed:

- Total throughput
- Items consumed per consumer
- Average buffer occupancy

## Data race detection

All programs run cleanly with the race detector:

```bash
go run -race .
```

## Report

The PDF report is in `report/` and covers the design decisions, deadlock/starvation analysis (Coffman conditions), fairness data, the effect of buffer size, how orderly shutdown was implemented, and the use of AI tools.

## References

See the references section of the report.