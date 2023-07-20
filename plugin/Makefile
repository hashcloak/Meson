GOPATH=$(shell go env GOPATH)

.PHONY: default
default: lint test

.PHONY: lint
lint:
	go vet ./...
	go fmt ./...
	go mod tidy

.PHONY: test
test:
	go test --race ./...

.PHONY: build
build:
	go build ./cmd/meson-plugin
