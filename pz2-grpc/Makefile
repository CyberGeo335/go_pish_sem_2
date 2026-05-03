MODULE := github.com/CyberGeo335/pz2-grpc
PROTO := proto/student.proto

.PHONY: help setup-tools deps proto tidy server client clean-generated check-protoc

help:
	@echo "Commands:"
	@echo "  make setup-tools      install Go protobuf/gRPC generators"
	@echo "  make deps             add Go dependencies to go.mod/go.sum"
	@echo "  make proto            generate Go files from proto/student.proto"
	@echo "  make server           run gRPC server on :50051"
	@echo "  make client           run gRPC client"
	@echo "  make tidy             run go mod tidy"
	@echo "  make clean-generated  remove generated protobuf Go files"

check-protoc:
	@command -v protoc >/dev/null 2>&1 || (echo "protoc not found. On macOS install it with: brew install protobuf" && exit 1)

setup-tools:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@echo "Make sure this path is in PATH: $$(go env GOPATH)/bin"

deps:
	go get google.golang.org/grpc
	go get google.golang.org/protobuf
	go mod tidy

proto: check-protoc
	protoc --proto_path=proto \
		--go_out=. --go_opt=module=$(MODULE) \
		--go-grpc_out=. --go-grpc_opt=module=$(MODULE) \
		$(PROTO)

tidy:
	go mod tidy

server:
	go run ./cmd/server

client:
	go run ./cmd/client

clean-generated:
	rm -f gen/studentpb/*.pb.go
