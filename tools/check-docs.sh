#!/usr/bin/env bash
set -euo pipefail

files=(
  "API/docs/API.md"
  "API/docs/GRAPH_MODEL.md"
  "API/docs/api/graph_endpoints.md"
  "API/docs/api/dto.md"
  "API/docs/openapi.yaml"
  "API/docs/EXAMPLES.graph.jsonl"
)

for f in "${files[@]}"; do
  if [[ ! -f "$f" ]]; then
    echo "Missing required doc: $f" >&2
    exit 1
  fi
  if [[ ! -s "$f" ]]; then
    echo "Doc is empty: $f" >&2
    exit 1
  fi
done

grep -q "NodeDTO" API/docs/api/dto.md || { echo "DTO docs missing NodeDTO section" >&2; exit 1; }
grep -q "Endpoints" API/docs/api/graph_endpoints.md || { echo "Endpoints doc missing heading" >&2; exit 1; }

echo "Docs consistency checks passed."
