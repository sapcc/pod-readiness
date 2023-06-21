IMAGE   ?= keppel.eu-de-1.cloud.sap/ccloud/pod-readiness
VERSION = $(shell git rev-parse --verify HEAD | head -c 8)

GOOS    ?= $(shell go env | grep GOOS | cut -d'"' -f2)
BINARIES := pod_readiness

LDFLAGS := -X github.com/sapcc/pod-readiness/pkg/pod-readiness.VERSION=$(VERSION)
GOFLAGS := -ldflags "$(LDFLAGS)"

SRCDIRS  := cmd pkg internal
PACKAGES := $(shell find $(SRCDIRS) -type d)
GOFILES  := $(addsuffix /*.go,$(PACKAGES))
GOFILES  := $(wildcard $(GOFILES))


all: $(BINARIES:%=bin/$(GOOS)/%)

bin/%: $(GOFILES) Makefile
	GOOS=$(*D) GOARCH=amd64 go build $(GOFLAGS) -v -o $(@D)/$(@F) ./cmd/

build:
	docker build -t $(IMAGE):$(VERSION) .
	docker tag $(IMAGE):$(VERSION) $(IMAGE):latest

push: build
	docker push $(IMAGE):$(VERSION)
	docker push $(IMAGE):latest

clean:
	rm -rf bin/*

vendor:
	GO111MODULE=on go get -u ./... && go mod tidy && go mod vendor
