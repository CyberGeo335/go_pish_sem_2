.PHONY: run tidy test docker-up docker-down traffic

run:
	go run ./cmd/server

tidy:
	go mod tidy

test:
	go test ./...

docker-up:
	docker compose up --build

docker-down:
	docker compose down

traffic:
	./scripts/generate_traffic.sh
