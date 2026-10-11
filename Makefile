BINARY ?= bin/mini-siem-worker
CONFIG ?= configs/worker.yaml

.PHONY: build
build:
	go build -o $(BINARY) ./cmd/worker

.PHONY: run
run:
	go run ./cmd/worker --config $(CONFIG)

.PHONY: test
test:
	go test ./...

.PHONY: tidy
tidy:
	go mod tidy
