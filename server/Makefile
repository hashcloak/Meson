GOPATH=$(shell go env GOPATH)
GOTAGS="badgerdb"

.PHONY: default
default: lint test

.PHONY: lint
lint:
	go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.42.0
	$(GOPATH)/bin/golangci-lint run -e gosec ./... --timeout 2m
	go fmt ./...
	go mod tidy

# added -race in future (badger fatal error: checkptr: pointer arithmetic result points to invalid allocation)
# https://github.com/golang/go/issues/40917
.PHONY: test
test:
	go test -tags=$(GOTAGS) ./...

.PHONY: build
build:
	go build -tags=$(GOTAGS) -o meson-server cmd/meson-server/*.go

.PHONY: build-docker
build-docker:
	docker build --no-cache -t meson/server .
