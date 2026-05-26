# Examples

This folder contains scripts and configuration to run a local CDC test environment (NATS + PostgreSQL source/sink + Prometheus) and to generate test data.

## Files

- `docker-compose.yaml` — Docker Compose for the test environment (NATS, source/sink Postgres, Prometheus).
- `init-source.sql` — Source DB schema (users, products, orders).
- `init-sink.sql` — Sink DB schema.
- `seed-data.sh` — Seed sample data into the source DB.
- `continuous-insert.sh` — Continuous steady inserter (bulk COPY into a temp table then INSERT with ON CONFLICT). Use for long-running, steady load.
- `continuous-insert-fast.sh` — High-throughput bulk benchmark using COPY (generates large volumes quickly).
- `nats-data/` — persisted NATS data used by examples.

## Quickstart

1. Start the test environment

```bash
cd examples
docker compose -f docker-compose.yaml up -d
```

Services started:
- NATS (port 4222)
- Source PostgreSQL (port 5433) — configured with `wal_level=logical`
- Sink PostgreSQL (port 5434)
- Prometheus (port 9095)

2. Seed sample data

```bash
./seed-data.sh [NUM_USERS] [NUM_PRODUCTS] [NUM_ORDERS]
```

Defaults: `50 30 100`.

3. Continuous steady load (safe for long-running tests)

```bash
# Usage: ./continuous-insert.sh [USERS] [PRODUCTS] [ITERATIONS] [CONCURRENCY]
# Defaults: USERS=10000 PRODUCTS=1000 ITERATIONS=0(infinite) CONCURRENCY=1
nohup ./continuous-insert.sh 1000 200 0 2 > continuous.log 2>&1 &
```

- `continuous-insert.sh` is designed for steady traffic: it copies generated rows into a temporary table then inserts into the target table with `ON CONFLICT DO NOTHING` to avoid COPY failures from duplicates.

4. Fast bulk benchmark (max throughput)

```bash
# Usage: ./continuous-insert-fast.sh [USERS] [PRODUCTS] [ORDERS] [ITERATIONS] [CONCURRENCY]
# Example: generate very large volume once with 4 workers
./continuous-insert-fast.sh 20000 2000 100000 1 4
```

- `continuous-insert-fast.sh` uses bulk `COPY` and parallel workers to saturate disk/CPU for throughput measurement. Expect high resource usage.

## Stopping the environment

```bash
cd examples
docker compose -f docker-compose.yaml down
```

## Benchmarking tips

- For ephemeral benchmarks only, consider temporarily setting `synchronous_commit = off` in Postgres to reduce commit latency (DO NOT use in production).
- Increase `CONCURRENCY` and batch sizes to saturate the system.
- Use `nohup`/`tmux`/`screen` or systemd for background runs and collect logs.
- If you see UNIQUE constraint errors, prefer the continuous script (it handles duplicates) or adjust scripts to generate guaranteed-unique keys.

## Notes

- All scripts use `docker exec` so `psql` is not required on the host machine.
- The source DB is configured for logical replication (`wal_level=logical`) required by CDC.
- These scripts are for testing/benchmarking only — do not apply the same settings to production databases.

If you need further customization or a benchmarking playbook, open an issue or ask here.
