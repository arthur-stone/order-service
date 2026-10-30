APP_NAME ?= order-service
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "1.0.0")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
DOCKER_IMAGE ?= $(APP_NAME):$(VERSION)

LDFLAGS = -s -w \
	-X main.version=$(VERSION) \
	-X main.commit=$(COMMIT) \
	-X main.buildDate=$(BUILD_DATE)

.PHONY: all fmt lint test build run docker-build docker-run clean

all: fmt lint test build

fmt:
	@echo "==> Formatting code..."
	go fmt ./...

lint:
	@echo "==> Running linters..."
	golangci-lint run ./...

test:
	@echo "==> Running tests with race detector and coverage..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out | tail -n 1

build:
	@echo "==> Building binary [$(APP_NAME)]..."
	CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o bin/$(APP_NAME) ./cmd/server

run:
	@echo "==> Running application..."
	go run -ldflags="$(LDFLAGS)" ./cmd/server

docker-build:
	@echo "==> Building Docker image [$(DOCKER_IMAGE)]..."
	docker build -t $(DOCKER_IMAGE) -t $(APP_NAME):latest .

docker-run:
	@echo "==> Running Docker container..."
	docker run --rm -p 8080:8080 --name $(APP_NAME) $(DOCKER_IMAGE)

clean:
	@echo "==> Cleaning up build artifacts..."
	rm -rf bin/ coverage.out
