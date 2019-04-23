.PHONY: build

test:
	go test $(FLAGS) -coverprofile=/tmp/go-code-cover ./...

build:
	go build -o awsinventory ./cmd/awsinventory