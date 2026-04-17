# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Common development commands

### Backend (Go)
- Build backend binary: `make build`
- Run backend: `make run`
- Run all backend tests: `make test` (equivalent to `go test -v ./...`)
- Run a single test: `go test -v ./path/to/package -run TestName`
- Tidy dependencies: `make tidy`
- Generate protobuf + gateway code: `make gen-proto`

### Frontend (Next.js, `client/`)
- Install deps: `cd client && npm install`
- Run dev server (port 9092): `cd client && npm run dev`
- Build frontend: `cd client && npm run build`
- Start production frontend (port 9092): `cd client && npm run start`
- Lint frontend: `cd client && npm run lint`

### Local infrastructure
- Start docker dependencies: `make up`
- Stop docker dependencies: `make down`
- Optional multi-node NATS setup: `docker compose -f docker-compose-cluster.yaml up -d`

## Architecture overview

### Runtime boot flow
- Entrypoint: `cmd/cdc/main.go`.
- Startup order:
  1. Load static config from `pkg/config/config.yaml` via `pkg/config`.
  2. Create NATS JetStream client (`pkg/nats/client.go`).
  3. Restore persisted source/sink configs from NATS KV and recompile CEL transformations.
  4. Build concrete source/sink instances via registry (`pkg/registry`).
  5. Ensure main stream + DLQ stream exist.
  6. Start pipeline engine and gRPC/REST server.

### Source → NATS → Sink pipeline
- Core engine: `pkg/pipeline/engine.go`.
- Sources push `models.Event` into `eventCh`.
- Producer batches events and publishes to JetStream with subject format:
  - `<topic>.<instance_id>.<schema>.<table>.<partition>`
- Publish path stores routing metadata in NATS headers (`pkg/nats/publisher.go`) and sets `Nats-Msg-Id` from `instance_id + offset` for dedupe.
- Sink workers are partition-aware pull consumers (`pipeline-<sink>-p<partition>`), fetch from filtered subjects, write batches to sinks, flush, then ACK.
- After successful sink write:
  - source offsets are persisted in KV (`CDC_STATE`),
  - ACK LSN is sent back to source via per-source ack channels.
- On repeated sink failure, messages are moved to DLQ (`dlq.>`).

### Dynamic config and hot reload
- Management API (`pkg/server/source.go`, `pkg/server/sink.go`) supports add/update/remove of sources and sinks at runtime.
- Config is persisted in NATS KV bucket `CDC_CONFIG` under keys:
  - `cfg.sources.<instance_id>`
  - `cfg.sinks.<instance_id>`
- Engine watches KV updates (`configWatcherLoop`) and hot-reloads connectors.
- Important: persisted KV configs can override static YAML source/sink definitions on startup.

### Connector/plugin model
- Interfaces are in `pkg/interfaces/provider.go`.
- Implementations register themselves via `init()` in connector packages and are instantiated through `pkg/registry/registry.go`.
- Sources: `pkg/source/postgres`, `pkg/source/mysql`, `pkg/source/rest` (plus snapshot helpers under `pkg/source/snapshot`).
- Sinks: `pkg/sink/postgres`, `pkg/sink/elasticsearch`, `pkg/sink/clickhouse`, `pkg/sink/redis`, `pkg/sink/webhook`, `pkg/sink/stdout`.

### API surface and transport
- API contract: `api/proto/v1/cdc.proto`.
- Generated files (`*.pb.go`, `*.pb.gw.go`) are in the same directory; regenerate with `make gen-proto` (Buf config: `buf.yaml`, `buf.gen.yaml`).
- Server wiring: `pkg/server/server.go`.
  - gRPC: port 9090
  - REST gateway (grpc-gateway): port 9091
  - Prometheus metrics: `/metrics`
  - Health: `/health`

### Frontend integration
- Frontend app is in `client/` (Next.js).
- `client/next.config.ts` rewrites `/api/v1/*` to `http://localhost:9091/api/v1/*`.
- Docker image builds backend + standalone Next.js runtime (`Dockerfile`, `entrypoint.sh`) and exposes 9090/9091/9092.
