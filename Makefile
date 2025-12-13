# Makefile for the loam project

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOINSTALL=$(GOCMD) install
BINARY_NAME=loam
BINARY_UNIX=$(BINARY_NAME)

# Default target
all: build

# Build the binary for the current platform
build:
	@echo "Building for $(GOOS)/$(GOARCH)..."
	$(GOBUILD) -v -o $(BINARY_NAME) ./cmd/loam

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# Clean build artifacts
clean:
	@echo "Cleaning up..."
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_NAME).exe

# Install the binary
install:
	@echo "Installing loam..."
	$(GOINSTALL) ./cmd/loam

# Cross-platform builds
cross-build: build-linux build-windows build-darwin

build-linux:
	@echo "Building for Linux..."
	GOOS=linux GOARCH=amd64 $(GOBUILD) -v -o $(BINARY_NAME)-linux-amd64 ./cmd/loam

build-windows:
	@echo "Building for Windows..."
	GOOS=windows GOARCH=amd64 $(GOBUILD) -v -o $(BINARY_NAME)-windows-amd64.exe ./cmd/loam

build-darwin:
	@echo "Building for macOS..."
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -v -o $(BINARY_NAME)-darwin-amd64 ./cmd/loam

.PHONY: all build test clean install cross-build build-linux build-windows build-darwin
