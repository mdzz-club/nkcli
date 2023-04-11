BIN := nkcli

VERSION := $(shell git describe --tags || git rev-parse --short HEAD)
BUILD_LDFLAGS := "-s -w -X main.version=$(VERSION)"

OSS := darwin linux windows
ARCHS := amd64 arm64

nkcli:
	go build -ldflags=$(BUILD_LDFLAGS) -o $(BIN) .

.PHONY: clean
clean:
	rm -rf dist
	go clean

.PHONY: cross
cross: 
	@$(foreach os,$(OSS),$(foreach arch,$(ARCHS), \
		mkdir -p dist/$(BIN)_$(VERSION)_$(os)_$(arch); \
		if [[ $(os) == "windows" ]]; then \
	 		GOOS=$(os) GOARCH=$(arch) go build -o dist/$(BIN)_$(VERSION)_$(os)_$(arch)/$(BIN).exe -ldflags $(BUILD_LDFLAGS) .; \
	 		tar -czf dist/$(BIN)_$(VERSION)_$(os)_$(arch).tgz -C dist $(BIN)_$(VERSION)_$(os)_$(arch)/; \
		else \
	 		GOOS=$(os) GOARCH=$(arch) go build -o dist/$(BIN)_$(VERSION)_$(os)_$(arch)/$(BIN) -ldflags $(BUILD_LDFLAGS) .; \
	 		tar -cJf dist/$(BIN)_$(VERSION)_$(os)_$(arch).tar.xz -C dist $(BIN)_$(VERSION)_$(os)_$(arch)/; \
		fi; \
 		rm -rf dist/$(BIN)_$(VERSION)_$(os)_$(arch); \
	))
