.PHONY: build ci test test-full

test:
	go test $(FLAGS) ./...

test-full:
	go test $(FLAGS) -race -cover ./...

ci:
	./scripts/codecov.sh

build:
	go build -o awsinventory ./cmd/awsinventory