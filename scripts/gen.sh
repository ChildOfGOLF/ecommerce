#!/usr/bin/env bash

set -euo pipefail

mkdir -p gen

protoc \
  --proto_path=proto \
  --go_out=gen --go_opt=paths=source_relative \
  --go-grpc_out=gen --go-grpc_opt=paths=source_relative \
  $(find proto -name '*.proto')

echo "protobuf generated"