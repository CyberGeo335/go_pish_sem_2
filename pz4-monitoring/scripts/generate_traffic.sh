#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080}"

for i in {1..20}; do curl -s "$BASE_URL/health" > /dev/null; done
for i in {1..15}; do curl -s "$BASE_URL/students/1" > /dev/null; done
for i in {1..10}; do curl -s "$BASE_URL/students/2" > /dev/null; done
for i in {1..5}; do curl -s "$BASE_URL/students/999" > /dev/null; done

echo "Traffic generated for $BASE_URL"
