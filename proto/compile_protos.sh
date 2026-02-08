#!/bin/bash
set -e

# Generate Go protos
echo "Generating Go protos..."
protoc --go_out=. --go_opt=paths=source_relative \
  --go-grpc_out=. --go-grpc_opt=paths=source_relative \
  --connect-go_out=. --connect-go_opt=paths=source_relative \
  bouncebot.proto

# Generate TypeScript protos
echo "Generating TypeScript protos..."
cd ../client/web
npx buf generate

echo "Done!"
