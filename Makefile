BINARY := chunky
BIN_DIR := bin

.PHONY: all build run test clean

all: build

build:
	@mkdir -p $(BIN_DIR)
	@go build -o $(BIN_DIR)/$(BINARY) ./cmd/chunky

run: build
	@$(BIN_DIR)/$(BINARY)

test:
	@go test ./...

clean:
	@rm -rf $(BIN_DIR)