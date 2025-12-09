#!/bin/bash
protoc --go_out=. --go_opt=paths=source_relative \
  --go-grpc_out=. --go-grpc_opt=paths=source_relative \
  --connect-go_out=. --connect-go_opt=paths=source_relative \
  bouncebot.proto
