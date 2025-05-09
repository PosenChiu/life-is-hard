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

.PHONY: dev
dev: $(AIR) $(SWAG)
	@$(AIR) \
		-build.exclude_dir "docs" \
		-build.cmd "\
			$(SWAG) init -g main.go -d cmd/service,internal/dto,internal/handler \
			&& go mod tidy \
			&& go fmt ./... \
			&& go vet ./... \
			&& go build -o ./tmp/main cmd/service/main.go \
			&& printf '# Created by Makefile automatically.\n*\n' | tee {docs,tmp}/.gitignore >/dev/null \
		"

.PHONY: swagger
swagger:
	open http://localhost:8080/swagger/index.html
