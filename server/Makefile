GOPATH=$(shell go env GOPATH)
GOTAGS="badgerdb"

.PHONY: default
default: lint test

.PHONY: lint
lint:
	go vet ./...
	go fmt ./...
	go mod tidy

.PHONY: test
test:
	go test --race -tags=$(GOTAGS) ./...

.PHONY: build
build:
	go build -tags=$(GOTAGS) -o meson-server cmd/meson-server/*.go