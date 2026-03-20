.PHONY: build test deb deb-all archives clean

BINARY ?= sshub-mcp
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null | sed 's/^v//' || echo 0.0.0-dev)

build:
	CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o $(BINARY) ./cmd/sshub-mcp

test:
	go test ./...

# Tarballs: linux/darwin × amd64/arm64
archives:
	./packaging/scripts/build-release-archives.sh

# Requires: Linux binary in dist/sshub-mcp_linux_amd64 (or build via archives first)
deb-amd64: archives
	./packaging/scripts/build-deb.sh "$(VERSION)" amd64 dist/sshub-mcp_linux_amd64

deb-arm64: archives
	./packaging/scripts/build-deb.sh "$(VERSION)" arm64 dist/sshub-mcp_linux_arm64

deb-all: archives deb-amd64 deb-arm64

clean:
	rm -f $(BINARY) dist/sshub-mcp_* dist/*.deb
