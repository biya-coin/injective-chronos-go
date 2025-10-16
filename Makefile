# Project settings
APP ?= chronos-api
BIN ?= bin/$(APP)
MAIN_PKG ?= .
PKG ?= ./...

# Build metadata
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

# Go build flags
GOFLAGS ?= -trimpath
LDFLAGS ?= -s -w

# Docker settings
IMAGE ?= $(APP)
TAG ?= $(VERSION)
DOCKERFILE ?= Dockerfile
PLATFORMS ?= linux/amd64,linux/arm64

.PHONY: help tidy fmt vet test build clean run image image-push image-multi ci

help:
	@echo "Available targets:"
	@echo "  tidy          - go mod tidy"
	@echo "  fmt           - gofmt -s -w ."
	@echo "  vet           - go vet ./..."
	@echo "  test          - go test ./..."
	@echo "  build         - build binary to $(BIN)"
	@echo "  run           - go run . (needs main package present)"
	@echo "  clean         - remove bin/"
	@echo "  image         - docker build $(IMAGE):$(TAG)"
	@echo "  image-push    - docker push $(IMAGE):$(TAG)"
	@echo "  image-multi   - docker buildx build --platform $(PLATFORMS) --push"
	@echo "  ci            - tidy, fmt, vet, test, build"

tidy:
	go mod tidy

fmt:
	gofmt -s -w .

vet:
	go vet $(PKG)

test:
	go test $(PKG)

build:
	@mkdir -p bin
	go build $(GOFLAGS) -ldflags '$(LDFLAGS)' -o $(BIN) $(MAIN_PKG)

run:
	go run .

clean:
	rm -rf bin

image:
	docker build -t $(IMAGE):$(TAG) -f deploy/$(DOCKERFILE) .

image-push:
	docker push $(IMAGE):$(TAG)

image-multi:
	docker buildx build --platform $(PLATFORMS) -t $(IMAGE):$(TAG) -f $(DOCKERFILE) --push .

ci: tidy fmt vet test build


docker-run-dev:
	@sudo mkdir -p /data/chronos-api-dev/log && sudo chown -R 65532:65532 /data/chronos-api-dev/log
	@sudo mkdir -p /data/chronos-api-dev/etc && sudo chown -R 65532:65532 /data/chronos-api-dev/etc
	docker run -p 4442:4442 --restart=always --name=chronos-api-dev \
	--network=middleware-net \
	-e ENV=dev \
	-v /etc/localtime:/etc/localtime:ro \
	-v /etc/timezone:/etc/timezone:ro \
	-v /data/chronos-api-dev/log:/app/log \
	-v /data/chronos-api-dev/etc:/app/etc \
	chronos-api:$(TAG) 