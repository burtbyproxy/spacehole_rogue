.PHONY: build run clean test docker-build release

BINARY_NAME=spacehole
BUILD_DIR=build
CMD_DIR=cmd/spacehole

# Build for current platform
build:
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./$(CMD_DIR)

# Run the game
run:
	go run ./$(CMD_DIR)

# Clean build artifacts
clean:
	rm -rf $(BUILD_DIR)

# Run tests
test:
	go test ./...

# Build for Linux
build-linux:
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./$(CMD_DIR)

# Build for Windows
build-windows:
	GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./$(CMD_DIR)

# Build all release targets
release: build-linux build-windows

# Docker build for cross-compilation
docker-build:
	docker build -t spacehole-builder .
	docker run --rm -v $(PWD)/$(BUILD_DIR):/out spacehole-builder

# Tidy modules
tidy:
	go mod tidy
