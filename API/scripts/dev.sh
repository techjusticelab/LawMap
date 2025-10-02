#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
export EXAMPLES_FILE=${EXAMPLES_FILE:-docs/EXAMPLES.graph.jsonl}
export PORT=${PORT:-8080}
echo "Starting API on :$PORT using $EXAMPLES_FILE"
GOCACHE="$(pwd)/.gocache" go run ./cmd/api
