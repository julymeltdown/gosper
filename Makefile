SHELL := /bin/bash

.PHONY: deps build-all build-cli build-server build-cli-only build-server-only test itest lint tidy clean help

CGO_INCLUDE_PATH := $(PWD)/whisper.cpp/include:$(PWD)/whisper.cpp/ggml/include
CGO_LIBRARY_PATH := $(PWD)/whisper.cpp/build_go/src:$(PWD)/whisper.cpp/build_go/ggml/src
CGO_LDFLAGS_VALUE := -L$(CGO_LIBRARY_PATH) -lwhisper -lggml -lggml-base -lggml-cpu -lm -lstdc++ -fopenmp

WHISPER_GO_DIR := whisper.cpp/bindings/go
WHISPER_BUILD  := $(WHISPER_GO_DIR)/build

deps:
	git submodule update --init --recursive
	$(MAKE) -C $(WHISPER_GO_DIR) whisper

tidy:
	go mod tidy

# Build all binaries (runs deps once)
build-all: deps tidy build-cli-only build-server-only

# Build CLI with deps check
build-cli: deps build-cli-only

# Build server with deps check
build-server: deps build-server-only

# Build CLI binary only (no deps check)
build-cli-only:
	CGO_CFLAGS="-I$(CGO_INCLUDE_PATH)" \
	CGO_LDFLAGS="$(CGO_LDFLAGS_VALUE)" \
		go build -tags "cli malgo whisper" -o dist/gosper ./cmd/gosper

# Build server binary only (no deps check)
build-server-only:
	CGO_CFLAGS="-I$(CGO_INCLUDE_PATH)" \
	CGO_LDFLAGS="$(CGO_LDFLAGS_VALUE)" \
		go build -tags "whisper" -o dist/server ./cmd/server

test:
	go test ./... -short -count=1 -race

itest: deps
	GOSPER_INTEGRATION=1 \
	CGO_CFLAGS="-I$(CGO_INCLUDE_PATH)" \
	CGO_LDFLAGS="$(CGO_LDFLAGS_VALUE)" \
		go test ./test/integration -count=1 -v

lint:
	@command -v golangci-lint >/dev/null 2>&1 && golangci-lint run || echo "golangci-lint not installed; skipping"

clean:
	rm -rf dist/

help:
	@echo "Available targets:"
	@echo "  deps         - Initialize submodules and build whisper.cpp"
	@echo "  tidy         - Run go mod tidy"
	@echo "  build-all    - Build all binaries (CLI + server)"
	@echo "  build-cli    - Build CLI binary"
	@echo "  build-server - Build server binary"
	@echo "  test         - Run unit tests"
	@echo "  itest        - Run integration tests"
	@echo "  lint         - Run golangci-lint"
	@echo "  clean        - Remove build artifacts"
	@echo "  help         - Show this help message"
