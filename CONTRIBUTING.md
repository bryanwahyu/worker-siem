# Contributing

Thanks for taking the time.

## Getting set up

```bash
git clone https://github.com/bryanwahyu/worker-siem.git
cd worker-siem
cp configs/worker.yaml.example configs/worker.yaml
go test ./...
```

You need Go 1.21+. The module has no third-party dependencies, and the tests
need no running API.

To try the worker without a Mini SIEM API, run it in dry-run mode. It parses
and spools events but never uploads them:

```bash
go run ./cmd/worker run --config configs/worker.yaml --dry-run
```

## Before you open a pull request

```bash
gofmt -l .          # must print nothing
go vet ./...
go test -race ./...
```

CI runs exactly these, plus a Docker build.

## What a good change looks like

- **One concern per PR.** A bug fix and a refactor in the same diff are hard to
  review and harder to revert.
- **Tests for behaviour, not for coverage.** If you fix a bug, add the test that
  would have caught it.
- **Respect the layers.** `internal/domain` holds entities, value objects and
  services, `internal/application` holds use cases and ports, and
  `internal/infrastructure` holds adapters (tailers, spool, HTTP, metrics).
  Domain code must not import infrastructure. See
  [architecture.md](architecture.md).
- **Do not weaken delivery.** The worker promises at-least-once delivery. A
  change to the spool or the upload path must not drop events on a failed
  request or a crash between batching and upload.
- **Explain the why in the commit message,** not the what.

## Security

Do not open a public issue for a vulnerability. See [SECURITY.md](SECURITY.md).

## Licence

By contributing you agree that your contributions are licensed under the MIT
Licence, the same terms that cover the rest of the project.
