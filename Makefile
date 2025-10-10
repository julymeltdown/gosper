SHELL := /bin/bash

.PHONY: deps build test itest lint

WHISPER_GO_DIR := whisper.cpp/bindings/go
WHISPER_BUILD  := $(WHISPER_GO_DIR)/build

deps:
	$(MAKE) -C $(WHISPER_GO_DIR) whisper

build: deps
	C_INCLUDE_PATH=$(PWD)/whisper.cpp \
	LIBRARY_PATH=$(PWD)/$(WHISPER_BUILD) \
	go build -o dist/gosper ./cmd/gosper

test:
	go test ./... -short -count=1 -race

itest: deps
	GOSPER_INTEGRATION=1 C_INCLUDE_PATH=$(PWD)/whisper.cpp LIBRARY_PATH=$(PWD)/$(WHISPER_BUILD) go test ./test/integration -count=1 -v

lint:
	@command -v golangci-lint >/dev/null 2>&1 && golangci-lint run || echo "golangci-lint not installed; skipping"

