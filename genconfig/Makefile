GOPATH=$(shell go env GOPATH)

.PHONY: default
default: lint test

# Exclude S1034: assigning the result of this type assertion to a variable (switch cfg := cfg.(type)) could eliminate type assertions in switch cases
.PHONY: lint
lint:
	go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.37.0
	$(GOPATH)/bin/golangci-lint run -e gosec -e S1034 ./... --timeout 2m
	go fmt ./...
	go mod tidy

.PHONY: test
test:
	go test ./...

.PHONY: build
build:
	go build -o genconfig cmd/gen/main.go

.PHONY: build-updateconfig
build-updateconfig:
	go build -o updateconfig cmd/update/main.go
