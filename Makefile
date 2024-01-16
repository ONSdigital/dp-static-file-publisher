BINPATH ?= build

BUILD_TIME=$(shell date +%s)
GIT_COMMIT=$(shell git rev-parse HEAD)
VERSION ?= $(shell git tag --points-at HEAD | grep ^v | head -n 1)

LDFLAGS = -ldflags "-X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT) -X main.Version=$(VERSION)"

.PHONY: all
all: audit test build

.PHONY: audit
audit:
	go list -m all | nancy sleuth

.PHONY: build
build:
	go build -tags 'production' $(LDFLAGS) -o $(BINPATH)/dp-static-file-publisher

.PHONY: debug
debug:
	go build -tags 'debug' $(LDFLAGS) -o $(BINPATH)/dp-static-file-publisher
	HUMAN_LOG=1 DEBUG=1 $(BINPATH)/dp-static-file-publisher

.PHONY: test
test:
	go test -race -cover ./...

.PHONY: docker-test
docker-test-component:
	docker-compose  -f docker-compose-services.yml -f docker-compose.yml down
	docker build -f Dockerfile . -t template_test --target=test
	docker-compose  -f docker-compose-services.yml -f docker-compose.yml up -d
	docker-compose  -f docker-compose-services.yml -f docker-compose.yml exec -T dp-static-file-publisher go test -component
	docker-compose  -f docker-compose-services.yml -f docker-compose.yml down

docker-local:
	docker-compose -f docker-compose-services.yml -f docker-compose-local.yml down
	docker-compose -f docker-compose-services.yml -f docker-compose-local.yml up -d
	docker-compose -f docker-compose-services.yml -f docker-compose-local.yml exec dp-static-file-publisher bash

lint:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.52.2
	golangci-lint run ./... --timeout 10m --tests=false --skip-dirs=features