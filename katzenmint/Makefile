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
	# cd ../tools && go install github.com/golangci/golangci-lint/cmd/golangci-lint
	# $(GOPATH)/bin/golangci-lint run -e gosec ./... --timeout=2m
	# cd ../ test && go install honnef.co/go/tools/cmd/staticcheck@v0.4.6
	# $(GOPATH)/bin/staticcheck ./...
	go vet ./...
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
	go build -tags=$(GOTAGS) -ldflags '$(LD_FLAGS)' -o katzenmint cmd/katzenmint/*.go