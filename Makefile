# ==================== [START] Global Variable Declaration =================== #
SHELL := /bin/bash
BASE_DIR := $(shell pwd)
UNAME_S := $(shell uname -s)
APP_NAME := imup-io/client
WORK_DIR := imup-io/client

# default client version
CLIENT_VERSION = v0.5.2
NDT7_CLIENT_NAME = ndt7-client-go
BUILD_LD_FLAGS = -ldflags="-X 'main.ClientVersion=$(CLIENT_VERSION)' -X 'main.ClientName=$(NDT7_CLIENT_NAME)' -X 'main.HoneybadgerAPIKey=$(HONEYBADGER_API_KEY)'"

export
# ===================== [END] Global Variable Declaration ==================== #

# =========================== [START] Build Scripts ========================== #
clean:
	@rm -f ./imup/*

# usage: 'make -e CLIENT_VERSION=v0.2.3 build'
.PHONY: build
build: clean
	@cd $(BASE_DIR) && env GOOS=linux GOARCH=amd64 go build $(BUILD_LD_FLAGS) -o imup/imup-linux-amd64
	@cd $(BASE_DIR) && env GOOS=darwin GOARCH=amd64 go build $(BUILD_LD_FLAGS) -o imup/imup-darwin-amd64
	@cd $(BASE_DIR) && env GOOS=darwin GOARCH=arm64 go build $(BUILD_LD_FLAGS) -o imup/imup-darwin-arm64
	@cd $(BASE_DIR) && env GOOS=windows GOARCH=amd64 go build $(BUILD_LD_FLAGS) -o imup/imup-windows-amd64.exe

# find . -name *.go -exec sed -i -E "s#//go:build oss##" {} {} \;
# ============================ [END] Build Scripts =========================== #

# =========================== [START] Test Scripts =========================== #
test:
	@go test -count 1 -race -v -coverprofile c.out ./...

test_with_caching:
	@go test -race -v -coverprofile c.out ./...
# ============================ [END] Test Scripts ============================ #

# ========================= [START] Formatting Script ======================== #
gofmt:
	@go fmt github.com/$(WORK_DIR)/...

golint:
	@golint github.com/$(WORK_DIR)/...

govet:
	@go vet github.com/$(WORK_DIR)/...

run:
	@go run auth.go connectiontest.go environment.go http_client.go speedtest.go update.go main.go

lint: gofmt golint govet
# ========================== [END] Formatting Script ========================= #
