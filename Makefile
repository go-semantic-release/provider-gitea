

all: build

.PHONY: build

build:
	go build -o build/provider-gitea ./cmd/provider-gitea/

lint:
	golangci-lint run --print-issued-lines=false --fix ./...

test:
	go test --coverprofile coverage.out -v -parallel 20 ./...
