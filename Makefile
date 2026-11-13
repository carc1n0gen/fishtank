GO ?= go
BIN_DIR := bin
BINARY := $(BIN_DIR)/fishtank

.PHONY: all build run test test-race vet fmt clean

all: build

build:
	mkdir -p $(BIN_DIR)
	$(GO) build -o $(BINARY) .

run: build
	./$(BINARY) $(ARGS)

test:
	$(GO) test ./...

test-race:
	$(GO) test -race ./...

vet:
	$(GO) vet ./...

fmt:
	$(GO) fmt ./...

clean:
	rm -rf $(BIN_DIR)
