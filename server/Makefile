GOPATH=$(shell go env GOPATH)
GOTAGS="badgerdb"

.PHONY: default
default: lint test

.PHONY: lint
lint:
	# cd ../tools && go install github.com/golangci/golangci-lint/cmd/golangci-lint
	# $(GOPATH)/bin/golangci-lint run -e gosec ./... --timeout=2m
	# cd ../ test && go install honnef.co/go/tools/cmd/staticcheck@v0.4.6
	# $(GOPATH)/bin/staticcheck ./...
	go vet ./...
	go fmt ./...
	go mod tidy

.PHONY: test
test:
	go test -race -tags=$(GOTAGS) ./...

.PHONY: build
build:
	go build -tags=$(GOTAGS) -o meson-server cmd/meson-server/*.go