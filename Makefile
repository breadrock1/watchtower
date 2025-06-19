BIN_LINTER := "${GOPATH}/bin/golangci-lint"
SERVICE_BIN_FILE_PATH := "./bin/watchtower"

GIT_HASH := $(shell git log --format="%h" -n 1)

build:
	go build -v -o $(SERVICE_BIN_FILE_PATH) ./cmd/watchtower

run: build
	$(SERVICE_BIN_FILE_PATH) -c ./configs/config.toml

test:
	go test -race ./tests/...

.PHONY: build run test