#!/usr/bin/env sh
set -e

CMD="./cmd/syspeek"
OUT="syspeek"

case "$1" in
  test)
    echo "Running unit tests..."
    CGO_ENABLED=0 go test -v ./...
    ;;
  cross-compile)
    echo "Cross-compiling syspeek..."
    mkdir -p build/dist
    GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o build/dist/syspeek-linux-amd64 "$CMD"
    GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -o build/dist/syspeek-linux-arm64 "$CMD"
    GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o build/dist/syspeek-darwin-amd64 "$CMD"
    GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -o build/dist/syspeek-darwin-arm64 "$CMD"
    GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o build/dist/syspeek-windows-amd64.exe "$CMD"
    echo "Artifacts built in build/dist/"
    ;;
  clean)
    rm -rf build syspeek
    echo "Cleaned."
    ;;
  *)
    echo "Building syspeek..."
    CGO_ENABLED=0 go build -ldflags="-s -w" -o "$OUT" "$CMD"
    echo "Built ./$OUT"
    ;;
esac
