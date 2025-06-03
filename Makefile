include config.mk

AIR ?= $(shell go env GOBIN)/air
SWAG ?= $(shell go env GOBIN)/swag

$(AIR):
	go install github.com/air-verse/air@latest

$(SWAG):
	go install github.com/swaggo/swag/cmd/swag@latest

.PHONY: clean
clean: kill
	@docker volume ls -qf "name=^$(NAMESPACE)-" | xargs -r docker volume rm

.PHONY: kill
kill:
	@docker ps -qaf "name=^$(NAMESPACE)-" | xargs -r docker stop | xargs -r docker rm

.PHONY: run
run: kill
	@docker run -d \
		--env POSTGRES_PASSWORD=$(DB_PASSWORD) \
		--name $(NAMESPACE)-postgres \
		--publish $(DB_PORT):5432 \
		--volume $(NAMESPACE)-postgres:/var/lib/postgresql/data \
		postgres:latest
	@docker run -d \
		--name $(NAMESPACE)-redis \
		--publish $(REDIS_PORT):6379 \
		--volume $(NAMESPACE)-redis:/data \
		redis:latest \
		redis-server --requirepass $(REDIS_PASSWORD)

.PHONY: init
init: $(SWAG)
	@$(SWAG) init -g service.go -d cmd/service,internal/api,internal/handler
	@go mod tidy
	@go fmt ./...
	@go vet ./...
	@go build -o ./tmp/main cmd/service/service.go
	@printf "# Created by Makefile automatically.\n*\n" | tee {docs,tmp}/.gitignore >/dev/null

.PHONY: dev
dev: $(AIR) $(SWAG)
	@$(AIR) \
		-build.bin "./tmp/main" \
		-build.exclude_dir "docs,tmp" \
		-build.cmd "$(MAKE) init && printf '\nOpen Swagger: \033[36mhttp://localhost:8080/swagger/index.html\033[0m\n\n'"

.PHONY: test
test:
	@go test -coverprofile=coverage.out -coverpkg=./... ./...
	@go tool cover -func=coverage.out
