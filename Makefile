.PHONY: help setup setup-local setup-docker env-local env-docker db-init db-reset migrate-up migrate-status migrate-version migrate-down up up-db up-full down down-volumes logs ps build test lint clean install-docker

COMPOSE := docker compose --profile full
COMPOSE_INFRA := docker compose --profile infra

help: ## Show available commands
	@grep -E '^[a-zA-Z0-9_-]+:.*##' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

setup: setup-local ## Default: local MySQL + env + migrations

setup-local: env-local db-init ## Bootstrap local dev (Homebrew MySQL)
	@echo "Local setup complete. Run services with: make run-control (or use VS Code launch configs)"

setup-docker: env-docker install-docker ## Bootstrap Docker-based dev
	@echo "Run: make up-full"

env-local: ## Create .env for local MySQL from template
	@test -f .env || cp .env.local.example .env
	@echo "Using .env (local MySQL on localhost)"

env-docker: ## Create .env for Docker Compose from template
	@cp .env.docker.example .env
	@echo "Using .env (Docker service hostnames)"

db-init: ## Apply pending migrations (local Homebrew MySQL)
	@./scripts/init-db.sh local

db-init-docker: ## Apply pending migrations (Docker MySQL — db must be running)
	@docker compose --profile infra run --rm migrate up

db-reset: ## Wipe Docker DB volume and re-apply all migrations
	@echo "WARNING: This deletes all Docker MySQL data"
	@docker compose --profile full --profile infra down -v
	@$(MAKE) up-db
	@sleep 5
	@docker compose --profile infra run --rm migrate up
	@echo "Database reset complete"

migrate-up: ## Apply pending migrations
	@docker compose --profile infra run --rm migrate up

migrate-status: ## Show applied migration status
	@docker compose --profile infra run --rm migrate status

migrate-version: ## Print current migration version
	@docker compose --profile infra run --rm migrate version

migrate-down: ## Roll back the last migration
	@docker compose --profile infra run --rm migrate down

up-db: ## Start MySQL in Docker and run migrations
	$(COMPOSE_INFRA) up db -d
	@sleep 3
	$(COMPOSE_INFRA) run --rm migrate up

up-full: ## Start full stack (uses existing images; no registry pull)
	$(COMPOSE) up db -d
	@sleep 3
	$(COMPOSE) run --rm migrate up
	$(COMPOSE) up -d control market gateway

up-full-build: ## Rebuild images then start (use when code changed)
	DOCKER_BUILDKIT=1 $(COMPOSE) build --pull=false
	@$(MAKE) up-full

up: up-full ## Alias for full stack

down: ## Stop containers (data persists in spice_ledger_mysql_data volume)
	docker compose --profile full --profile infra down

down-volumes: ## Stop containers AND delete persistent DB data
	docker compose --profile full --profile infra down -v

logs: ## Tail logs for all services
	$(COMPOSE) logs -f

ps: ## Show running containers
	docker compose ps -a

build: ## Build all Docker images
	$(COMPOSE) build

test: ## Run Go tests
	go test ./...

lint: ## Verify module builds
	go build ./...

clean: ## Remove build artifacts
	go clean -cache
	rm -f control-server market-server gateway-server migrate

install-docker: ## Install Docker Desktop (macOS)
	@command -v docker >/dev/null 2>&1 && echo "Docker CLI already available" || (brew install --cask docker && echo "Open Docker Desktop from Applications, then re-run make up-full")

run-control: ## Run control gRPC service locally
	go run ./control/cmd/control/main.go

run-market: ## Run market gRPC service locally
	go run ./market/cmd/market/main.go

run-gateway: ## Run unified API gateway locally (REST + GraphQL on :8080)
	go run ./gateway/cmd/gateway/main.go
