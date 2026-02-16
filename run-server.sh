#!/bin/bash
go run go-server/cmd/server.go \
  -nacos-addr=192.168.139.230:8848 \
  -namespace=public \
  -group=DEFAULT_GROUP \
  -data-id=go-server-config \
  -timeout=3s \
  -app-name=go-server \
  -port=20001 \
  -log-level=info \
  -help=true
