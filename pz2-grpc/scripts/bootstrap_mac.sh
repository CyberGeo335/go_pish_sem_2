#!/usr/bin/env bash
set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_ROOT"

echo "Project root: $PROJECT_ROOT"

if ! command -v go >/dev/null 2>&1; then
  echo "ERROR: Go is not installed."
  echo "Install Go from https://go.dev/dl/ and rerun this script."
  exit 1
fi

if ! command -v protoc >/dev/null 2>&1; then
  if command -v brew >/dev/null 2>&1; then
    echo "protoc not found. Installing protobuf via Homebrew..."
    brew install protobuf
  else
    echo "ERROR: protoc not found and Homebrew is not installed."
    echo "Install Homebrew or install protoc manually."
    exit 1
  fi
fi

GOBIN_PATH="$(go env GOPATH)/bin"
export PATH="$PATH:$GOBIN_PATH"

echo "Installing Go protobuf plugins..."
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

if ! command -v protoc-gen-go >/dev/null 2>&1; then
  echo "ERROR: protoc-gen-go is still not available in PATH."
  echo "Add this to ~/.zshrc:"
  echo 'export PATH="$PATH:$(go env GOPATH)/bin"'
  exit 1
fi

if ! command -v protoc-gen-go-grpc >/dev/null 2>&1; then
  echo "ERROR: protoc-gen-go-grpc is still not available in PATH."
  echo "Add this to ~/.zshrc:"
  echo 'export PATH="$PATH:$(go env GOPATH)/bin"'
  exit 1
fi

if [ ! -f "go.mod" ]; then
  echo "go.mod not found. Initializing module..."
  go mod init github.com/CyberGeo335/pz2-grpc
fi

MODULE_PATH="$(awk '/^module / {print $2}' go.mod)"

echo "Module path: $MODULE_PATH"

if [ ! -f "proto/student.proto" ]; then
  echo "ERROR: proto/student.proto not found."
  echo "Run this script from the project or keep it inside scripts/bootstrap_mac.sh."
  exit 1
fi

echo "Cleaning old generated files..."
rm -rf gen/studentpb
rm -rf github.com
mkdir -p gen/studentpb

echo "Generating protobuf and gRPC Go code..."
protoc \
  --proto_path=proto \
  --go_out=. \
  --go_opt=module="$MODULE_PATH" \
  --go-grpc_out=. \
  --go-grpc_opt=module="$MODULE_PATH" \
  proto/student.proto

if [ ! -f "gen/studentpb/student.pb.go" ]; then
  echo "ERROR: gen/studentpb/student.pb.go was not generated."
  exit 1
fi

if [ ! -f "gen/studentpb/student_grpc.pb.go" ]; then
  echo "ERROR: gen/studentpb/student_grpc.pb.go was not generated."
  exit 1
fi

echo "Installing project dependencies..."
go get google.golang.org/grpc
go get google.golang.org/protobuf

echo "Tidying go.mod..."
go mod tidy

echo "Checking build..."
go build ./cmd/server
go build ./cmd/client

echo ""
echo "Project is ready."
echo "Run server:"
echo "  make server"
echo ""
echo "In another terminal run client:"
echo "  make client"
echo ""
echo "Generated files:"
echo "  gen/studentpb/student.pb.go"
echo "  gen/studentpb/student_grpc.pb.go"

