GOPATH=$(shell go env GOPATH)
GOTAGS="badgerdb"
VERSION:=$(shell git describe --tags 2>/dev/null)
BUILDTIME=$(shell date +"%d %b %Y")
LD_FLAGS=
ifneq ($(VERSION), )
LD_FLAGS:= -X "github.com/hashcloak/katzenmint-pki.protocolVersion=$(VERSION)"
endif

ifneq ($(BUILDTIME), )
LD_FLAGS:= $(LD_FLAGS) -X "github.com/hashcloak/katzenmint-pki.buildTime=${BUILDTIME}"
endif

.PHONY: default
default: lint test

.PHONY: lint
lint:
	go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.42.0
	$(GOPATH)/bin/golangci-lint run --timeout 2m0s -e gosec ./...
	go fmt ./...
	go mod tidy

# added -race in future (badger fatal error: checkptr: pointer arithmetic result points to invalid allocation)
# https://github.com/golang/go/issues/40917
.PHONY: test
test:
	go test ./s11n
	go test ./


.PHONY: setup
setup:
	sh setup.sh

.PHONY: build
build:
	go build -tags=$(GOTAGS) -ldflags '$(LD_FLAGS)' -o katzenmint cmd/katzenmint/katzenmint.go

.PHONY: docker-build
docker-build:
	docker build --no-cache -t katzenmint/pki .
