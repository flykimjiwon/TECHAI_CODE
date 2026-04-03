VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS = -ldflags "-s -w -X main.version=$(VERSION)"
BINARY = techai

.PHONY: build run clean build-all install

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

lint:
	go vet ./...

test:
	go test ./...
