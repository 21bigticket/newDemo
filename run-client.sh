 go run go-client/cmd/client.go \
    -nacos-addr=192.168.139.230:8848 \
    -namespace=public \
    -group=DEFAULT_GROUP \
    -timeout=3s \
    -app-name=go-client \
    -port=20002 \
    -log-level=info \
    -help=true