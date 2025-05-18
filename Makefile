# Makefile for go-backup project

# Variables
APP_NAME := go-backup
VERSION := 0.1.0
BUILD_DIR := bin
PACKAGE := github.com/kennycyb/go-backup
LDFLAGS := -ldflags "-X main.Version=$(VERSION)"

# Go build environment variables
GO := go
GOFLAGS :=

# List of OS and architectures to build for
OS_ARCH := \
	linux/amd64 \
	linux/arm64 \
	darwin/amd64 \
	darwin/arm64 \
	windows/amd64 \
	windows/386

# Default target
.PHONY: all
all: build

# Build for the current platform
.PHONY: build
build:
	@echo "Building $(APP_NAME) for current platform..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME) .

# Build for all platforms
.PHONY: build-all
build-all: clean
	@echo "Building $(APP_NAME) for all platforms..."
	@mkdir -p $(BUILD_DIR)
	$(foreach os_arch, $(OS_ARCH), \
		$(eval OS := $(word 1, $(subst /, ,$(os_arch)))) \
		$(eval ARCH := $(word 2, $(subst /, ,$(os_arch)))) \
		$(eval SUFFIX := $(if $(filter windows,$(OS)),.exe,)) \
		$(eval OUTPUT := $(BUILD_DIR)/$(APP_NAME)-$(OS)-$(ARCH)$(SUFFIX)) \
		GOOS=$(OS) GOARCH=$(ARCH) $(GO) build $(GOFLAGS) $(LDFLAGS) -o $(OUTPUT) . && \
		echo "Built: $(OUTPUT)" && \
	) true

# Test the application
.PHONY: test
test:
	$(GO) test -v ./...

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning up..."
	@rm -rf $(BUILD_DIR)

# Install the application locally
.PHONY: install
install:
	$(GO) install $(GOFLAGS) $(LDFLAGS) .

# Run the application
.PHONY: run
run:
	$(GO) run $(GOFLAGS) $(LDFLAGS) .

# List backups
.PHONY: list
list:
	$(GO) run $(GOFLAGS) $(LDFLAGS) . list

# List backups with detailed information
.PHONY: list-detailed
list-detailed:
	$(GO) run $(GOFLAGS) $(LDFLAGS) . list --detailed

# List backups from a specific location
.PHONY: list-location
list-location:
	$(GO) run $(GOFLAGS) $(LDFLAGS) . list --path $(LOCATION)

# Help target
.PHONY: help
help:
	@echo "Makefile targets:"
	@echo "  all          - Default target, builds for current platform"
	@echo "  build        - Build for the current platform"
	@echo "  build-all    - Build for all configured platforms"
	@echo "  clean        - Remove build artifacts"
	@echo "  install      - Install the application locally"
	@echo "  run          - Run the application"
	@echo "  list         - List all backups"
	@echo "  list-detailed - List all backups with detailed information"
	@echo "  list-location - List backups from a specific location (LOCATION=path)"
	@echo "  test         - Run tests"
	@echo "  help         - Show this help message"
