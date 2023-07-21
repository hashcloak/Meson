GOPATH=$(shell go env GOPATH)
GOTAGS="badgerdb"
VERSION:=$(shell git describe --tags 2>/dev/null)
BUILDTIME=$(shell date +"%d-%b-%Y")
LD_FLAGS=
ifneq ($(VERSION), )
LD_FLAGS:= -X "github.com/hashcloak/Meson/katzenmint.protocolVersion=$(VERSION)"
endif

ifneq ($(BUILDTIME), )
LD_FLAGS:= $(LD_FLAGS) -X "github.com/hashcloak/Meson/katzenmint.buildTime=${BUILDTIME}"
endif

.PHONY: default
default: lint test

.PHONY: lint
lint:
	go vet ./...
	go fmt ./...
	go mod tidy

.PHONY: test
test:
	go test -race ./...


.PHONY: setup
setup:
	sh setup.sh

.PHONY: build
build:
	go build -tags=$(GOTAGS) -ldflags '$(LD_FLAGS)' -o katzenmint cmd/katzenmint/*.go