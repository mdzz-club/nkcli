VERSION := $(shell git describe --tags || git rev-parse --short HEAD)
BUILD_LDFLAGS := "-s -w -X main.version=$(VERSION)"

nkcli:
	go build -ldflags=$(BUILD_LDFLAGS) -o nkcli ./main

.PHONY: clean
clean:
	rm ./nkcli
	go clean
