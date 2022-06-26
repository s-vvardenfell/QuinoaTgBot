#!/bin/sh
protoc --go-grpc_out=. proto/proto.proto
protoc --go_out=. proto/proto.proto