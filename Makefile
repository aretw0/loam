.PHONY: all build test test-fast coverage clean check install \
	cross-build build-linux build-windows build-darwin \
	work-on-lifecycle work-on-procio work-on-introspection \
	work-off-lifecycle work-off-procio work-off-introspection work-off-all

# Go parameters
GOCMD=go
GOWORK=$(GOCMD) work
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOTOOL=$(GOCMD) tool
GOINSTALL=$(GOCMD) install
BINARY_NAME=loam
BINARY_UNIX=$(BINARY_NAME)

# --- OS Detection & Command Abstraction ---
ifeq ($(OS),Windows_NT)
BINARY := $(BINARY_NAME).exe
RM := del /F /Q
# Windows needs backslashes for 'go work edit -dropuse' to match go.work content
DROP_WORK = if exist go.work ( $(GOWORK) edit -dropuse $(subst /,\,$(1)) )
INIT_WORK = if not exist go.work ( echo "Initializing go.work..." & $(GOWORK) init . )
else
BINARY := $(BINARY_UNIX)
RM := rm -f
# Linux/macOS uses forward slashes
DROP_WORK = [ -f go.work ] && $(GOWORK) edit -dropuse $(1)
INIT_WORK = [ -f go.work ] || ( echo "Initializing go.work..." && $(GOWORK) init . )
endif

# Default target
all: build

# Build the binary for the current platform
build:
	@echo "Building for $(GOOS)/$(GOARCH)..."
	$(GOBUILD) -v -o $(BINARY_NAME) ./cmd/loam

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -race -timeout 120s ./...

# Run tests excluding stress/benchmarks (Fast Feedback)
test-fast:
	@echo "Running fast tests (skipping stress/benchmarks)..."
	$(GOTEST) -timeout 60s ./pkg/... ./cmd/... ./internal/... ./tests/e2e ./tests/reactivity ./tests/typed

# Run coverage tests
coverage:
	$(GOTEST) -race -timeout 120s -coverprofile="coverage.out" ./...
	$(GOTOOL) cover -func="coverage.out"

# Clean build artifacts
clean:
	@echo "Cleaning up..."
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_NAME).exe
	rm -f $(BINARY_NAME)-*

# Verify all code and examples compile
check:
	@echo "Running static analysis and compilation check..."
	$(GOCMD) vet ./...

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

# --- Dependency Management (Dev vs Prod) ---

# Helper to get the correct path (uses WORK_PATH if provided, else default)
GET_PATH = $(if $(WORK_PATH),$(WORK_PATH),$(1))

# Enable local development mode for lifecycle
# Usage: make work-on-lifecycle [WORK_PATH=../lifecycle]
work-on-lifecycle:
	@echo "Enabling local lifecycle..."
	@$(INIT_WORK)
	$(GOWORK) use $(call GET_PATH,../lifecycle)
# Enable local development mode for procio
# Usage: make work-on-procio [WORK_PATH=../procio]
work-on-procio:
	@echo "Enabling local procio..."
	@$(INIT_WORK)
	$(GOWORK) use $(call GET_PATH,../procio)

# Enable local development mode for introspection
# Usage: make work-on-introspection [WORK_PATH=../introspection]
work-on-introspection:
	@echo "Enabling local introspection..."
	@$(INIT_WORK)
	$(GOWORK) use $(call GET_PATH,../introspection)

# Disable local lifecycle
# Usage: make work-off-lifecycle [WORK_PATH=../lifecycle]
work-off-lifecycle:
	@echo "Disabling local lifecycle..."
	@$(call DROP_WORK,$(call GET_PATH,../lifecycle))

# Disable local procio
# Usage: make work-off-procio [WORK_PATH=../procio]
work-off-procio:
	@echo "Disabling local procio..."
	@$(call DROP_WORK,$(call GET_PATH,../procio))

# Disable local introspection
# Usage: make work-off-introspection [WORK_PATH=../introspection]
work-off-introspection:
	@echo "Disabling local introspection..."
	@$(call DROP_WORK,$(call GET_PATH,../introspection))

# Disable local development mode by removing go.work (nuclear option)
work-off-all:
	@echo "Disabling local workspace mode..."
	@$(RM) go.work