BIN := nkcli

VERSION := $(shell git describe --tags || git rev-parse --short HEAD)
BUILD_LDFLAGS := "-s -w -X main.version=$(VERSION)"

OSS := darwin linux windows
ARCHS := amd64 arm64

nkcli:
	go build -ldflags=$(BUILD_LDFLAGS) -o nkcli .

.PHONY: clean
clean:
	rm -rf dist
	go clean

.PHONY: cross
cross: 
	@$(foreach os,$(OSS),$(foreach arch,$(ARCHS), \
		mkdir -p dist/$(BIN)_$(VERSION)_$(os)_$(arch); \
 		GOOS=$(os) GOARCH=$(arch) go build -o dist/$(BIN)_$(VERSION)_$(os)_$(arch)/$(BIN) -ldflags $(BUILD_LDFLAGS) .; \
 		tar -cJf dist/$(BIN)_$(VERSION)_$(os)_$(arch).tar.xz -C dist $(BIN)_$(VERSION)_$(os)_$(arch)/; \
 		rm -rf dist/$(BIN)_$(VERSION)_$(os)_$(arch); \
	))
