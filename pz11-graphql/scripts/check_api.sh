#!/usr/bin/env bash
set -euo pipefail

ENDPOINT="${ENDPOINT:-http://localhost:8080/query}"

echo "Checking API endpoint: ${ENDPOINT}"

curl -sS -X POST "$ENDPOINT" \
  -H 'Content-Type: application/json' \
  -d '{"query":"query { tasks { id title description done } }"}'

echo
