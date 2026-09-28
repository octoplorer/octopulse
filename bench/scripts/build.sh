#!/usr/bin/env bash
# Cross-compile the size/resource probes for the two release architectures.
# CGO is off so the result is a static binary that runs on distroless/scratch.
# Run through `mise run bench:build` so the Go version comes from mise.toml.
# -buildvcs=false keeps the recorded hashes comparable between commits: VCS
# stamping is on by default and makes the same source hash differently.
set -euo pipefail
cd "$(dirname "$0")/.."

rm -rf dist && mkdir -p dist
for arch in amd64 arm64; do
  for cmd in base combo sched; do
    GOOS=linux GOARCH="$arch" CGO_ENABLED=0 \
      go build -trimpath -buildvcs=false -ldflags="-s -w" -o "dist/${cmd}-linux-${arch}" "./cmd/${cmd}"
  done
done
echo "built with $(go version): $(ls dist | tr '\n' ' ')"
