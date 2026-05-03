#!/usr/bin/env bash
set -euo pipefail

curl -k https://localhost:8443/health
printf '\n'
curl -k "https://localhost:8443/students?id=1"
printf '\n'
curl -k "https://localhost:8443/students/by-email?email=ivanov@example.com"
printf '\n'
curl -k -i "https://localhost:8443/students?id=1%20OR%201=1"
printf '\n'
