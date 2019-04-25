.PHONY: build coverage test

coverage:
	go test $(FLAGS) -coverprofile=/tmp/go-code-cover ./...

codecov:
	./scripts/codecov.sh

build:
	go build -o awsinventory ./cmd/awsinventory

test:
	go test $(FLAGS) ./...