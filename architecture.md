# Mini SIEM Worker Architecture

## Overview
The `mini-siem-worker` follows a Domain-Driven Design (DDD) inspired architecture with explicit layering:

- **Domain**: Core business logic, entities, value objects, services, and repository abstractions.
- **Application**: Use cases that orchestrate domain services and repositories using DTOs.
- **Infrastructure**: Technology-specific implementations (tailers, persistence, config, worker pool, observability).
- **Interfaces**: CLI-facing commands and the program entrypoint.

The process ingests log lines from tailers, transforms them into domain `Event`s, batches uploads via a worker pool, spools failures to disk, and periodically flushes the spool. Observability is provided through Prometheus metrics and structured logging.

## Layer Interactions
```
Interfaces → Application → Domain ←→ Infrastructure
```

- Interfaces wire dependencies and expose CLI commands.
- Application coordinates workflows, translating DTOs into domain objects and invoking repositories/services.
- Domain enforces invariants via value objects and services, independent of infrastructure.
- Infrastructure provides concrete adapters for repositories, tailers, logging, metrics, and config.

## Concurrency Model
- Tailers feed parsed `Event`s into a buffered `eventsChan`.
- A worker pool (configurable 4–8 goroutines) consumes batches and uploads them via the API client.
- A dedicated flusher goroutine retries spooled events every 30 seconds.
- `context.Context` and `sync.WaitGroup` coordinate graceful shutdown.

## Resilience & Delivery
- Upload failures trigger exponential backoff retries.
- Persistent spool stored under `/var/lib/mini-siem/spool` ensures at-least-once delivery.
- Shutdown drains channels and flushes spool.

## Security & Configuration
- HTTP uploads attach an API key header sourced from configuration.
- Configuration is loaded from YAML files and environment variables.
- Dry-run mode bypasses API pushes while still tailing and spooling.

## Observability
- Prometheus metrics: `events_ingested_total`, `events_uploaded_total`, `events_failed_total`, `spool_queue_size`.
- Zerolog provides structured JSON logs across layers.

## Packages & Key Components
- `cmd/worker`: Main entry and dependency injection.
- `internal/domain`: Entities (`Event`, `SpoolItem`), value objects (`IP`, `StatusCode`), domain services (validation, transformation), repository interfaces.
- `internal/application`: Use cases (`IngestLogsUsecase`, `UploadEventsUsecase`, `FlushSpoolUsecase`) and DTOs.
- `internal/infrastructure`:
  - `tailer`: File/journald (stubbed interfaces) implementations.
  - `persistence`: HTTP API repository and spool repository.
  - `workerpool`: Concurrent upload pool with batching/backoff.
  - `config`: Loader merging YAML/env.
  - `observability`: Metrics and logging setup.
- `internal/interfaces/cli`: CLI commands (`worker run`, `worker flush`, `worker test`).

## Testing Strategy
- Unit tests for domain services, use cases, and worker pool behaviors using TDD-inspired approach.
- Integration-style tests (dry-run) verifying spool interactions.

## Deployment Artifacts
- `Makefile` with `build`, `run`, and `test` targets.
- `Dockerfile` (multi-stage build) for containerized deployment.
- `deploy/systemd/mini-siem-worker.service` for hardened service setup.

