# syntax=docker/dockerfile:1

FROM golang:1.21-alpine AS builder
WORKDIR /workspace
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /workspace/bin/mini-siem-worker ./cmd/worker

FROM gcr.io/distroless/base-debian12
USER nonroot:nonroot
WORKDIR /home/nonroot
COPY --from=builder /workspace/bin/mini-siem-worker /usr/local/bin/mini-siem-worker

ENV MINI_SIEM_CONFIG=/etc/mini-siem/worker.yaml

ENTRYPOINT ["/usr/local/bin/mini-siem-worker", "run", "--config", "/etc/mini-siem/worker.yaml"]
