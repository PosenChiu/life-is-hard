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
