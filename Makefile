VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BINARY = techai

# v0.4.0: Debug always ON for all builds
LDFLAGS = -ldflags "-s -w -X main.version=$(VERSION) \
	-X 'github.com/kimjiwon/tgc/internal/config.DebugMode=true'"

# On-premise build config
ONPREM_URL = https://techai-web-prod.shinhan.com/v1
ONPREM_MODEL_SUPER = GPT-OSS-120B
ONPREM_MODEL_DEV = Qwen3-Coder-30B
ONPREM_LDFLAGS = -ldflags "-s -w -X main.version=$(VERSION)-onprem \
	-X 'github.com/kimjiwon/tgc/internal/config.DefaultBaseURL=$(ONPREM_URL)' \
	-X 'github.com/kimjiwon/tgc/internal/config.DefaultModel=$(ONPREM_MODEL_SUPER)' \
	-X 'github.com/kimjiwon/tgc/internal/config.DefaultDevModel=$(ONPREM_MODEL_DEV)' \
	-X 'github.com/kimjiwon/tgc/internal/config.ConfigDirName=.tgc-onprem' \
	-X 'github.com/kimjiwon/tgc/internal/config.DebugMode=true'"

# Gemma 4 build config — Novita endpoint, google/gemma-4-31b-it for
# both Super and Dev modes. Uses a separate config dir so it can live
# side-by-side with the default Novita install.
GEMMA_URL = https://api.novita.ai/openai
GEMMA_MODEL = google/gemma-4-31b-it
GEMMA_LDFLAGS = -ldflags "-s -w -X main.version=$(VERSION)-gemma \
	-X 'github.com/kimjiwon/tgc/internal/config.DefaultBaseURL=$(GEMMA_URL)' \
	-X 'github.com/kimjiwon/tgc/internal/config.DefaultModel=$(GEMMA_MODEL)' \
	-X 'github.com/kimjiwon/tgc/internal/config.DefaultDevModel=$(GEMMA_MODEL)' \
	-X 'github.com/kimjiwon/tgc/internal/config.ConfigDirName=.tgc-gemma' \
	-X 'github.com/kimjiwon/tgc/internal/config.DebugMode=true'"

.PHONY: build run clean build-all build-onprem build-gemma build-release install lint test build-index

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

build-gemma:
	mkdir -p dist
	GOOS=darwin GOARCH=arm64 go build $(GEMMA_LDFLAGS) -o dist/$(BINARY)-gemma-darwin-arm64 ./cmd/tgc
	GOOS=darwin GOARCH=amd64 go build $(GEMMA_LDFLAGS) -o dist/$(BINARY)-gemma-darwin-amd64 ./cmd/tgc
	GOOS=windows GOARCH=amd64 go build $(GEMMA_LDFLAGS) -o dist/$(BINARY)-gemma-windows-amd64.exe ./cmd/tgc
	GOOS=linux GOARCH=amd64 go build $(GEMMA_LDFLAGS) -o dist/$(BINARY)-gemma-linux-amd64 ./cmd/tgc
	GOOS=linux GOARCH=arm64 go build $(GEMMA_LDFLAGS) -o dist/$(BINARY)-gemma-linux-arm64 ./cmd/tgc

# One-shot release: clean dist, build the three variants sequentially.
build-release: clean build-index build-all build-onprem build-gemma
	@echo "=== release artifacts ==="
	@ls -lh dist/

lint:
	go vet ./...

test:
	go test ./...
