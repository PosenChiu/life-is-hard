SHELL := /bin/sh
NAMESPACE ?= life-is-hard

DB_SCHEME ?= postgres
DB_USERNAME ?= postgres
DB_PASSWORD ?= password
DB_HOST ?= localhost
DB_PORT ?= 55432
DB_NAME ?= postgres
DATABASE_URL ?= $(DB_SCHEME)://$(DB_USERNAME):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)

REDIS_PASSWORD ?= password
REDIS_HOST ?= localhost
REDIS_PORT ?= 6380
REDIS_ADDR ?= $(REDIS_HOST):$(REDIS_PORT)
REDIS_DB ?= 0

JWT_SECRET ?= jwt-secret-dev

export DATABASE_URL
export REDIS_ADDR
export REDIS_DB
export REDIS_PASSWORD
export JWT_SECRET

AIR ?= $(shell go env GOBIN)/air
SWAG ?= $(shell go env GOBIN)/swag

$(AIR):
	go install github.com/air-verse/air@latest

$(SWAG):
	go install github.com/swaggo/swag/cmd/swag@latest

.PHONY: kill
kill:
	@docker ps -qaf "name=^$(NAMESPACE)-" | xargs -r docker stop | xargs -r docker rm

.PHONY: db
db: kill
	@docker run -d \
		--env POSTGRES_PASSWORD=$(DB_PASSWORD) \
		--publish $(DB_PORT):5432 \
		--name $(NAMESPACE)-postgres postgres:latest
	@docker run -d \
		--publish $(REDIS_PORT):6379 \
		--name $(NAMESPACE)-redis redis:latest \
		redis-server --requirepass $(REDIS_PASSWORD)

.PHONY: dev
dev: $(AIR) $(SWAG)
	@$(AIR) \
		-build.exclude_dir "docs" \
		-build.cmd "\
			$(SWAG) init -g main.go -d cmd/service,internal/handler,internal/dto \
			&& go mod tidy \
			&& go fmt ./... \
			&& go vet ./... \
			&& go build -o ./tmp/main cmd/service/main.go \
			&& printf '# Created by Makefile automatically.\n*\n' | tee {docs,tmp}/.gitignore >/dev/null \
		"
