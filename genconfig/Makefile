GOPATH=$(shell go env GOPATH)

.PHONY: default
default: lint test

# Exclude S1034: assigning the result of this type assertion to a variable (switch cfg := cfg.(type)) could eliminate type assertions in switch cases
.PHONY: lint
lint:
	go vet ./...
	go fmt ./...
	go mod tidy

.PHONY: test
test:
	go test -race ./...

.PHONY: build
build:
	go build -o genconfig cmd/gen/main.go

.PHONY: build-updateconfig
build-updateconfig:
	go build -o updateconfig cmd/update/main.go
