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
		-build.exclude_dir "docs,prompt,tmp" \
		-build.cmd "\
			$(SWAG) init -g main.go -d cmd/service,internal/dto,internal/handler \
			&& go mod tidy \
			&& go fmt ./... \
			&& go vet ./... \
			&& go build -o ./tmp/main cmd/service/main.go \
			&& printf '# Created by Makefile automatically.\n*\n' | tee {docs,tmp}/.gitignore >/dev/null \
			&& printf '\nOpen Swagger: \033[36mhttp://localhost:8888/swagger/index.html\033[0m\n\n' \
		"

.PHONY: test
test:
	@go test -coverprofile=coverage.out -coverpkg=./... ./...
	@go tool cover -html=coverage.out -o coverage.html
	@go tool cover -func=coverage.out

override rglob = \
  $(wildcard $(foreach p,$(2),$(1)/$(p))) \
  $(foreach d,$(filter-out $(1)/. $(1)/..,$(wildcard $(1)/* $(1)/.*)),$(call rglob,$d,$(2)))

PROMPT_DIRS := cmd internal
PROMPT_PATTERNS := *.go *.sql *.md
PROMPT_FILES := config.mk go.mod go.sum

define FORMAT_FILE_TO_MD
printf '## %s\n\n' "$(1)" >> "$(2)"
printf '````````````````' >> "$(2)"
printf '\n'               >> "$(2)"
cat "$(1)"                >> "$(2)"
printf '\n'               >> "$(2)"
printf '````````````````' >> "$(2)"
printf '\n'               >> "$(2)"
printf '\n'               >> "$(2)"
endef

.PHONY: prompt
prompt: $(PROMPT_FILES) $(sort $(foreach dir,$(PROMPT_DIRS),$(call rglob,$(dir),$(PROMPT_PATTERNS))))
	@printf '# Created by Makefile automatically.\n.gitignore\n' > $@/.gitignore
	@printf 'CODE.md\n' >> $@/.gitignore
	@printf '# Aggregated Code\n\n' > $@/CODE.md
	@$(foreach f,$^,$(call FORMAT_FILE_TO_MD,$f,$@/code.md);)
	@printf 'LAYOUT.md\n' >> $@/.gitignore
	@printf '# Directory Layout\n\n' > $@/LAYOUT.md
	@$(foreach f,$^,printf '%s %s\n' "##" "$f" >> $@/LAYOUT.md;)
	@printf 'Prompt: \033[36mRead the README.md LAYOUT.md CODE.md I provided before typing my question.\033[0m\n'
