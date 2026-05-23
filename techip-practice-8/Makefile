SERVICE_DIR := services/tasks
IMAGE_NAME := techip-tasks
IMAGE_TAG ?= 0.1

.PHONY: test build run docker-build compose-up compose-down fmt

fmt:
	cd $(SERVICE_DIR) && gofmt -w .

test:
	cd $(SERVICE_DIR) && go test ./...

build:
	cd $(SERVICE_DIR) && go build ./...

run:
	cd $(SERVICE_DIR) && HTTP_ADDR=:8080 go run ./cmd/server

docker-build:
	docker build -t $(IMAGE_NAME):$(IMAGE_TAG) ./$(SERVICE_DIR)

compose-up:
	docker compose up --build

compose-down:
	docker compose down
