SHELL := /bin/sh

AIR ?= $(shell go env GOBIN)/air
SWAG ?= $(shell go env GOBIN)/swag

DATABASE_URL ?= postgres://postgres:password@localhost:5432/postgres
JWT_SECRET ?= jwt-secret-dev

export DATABASE_URL
export JWT_SECRET

$(AIR):
	go install github.com/air-verse/air@latest

$(SWAG):
	go install github.com/swaggo/swag/cmd/swag@latest

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
