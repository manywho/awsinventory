.PHONY: build ci release test test-full

BINARY=awsinventory
PLATFORMS=linux darwin windows
ARCHITECTURES=amd64 386

BUILD_DIR=./cmd/awsinventory
VERSION?=0.1.0
BUILD_HASH?=$(shell git rev-parse --short HEAD)

LDFLAGS := -ldflags '-extldflags "-static" -X main.version=$(VERSION) -X main.build=$(BUILD_HASH)'

default: test

ver:
	echo $(VERSION)

build:
	go build $(LDFLAGS) -o awsinventory $(BUILD_DIR)

ci:
	test $$(gofmt -l . | wc -l) -eq 0
	golint -set_exit_status ./...
	go test -race -coverprofile=coverage.txt -covermode=atomic ./...

clean:
	rm build/awsinventory-*

release:
	$(foreach GOOS, $(PLATFORMS),\
	$(foreach GOARCH, $(ARCHITECTURES), $(shell export GOOS=$(GOOS); export GOARCH=$(GOARCH); CGO_ENABLED=0 go build $(LDFLAGS) -o build/$(BINARY)-$(VERSION)-$(GOOS)-$(GOARCH) $(BUILD_DIR))))
	@echo releases built in the ./build directory

test:
	go test $(FLAGS) ./...

test-full:
	go test $(FLAGS) -race -cover ./...
