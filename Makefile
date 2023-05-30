VERSION ?= local
BUILDTIME := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')

GOLDFLAGS += -X main.Version=$(VERSION)
GOLDFLAGS += -X main.Buildtime=$(BUILDTIME)
GOFLAGS = -ldflags "$(GOLDFLAGS)"

PKG_LIST := $(shell go list ./... | grep -v /vendor/)

.PHONY: all build clean test test-unit test-race test-msan staticcheck vet

all: build

build:
	go build -race -o ./bin/lacuna $(GOFLAGS) .

staticcheck:
	staticcheck ${PKG_LIST}

vet:
	go vet ${PKG_LIST}

test: test-unit test-race

test-unit:
	go test -covermode=count -coverprofile=coverage.out ${PKG_LIST}

test-race:
	go test -race ${PKG_LIST}

test-msan:
	go test -msan ${PKG_LIST}

install:
	go install

run:
	go run main.go daemon -vvv

