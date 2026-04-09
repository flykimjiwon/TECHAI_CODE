VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BINARY = techai

# v0.4.0: Debug always ON for all builds
LDFLAGS = -ldflags "-s -w -X main.version=$(VERSION) \
	-X 'github.com/kimjiwon/tgc/internal/config.DebugMode=true'"

# On-premise build config
ONPREM_URL = https://techai-web-prod.shinhan.com/v1
ONPREM_MODEL_SUPER = openai/gpt-oss-120b
ONPREM_MODEL_DEV = qwen/qwen3-coder-30b
ONPREM_LDFLAGS = -ldflags "-s -w -X main.version=$(VERSION)-onprem \
	-X 'github.com/kimjiwon/tgc/internal/config.DefaultBaseURL=$(ONPREM_URL)' \
	-X 'github.com/kimjiwon/tgc/internal/config.DefaultModel=$(ONPREM_MODEL_SUPER)' \
	-X 'github.com/kimjiwon/tgc/internal/config.DefaultDevModel=$(ONPREM_MODEL_DEV)' \
	-X 'github.com/kimjiwon/tgc/internal/config.ConfigDirName=.tgc-onprem' \
	-X 'github.com/kimjiwon/tgc/internal/config.DebugMode=true'"

.PHONY: build run clean build-all build-onprem install lint test build-index

# Knowledge index
build-index:
	@echo "Building knowledge index..."
	@go run ./cmd/build-index/
	@echo "Index built."

build: build-index
	go build $(LDFLAGS) -o $(BINARY) ./cmd/tgc

run: build
	./$(BINARY)

install:
	go install $(LDFLAGS) ./cmd/tgc

clean:
	rm -f $(BINARY)
	rm -rf dist/

build-all: clean
	mkdir -p dist
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY)-darwin-arm64 ./cmd/tgc
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY)-darwin-amd64 ./cmd/tgc
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY)-windows-amd64.exe ./cmd/tgc
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY)-linux-amd64 ./cmd/tgc
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY)-linux-arm64 ./cmd/tgc

build-onprem:
	mkdir -p dist
	GOOS=darwin GOARCH=arm64 go build $(ONPREM_LDFLAGS) -o dist/$(BINARY)-onprem-darwin-arm64 ./cmd/tgc
	GOOS=darwin GOARCH=amd64 go build $(ONPREM_LDFLAGS) -o dist/$(BINARY)-onprem-darwin-amd64 ./cmd/tgc
	GOOS=windows GOARCH=amd64 go build $(ONPREM_LDFLAGS) -o dist/$(BINARY)-onprem-windows-amd64.exe ./cmd/tgc
	GOOS=linux GOARCH=amd64 go build $(ONPREM_LDFLAGS) -o dist/$(BINARY)-onprem-linux-amd64 ./cmd/tgc
	GOOS=linux GOARCH=arm64 go build $(ONPREM_LDFLAGS) -o dist/$(BINARY)-onprem-linux-arm64 ./cmd/tgc

lint:
	go vet ./...

test:
	go test ./...
