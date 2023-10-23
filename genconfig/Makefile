GOPATH=$(shell go env GOPATH)

.PHONY: default
default: lint test

# Exclude S1034: assigning the result of this type assertion to a variable (switch cfg := cfg.(type)) could eliminate type assertions in switch cases
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
	go test ./...

.PHONY: build
build:
	go build -o genconfig cmd/gen/main.go

.PHONY: build-updateconfig
build-updateconfig:
	go build -o updateconfig cmd/update/main.go
