SHELL := /bin/bash
.ONESHELL:

APP_NAME ?= openproxy
BIN_DIR ?= ../bin

BUILD_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo dev)
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -X github.com/fdddf/openproxy/pkg/version.Commit=$(BUILD_COMMIT) -X github.com/fdddf/openproxy/pkg/version.BuildDate=$(BUILD_DATE)

ENUM_SOURCES := $(filter-out %_enum.go,$(wildcard consts/*.go))

.PHONY: all
all: build

.PHONY: build
build: ## Build the Go application binary into ../bin
	@mkdir -p $(BIN_DIR)
	go build -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/$(APP_NAME) .

.PHONY: migrate
migrate: ## Run database migrations
	@echo "Running database migrations..."
# 	GRANT ALL PRIVILEGES ON DATABASE gptproxy TO gptproxy;
# 	\c gptproxy;
#	GRANT ALL PRIVILEGES ON SCHEMA public TO gptproxy;
	go run . migrate


.PHONY: dao
dao: ## Generate DAO
	@echo "Generating DAO..."
	go run . gendao

.PHONY: enum
enum: ## Run go-enum across consts/*.go (excluding generated files)
	@if [ -z "$(ENUM_SOURCES)" ]; then \
		echo "No enum sources found under consts/"; \
	else \
		for f in $(ENUM_SOURCES); do \
			echo "Generating enums from $$f"; \
			go-enum --names --values -f $$f; \
		done; \
	fi

.PHONY: sql
sql: ## Generate SQL
	@echo "Generating SQL..."
	migrate create -ext sql -format unix -dir ./migrations/ $(NAME)

.PHONY: help
help: ## Show this help message
	@echo "Usage: make [target]"
	@echo ""
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
	
