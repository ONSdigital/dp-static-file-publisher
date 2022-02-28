BINPATH ?= build

VAULT_ADDR?='http://127.0.0.1:8200'

# The following variables are used to generate a vault token for the app. The reason for declaring variables, is that
# its difficult to move the token code in a Makefile action. Doing so makes the Makefile more difficult to
# read and starts introduction if/else statements.
#VAULT_POLICY:="$(shell vault policy write -address=$(VAULT_ADDR) read-psk policy.hcl)"
#TOKEN_INFO:="$(shell vault token create -address=$(VAULT_ADDR) -policy=read-psk -period=24h -display-name=dp-download-service)"
#APP_TOKEN:="$(shell echo $(TOKEN_INFO) | awk '{print $$6}')"

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
	HUMAN_LOG=1 DEBUG=1 VAULT_TOKEN=$(APP_TOKEN) VAULT_ADDR=$(VAULT_ADDR) $(BINPATH)/dp-static-file-publisher

.PHONY: test
test:
	go test -race -cover ./...

.PHONY: convey
convey:
	goconvey ./...

.PHONY: vault
vault:
	@echo "$(VAULT_POLICY)"
	@echo "$(TOKEN_INFO)"
	@echo "$(APP_TOKEN)"

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

