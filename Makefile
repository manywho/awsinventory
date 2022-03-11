.PHONY: build ci release test test-full
.ONESHELL:
.SHELL: /bin/sh

BINARY=awsinventory
TARGETS=linux:amd64 linux:386 darwin:amd64 windows:amd64 windows:386
TARGETS=linux:amd64 linux:386 darwin:amd64

#TARGETS=darwin:amd64

BUILD_DIR=./cmd/awsinventory
VERSION?=0.1.0
BUILD_HASH?=$(shell git rev-parse --short HEAD)

LDFLAGS := -ldflags '-extldflags "-static" -X main.version=$(VERSION) -X main.build=$(BUILD_HASH)'

default: build

ver:
	echo $(VERSION)

build:
	go build $(LDFLAGS) -o awsinventory $(BUILD_DIR)

ci:
	test $$(gofmt -l . | wc -l) -eq 0
	golint -set_exit_status ./...
	go test -race -coverprofile=coverage.txt -covermode=atomic ./...

clean:
	rm -rf build/*

release:
	for target in ${TARGETS}; do

		os=$${target%:*}
		arch=$${target#*:}
		echo "okay"
		if [[ "$${os}/$${arch}" == "darwin/386" ]]; then continue; fi
		output_file="$(BINARY)-$(VERSION)-$${os}-$${arch}"
		echo "==> building $${output_file}"
		GOOS=$${os} GOARCH=$${arch} CGO_ENABLED=0 go build $(LDFLAGS) -o build/$${output_file} $(BUILD_DIR)
	done
	cd build
	for bin in *; do
		echo "==> generating checksum for $${bin}"
		sha256sum $${bin} > $${bin}.sha256
	done

test:
	go test $(FLAGS) ./...

test-full:
	go test $(FLAGS) -race -cover ./...
