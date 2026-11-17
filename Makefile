.PHONY: all build test clean web-build studio

BINARY_NAME=agentdiff
BUILD_DIR=bin

all: build

build:
	go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/agentdiff

test:
	go test -v -race ./...

web-install:
	cd web && npm install

web-build:
	cd web && npm run build

clean:
	rm -rf $(BUILD_DIR) web/dist
