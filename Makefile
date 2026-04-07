VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS = -ldflags "-s -w -X main.version=$(VERSION)"
BINARY = techai

# On-premise build config
ONPREM_URL = https://techai-web-prod.shinhan.com/v1
ONPREM_MODEL = GPT-OSS-120B
ONPREM_LDFLAGS = -ldflags "-s -w -X main.version=$(VERSION)-onprem \
	-X 'github.com/kimjiwon/tgc/internal/config.DefaultBaseURL=$(ONPREM_URL)' \
	-X 'github.com/kimjiwon/tgc/internal/config.DefaultModel=$(ONPREM_MODEL)' \
	-X 'github.com/kimjiwon/tgc/internal/config.DefaultDevModel=$(ONPREM_MODEL)' \
	-X 'github.com/kimjiwon/tgc/internal/config.ConfigDirName=.tgc-onprem'"

# Debug build config (DebugMode=true, keeps debug symbols)
DEBUG_LDFLAGS = -ldflags "-X main.version=$(VERSION)-debug \
	-X 'github.com/kimjiwon/tgc/internal/config.DebugMode=true'"

DEBUG_ONPREM_LDFLAGS = -ldflags "-X main.version=$(VERSION)-debug-onprem \
	-X 'github.com/kimjiwon/tgc/internal/config.DefaultBaseURL=$(ONPREM_URL)' \
	-X 'github.com/kimjiwon/tgc/internal/config.DefaultModel=$(ONPREM_MODEL)' \
	-X 'github.com/kimjiwon/tgc/internal/config.DefaultDevModel=$(ONPREM_MODEL)' \
	-X 'github.com/kimjiwon/tgc/internal/config.ConfigDirName=.tgc-onprem' \
	-X 'github.com/kimjiwon/tgc/internal/config.DebugMode=true'"

.PHONY: build run clean build-all build-onprem build-debug build-debug-onprem install

build:
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

build-debug:
	mkdir -p dist
	go build $(DEBUG_LDFLAGS) -o dist/$(BINARY)-debug ./cmd/tgc

build-debug-onprem:
	mkdir -p dist
	GOOS=windows GOARCH=amd64 go build $(DEBUG_ONPREM_LDFLAGS) -o dist/$(BINARY)-debug-onprem-windows-amd64.exe ./cmd/tgc

lint:
	go vet ./...

test:
	go test ./...
