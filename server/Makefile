GOPATH=$(shell go env GOPATH)
GOTAGS="badgerdb"

.PHONY: default
default: lint test

.PHONY: lint
lint:
	# cd ../tools && go install github.com/golangci/golangci-lint/cmd/golangci-lint
	# $(GOPATH)/bin/golangci-lint run -e gosec ./... --timeout=2m
	cd ../ test && go install honnef.co/go/tools/cmd/staticcheck@v0.4.6
	$(GOPATH)/bin/staticcheck ./...
	go vet ./...
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