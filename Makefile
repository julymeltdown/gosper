SHELL := /bin/bash

.PHONY: deps build-all build-cli build-server build-cli-only build-server-only proto test itest lint tidy clean help

CGO_INCLUDE_PATH := $(PWD)/whisper.cpp:$(PWD)/whisper.cpp/include:$(PWD)/whisper.cpp/ggml/include
WHISPER_LIB_PATH := $(PWD)/whisper.cpp/build_go/src
GGML_LIB_PATH := $(PWD)/whisper.cpp/build_go/ggml/src
CGO_LDFLAGS_VALUE := -L$(WHISPER_LIB_PATH) -L$(GGML_LIB_PATH) -lwhisper -lggml -lggml-base -lggml-cpu -lm -lstdc++ -fopenmp

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
	C_INCLUDE_PATH="$(CGO_INCLUDE_PATH)" \
	CGO_LDFLAGS="$(CGO_LDFLAGS_VALUE)" \
		go build -tags "cli malgo whisper" -o dist/gosper ./cmd/gosper

# Build server binary only (no deps check)
build-server-only:
	C_INCLUDE_PATH="$(CGO_INCLUDE_PATH)" \
	CGO_LDFLAGS="$(CGO_LDFLAGS_VALUE)" \
		go build -tags "whisper" -o dist/server ./cmd/server

# Generate protobuf code
proto:
	@echo "Generating protobuf code..."
	@mkdir -p pkg/grpc/gen/go
	protoc \
		--go_out=pkg/grpc/gen/go \
		--go_opt=paths=source_relative \
		--go-grpc_out=pkg/grpc/gen/go \
		--go-grpc_opt=paths=source_relative \
		--proto_path=api/proto \
		api/proto/gosper/v1/*.proto
	@echo "Protobuf code generated successfully"

test:
	go test ./... -short -count=1 -race

itest: deps
	GOSPER_INTEGRATION=1 \
	C_INCLUDE_PATH="$(CGO_INCLUDE_PATH)" \
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
	@echo "  proto        - Generate protobuf code from .proto files"
	@echo "  test         - Run unit tests"
	@echo "  itest        - Run integration tests"
	@echo "  lint         - Run golangci-lint"
	@echo "  clean        - Remove build artifacts"
	@echo "  help         - Show this help message"
