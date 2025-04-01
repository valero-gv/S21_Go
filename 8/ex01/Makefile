DISABLED_LINTERS="depguard"

all: build

build:
	find cmd -name "*.go" -print0 | xargs -0 -n1 -t go build -o bin/

fmt:
	go fmt ./...

lint:
	go vet ./...
	golangci-lint run --fix --default=all --disable=$(DISABLED_LINTERS)

generate:
	protoc --go_out=. --go_opt=paths=import \
		--go-grpc_out=. --go-grpc_opt=paths=import \
		spec/*.proto

.PHONY: all build fmt lint generate
