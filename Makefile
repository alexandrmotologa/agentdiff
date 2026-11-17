.PHONY: all build test clean web-install web-build bench studio

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
	@mkdir -p internal/server/dist
	@cp -r web/dist/* internal/server/dist/

bench: build
	./$(BUILD_DIR)/$(BINARY_NAME) bench --config examples/agentdiff.yaml --output-dir ./eval_results

studio: build
	./$(BUILD_DIR)/$(BINARY_NAME) studio --port 7676 --traces-dir ./eval_results

clean:
	rm -rf $(BUILD_DIR) web/dist eval_results
