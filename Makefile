GO ?= go
DIST_DIR ?= dist
GOOS ?= $(shell $(GO) env GOOS)
GOARCH ?= $(shell $(GO) env GOARCH)
PLATFORMS := linux/386 linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/386 windows/amd64 windows/arm64

.PHONY: build build-all test

# Build the host target, or override GOOS/GOARCH for one cross-compiled target.
build:
	@mkdir -p "$(DIST_DIR)"
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) $(GO) build -trimpath -o "$(DIST_DIR)/godap-$(GOOS)-$(GOARCH)$(if $(filter windows,$(GOOS)),.exe)" .

# Cross-compile every supported Linux, macOS and Windows target.
build-all:
	@mkdir -p "$(DIST_DIR)"
	@set -e; for target in $(PLATFORMS); do \
		os=$${target%/*}; arch=$${target#*/}; ext=; \
		if [ "$$os" = windows ]; then ext=.exe; fi; \
		printf 'Building %s/%s\n' "$$os" "$$arch"; \
		CGO_ENABLED=0 GOOS=$$os GOARCH=$$arch $(GO) build -trimpath -o "$(DIST_DIR)/godap-$$os-$$arch$$ext" .; \
	done

test:
	$(GO) test ./...
