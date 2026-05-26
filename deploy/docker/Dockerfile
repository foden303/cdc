# --- Stage 1: Build the Go CDC binary ---
FROM golang:1.23-bookworm AS builder-go
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags "-s -w" -o /out/cdc-bin ./cmd/cdc/main.go

# --- Stage 2: Build the Next.js UI (Standalone) ---
FROM node:20-bookworm AS builder-node
WORKDIR /app
COPY client/package*.json ./
RUN npm ci --silent
COPY client/ .
RUN npm run build

# --- Final Stage: Minimal runtime image ---
FROM node:19-slim

# Install tini, curl (healthcheck) and bash
RUN apt-get update \
  && apt-get install -y --no-install-recommends tini curl bash \
  && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy binaries
COPY --from=builder-go /out/cdc-bin /app/cdc

# Copy config
COPY config.example.yaml /app/config.yaml

# Copy Next.js Standalone build files
COPY --from=builder-node /app/.next/standalone /app
COPY --from=builder-node /app/.next/static /app/.next/static
COPY --from=builder-node /app/public /app/public

# Copy entrypoint script
COPY deploy/docker/entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh

# Production settings
ENV NODE_ENV=production
ENV PORT=5173

# Expose app-specific ports
EXPOSE 9090 9091 5173

# Lightweight healthcheck
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
  CMD curl -fsS http://127.0.0.1:9091/api/v1/health || exit 1

ENTRYPOINT ["/usr/bin/tini", "--", "/app/entrypoint.sh"]
