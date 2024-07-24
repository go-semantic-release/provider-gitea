

all: build

.PHONY: build

build:
	go build -o build/provider-gitea ./cmd/provider-gitea/

lint:
	golangci-lint run --print-issued-lines=false --fix ./cmd/... ./pkg/...

test:
	go test --coverprofile coverage.out -json -v -parallel 20 ./tests 2>&1 | tee /tmp/gotest.log | gotestfmt
