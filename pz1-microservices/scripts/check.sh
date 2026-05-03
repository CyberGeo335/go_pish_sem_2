#!/usr/bin/env bash
set -euo pipefail

echo "1) user-service: GET /users/1"
curl -i http://localhost:8081/users/1
printf '\n\n'

echo "2) user-service: GET /users"
curl -i http://localhost:8081/users
printf '\n\n'

echo "3) order-service: GET /orders/101"
curl -i http://localhost:8082/orders/101
printf '\n\n'

echo "4) order-service: GET /orders/101/full"
curl -i http://localhost:8082/orders/101/full
printf '\n\n'

echo "5) order-service: GET /orders/by-user/1"
curl -i http://localhost:8082/orders/by-user/1
printf '\n\n'

echo "6) expected 404: GET /users/999"
curl -i http://localhost:8081/users/999 || true
printf '\n\n'

echo "Done."
