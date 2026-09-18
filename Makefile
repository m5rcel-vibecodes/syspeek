BINARY_NAME=syspeek
BUILD_DIR=build
VERSION=1.0.0
CMD_PATH=./cmd/syspeek

.PHONY: all build test clean install lint cross-compile run help

all: build test

build:
	@mkdir -p $(BUILD_DIR)
	go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_PATH)
	@echo "Built $(BUILD_DIR)/$(BINARY_NAME)"

test:
	go test -v ./...

lint:
	go vet ./...

clean:
	rm -rf $(BUILD_DIR) syspeek

install: build
	install -d /usr/local/bin
	install $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)

run: build
	./$(BUILD_DIR)/$(BINARY_NAME)

cross-compile:
	@mkdir -p $(BUILD_DIR)/dist
	@echo "Building for linux/amd64..."
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o $(BUILD_DIR)/dist/$(BINARY_NAME)-linux-amd64 $(CMD_PATH)
	@echo "Building for linux/arm64..."
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -o $(BUILD_DIR)/dist/$(BINARY_NAME)-linux-arm64 $(CMD_PATH)
	@echo "Building for darwin/amd64..."
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o $(BUILD_DIR)/dist/$(BINARY_NAME)-darwin-amd64 $(CMD_PATH)
	@echo "Building for darwin/arm64..."
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -o $(BUILD_DIR)/dist/$(BINARY_NAME)-darwin-arm64 $(CMD_PATH)
	@echo "Building for windows/amd64..."
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o $(BUILD_DIR)/dist/$(BINARY_NAME)-windows-amd64.exe $(CMD_PATH)
	@echo "Cross-compilation complete in $(BUILD_DIR)/dist/"

help:
	@echo "Available targets:"
	@echo "  build         - Compile the syspeek binary"
	@echo "  test          - Run all unit tests with race detection"
	@echo "  lint          - Run go vet"
	@echo "  clean         - Remove build artifacts"
	@echo "  install       - Install binary to /usr/local/bin"
	@echo "  cross-compile - Cross-compile for Linux, macOS, and Windows"
	@echo "  run           - Build and run syspeek"
