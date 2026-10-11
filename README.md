# mini-siem-worker

[![CI](https://github.com/bryanwahyu/worker-siem/actions/workflows/ci.yml/badge.svg)](https://github.com/bryanwahyu/worker-siem/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

A concurrent worker agent for the Mini SIEM API built with Go and a DDD-inspired architecture. The worker tails local logs, validates and transforms them into normalized events, uploads them to the Mini SIEM API, and guarantees at-least-once delivery via a persistent spool.

## Features
- File and journald tailers feeding a central ingestion pipeline
- Domain services enforcing event invariants and transformation logic
- Concurrent worker pool with configurable batch size, interval, and concurrency
- Disk-backed spool with exponential backoff retries and manual flush command
- Dry-run mode that parses and spools without API calls
- Prometheus-format metrics and structured JSON logging, standard library only
- Configuration via YAML and environment variables

## Getting Started
```bash
go install github.com/bryanwahyu/worker-siem/cmd/worker@latest   # installs as `worker`; or build from source:

make build    # Compile the mini-siem-worker binary
make test     # Run unit tests
make run      # Run worker using default configs/worker.yaml
```

### Configuration
Copy `configs/worker.yaml.example` (or create one) into `configs/worker.yaml` and adjust:

```yaml
api:
  base_url: https://mini-siem.example.com
  api_key: ${MINI_SIEM_API_KEY}
  timeout: 10s

worker:
  batch_size: 100
  batch_interval: 2s
  concurrency: 4

spool:
  directory: /var/lib/mini-siem/spool
  flush_interval: 30s

metrics:
  listen_addr: ":2112"

tailers:
  files:
    - /var/log/syslog
  journald_unit: "nginx.service"

host_id: my-node-1
```

Environment variables override config values, for example `MINI_SIEM_API_KEY`, `MINI_SIEM_API_BASE_URL`, `MINI_SIEM_SPOOL_DIR` and `MINI_SIEM_WORKER_CONCURRENCY`. The full list is in `internal/infrastructure/config/config.go`.

### Commands
- `mini-siem-worker run` – start the worker and all background goroutines
- `mini-siem-worker flush` – force a spool flush immediately
- `mini-siem-worker test` – run lightweight diagnostics (spool reachability)

Use `--config` to point to an alternate YAML file and `--dry-run` to enable parse + spool mode without uploads.

### Observability
Metrics are exposed at `/metrics` on the configured listen address. Key counters:
- `mini_siem_worker_events_ingested_total`
- `mini_siem_worker_events_uploaded_total`
- `mini_siem_worker_events_failed_total`
- `mini_siem_worker_spool_queue_size`

Logs are written to stdout as one JSON object per line.

## Deployment
- `Dockerfile` supports a multi-stage build producing a minimal deployment image.
- `deploy/systemd/mini-siem-worker.service` provides a hardened unit file suitable for Linux hosts.

## Testing
Unit tests focus on domain value objects, use cases, and persistence adapters. Extend with integration tests for tailers or HTTP repositories as needed.

## Contributing
See [CONTRIBUTING.md](CONTRIBUTING.md). Report vulnerabilities privately as described in [SECURITY.md](SECURITY.md).

## License
[MIT](LICENSE)
