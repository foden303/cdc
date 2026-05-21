#!/bin/sh
set -e

# This script runs inside the Docker container.
# It copies buf config to workspace root (required by buf v2),
# runs buf generate, then cleans up the temporary config files.

WORKSPACE="/workspace"
PROTO_DIR="${WORKSPACE}/proto"

# Copy buf config to workspace root (buf v2 requires buf.yaml at root relative to module path)
cp "${PROTO_DIR}/buf.yaml" "${WORKSPACE}/buf.yaml"
cp "${PROTO_DIR}/buf.gen.yaml" "${WORKSPACE}/buf.gen.yaml"
[ -f "${PROTO_DIR}/buf.lock" ] && cp "${PROTO_DIR}/buf.lock" "${WORKSPACE}/buf.lock"

# Run buf generate from workspace root
cd "${WORKSPACE}"
buf generate

# Clean up temporary config files
rm -f "${WORKSPACE}/buf.yaml" "${WORKSPACE}/buf.gen.yaml" "${WORKSPACE}/buf.lock"
