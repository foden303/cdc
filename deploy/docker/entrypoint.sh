#!/usr/bin/env bash
set -euo pipefail

CONFIG_PATH=${CONFIG_PATH:-/app/config.yaml}
PORT=${PORT:-5173}

log() { echo "[$(date -u +'%Y-%m-%dT%H:%M:%SZ')] $*"; }

NODE_PID=0
CDC_PID=0

# Start Next.js Frontend if present
if [ -f /app/server.js ]; then
  log "Starting Next.js Frontend on port ${PORT}..."
  node /app/server.js &
  NODE_PID=$!
else
  log "No /app/server.js found; skipping frontend."
fi

log "Starting CDC Backend..."
/app/cdc --config "${CONFIG_PATH}" &
CDC_PID=$!

_term() {
  log "Shutting down..."
  if [ "${NODE_PID}" -ne 0 ]; then kill -TERM "${NODE_PID}" 2>/dev/null || true; fi
  kill -TERM "${CDC_PID}" 2>/dev/null || true
  wait "${NODE_PID}" 2>/dev/null || true
  wait "${CDC_PID}" 2>/dev/null || true
  exit 0
}
trap _term SIGTERM SIGINT

# Wait for any process to exit
while true; do
  if ! kill -0 "${CDC_PID}" 2>/dev/null; then
    wait "${CDC_PID}" || true
    break
  fi
  if [ "${NODE_PID}" -ne 0 ] && ! kill -0 "${NODE_PID}" 2>/dev/null; then
    wait "${NODE_PID}" || true
    break
  fi
  sleep 1
done

log "One process exited; shutting down remaining processes..."
_term
