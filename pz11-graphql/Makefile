GQLGEN_VERSION ?= v0.17.49

.PHONY: bootstrap generate tidy fmt test run check

bootstrap:
	GQLGEN_VERSION=$(GQLGEN_VERSION) ./scripts/bootstrap.sh

generate:
	go run github.com/99designs/gqlgen@$(GQLGEN_VERSION) generate

tidy:
	go mod tidy

fmt:
	gofmt -w server.go graph/*.go internal/**/*.go

test:
	go test ./...

run:
	go run ./server.go

check:
	./scripts/check_api.sh
