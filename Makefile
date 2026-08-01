BINARY  := cyberspace
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X main.version=$(VERSION)
PKG     := ./cmd/cyberspace
GO      ?= go

PLATFORMS := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64

.PHONY: all build install uninstall test vet release clean

all: build

# Build a local binary into ./bin
build:
	$(GO) build -trimpath -ldflags "$(LDFLAGS)" -o bin/$(BINARY) $(PKG)

# Install into ~/.local/bin (a common, PATH-visible location for user binaries)
install:
	@mkdir -p $(HOME)/.local/bin
	$(GO) build -trimpath -ldflags "$(LDFLAGS)" -o $(HOME)/.local/bin/$(BINARY) $(PKG)
	@echo "Installed $(BINARY) to $(HOME)/.local/bin/$(BINARY)"
	@echo "Make sure $(HOME)/.local/bin is on your PATH to run it as 'cyberspace'."

uninstall:
	@rm -f $(HOME)/.local/bin/$(BINARY)
	@echo "Removed $(HOME)/.local/bin/$(BINARY)."

test:
	$(GO) test ./...

vet:
	$(GO) vet ./...

# Cross-compile static binaries for every platform into ./dist as tarballs.
# Each tarball contains the bare binary name so install scripts can extract it
# without knowing the platform.
release:
	@mkdir -p dist
	@set -e; for platform in $(PLATFORMS); do \
		os=$${platform%/*}; arch=$${platform#*/}; \
		ext=""; [ "$$os" = "windows" ] && ext=".exe"; \
		archive="$(BINARY)-$$os-$$arch.tar.gz"; \
		echo "Building $$archive"; \
		GOOS=$$os GOARCH=$$arch CGO_ENABLED=0 $(GO) build -trimpath -ldflags "$(LDFLAGS)" -o dist/$(BINARY)$$ext $(PKG); \
		tar -czf dist/$$archive -C dist $(BINARY)$$ext; \
		rm dist/$(BINARY)$$ext; \
	done
	@echo "Release artifacts written to dist/"

clean:
	rm -rf bin dist
