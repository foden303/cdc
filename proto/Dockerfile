# Minimal Docker image for protobuf code generation using Buf CLI.
# Uses Buf's remote plugins — no local protoc or Go plugins needed.
FROM alpine:3.21@sha256:48b0309ca019d89d40f670aa1bc06e426dc0931948452e8491e3d65087abc07d

# Pin Buf CLI version
ARG BUF_VERSION=1.50.0

# Install minimal dependencies: ca-certificates (HTTPS) and git (buf module resolution)
RUN apk add --no-cache ca-certificates git

# Download Buf CLI binary for the correct architecture
RUN set -eux; \
    ARCH="$(uname -m)"; \
    case "${ARCH}" in \
      x86_64)  BUF_ARCH="x86_64" ;; \
      aarch64) BUF_ARCH="aarch64" ;; \
      *)       echo "Unsupported architecture: ${ARCH}" >&2; exit 1 ;; \
    esac; \
    wget -q -O /usr/local/bin/buf \
      "https://github.com/bufbuild/buf/releases/download/v${BUF_VERSION}/buf-Linux-${BUF_ARCH}" \
    && chmod +x /usr/local/bin/buf \
    && buf --version

WORKDIR /workspace
