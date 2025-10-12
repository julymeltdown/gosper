SHELL := /bin/bash

.PHONY: deps build-all build-cli build-server test itest lint tidy

CGO_INCLUDE_PATH := $(PWD)/whisper.cpp/include:$(PWD)/whisper.cpp/ggml/include
CGO_LIBRARY_PATH := $(PWD)/whisper.cpp/build_go/src:$(PWD)/whisper.cpp/build_go/ggml/src
CGO_LDFLAGS := -L$(CGO_LIBRARY_PATH) -lwhisper -lggml -lggml-base -lggml-cpu -lm -lstdc++ -fopenmp

WHISPER_GO_DIR := whisper.cpp/bindings/go
WHISPER_BUILD  := $(WHISPER_GO_DIR)/build

deps:
	git submodule update --init --recursive
	$(MAKE) -C $(WHISPER_GO_DIR) whisper

tidy:
	go mod tidy

build-all: tidy build-cli build-server

build-cli: deps
	C_INCLUDE_PATH=$(CGO_INCLUDE_PATH) LIBRARY_PATH=$(CGO_LIBRARY_PATH) go build -tags "cli malgo whisper" -ldflags="$(CGO_LDFLAGS)" -o dist/gosper ./cmd/gosper

build-server: deps
	C_INCLUDE_PATH=$(CGO_INCLUDE_PATH) LIBRARY_PATH=$(CGO_LIBRARY_PATH) go build -tags "whisper" -ldflags="$(CGO_LDFLAGS)" -o dist/server ./cmd/server

test:
	go test ./... -short -count=1 -race

itest: deps
	GOSPER_INTEGRATION=1 C_INCLUDE_PATH=$(CGO_INCLUDE_PATH) LIBRARY_PATH=$(CGO_LIBRARY_PATH) go test ./test/integration -count=1 -v

lint:
	@command -v golangci-lint >/dev/null 2>&1 && golangci-lint run || echo "golangci-lint not installed; skipping"
