#!/bin/sh
set -e

# Run from the proto directory where buf.yaml lives.
cd /workspace/proto

# Lint proto files first — abort on failure.
buf lint

# Generate Go code (output to ../api/proto/v1 as configured in buf.gen.yaml).
buf generate
