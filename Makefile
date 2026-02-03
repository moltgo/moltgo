# Makefile for MoltGo

# Binary name
BINARY_NAME=moltgo

# Build directory
BIN_DIR=./bin

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Build the project
.PHONY: all
all: clean build

.PHONY: build
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BIN_DIR)
	$(GOBUILD) -o $(BIN_DIR)/$(BINARY_NAME) -v

.PHONY: clean
clean:
	@echo "Cleaning..."
	@$(GOCLEAN)
	@rm -rf $(BIN_DIR)

.PHONY: test
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

.PHONY: deps
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

.PHONY: run
run: build
	@$(BIN_DIR)/$(BINARY_NAME)

.PHONY: install
install:
	@echo "Installing $(BINARY_NAME)..."
	$(GOCMD) install

.PHONY: help
help:
	@echo "MoltGo Makefile targets:"
	@echo "  make build    - Build the binary to ./bin"
	@echo "  make clean    - Remove build artifacts"
	@echo "  make test     - Run tests"
	@echo "  make deps     - Download and tidy dependencies"
	@echo "  make run      - Build and run the binary"
	@echo "  make install  - Install binary to GOPATH/bin"
	@echo "  make all      - Clean and build (default)"
	@echo "  make help     - Show this help message"
