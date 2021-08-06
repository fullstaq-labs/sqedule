GOPATH          	?= $(shell go env GOPATH)
GOBIN           	?= $(firstword $(subst :, ,${GOPATH}))/bin
NPM_EXECUTABLE  	?= $(shell which npm)
NPM_TEST        	= $(shell command -v $(NPM_EXECUTABLE))
DOCKER_COMPOSE  	?= $(shell which docker-compose)
DOCKER_COMPOSE_TEST = $(shell command -v $(DOCKER_COMPOSE))
POSTGRES_USER		?=postgres
POSTGRES_PASSWORD 	?=changeme
POSTGRES_HOST 		?=localhost
POSTGRES_PORT 		?=5432
DB_TYPE				?=postgresql
DB_NAME				?=sqedule_dev
ORG					?=1

define check_npm
	@if [ ! $(NPM_TEST) ]; then \
		echo "Cannot find or execute NPM binary $(NPM_EXECUTABLE), you can override it by setting the NPM_EXECUTABLE env variable"; \
		exit 1; \
	fi
endef

define check_docker_compose
	@if [ ! $(DOCKER_COMPOSE_TEST) ]; then \
		echo "Cannot find or execute docker-compose binary $(DOCKER_COMPOSE), you can override it by setting the DOCKER_COMPOSE env variable"; \
		exit 1; \
	fi
endef

define make_env
	echo POSTGRES_USER=$(POSTGRES_USER) >> devtools/.env
	echo POSTGRES_PASSWORD=$(POSTGRES_PASSWORD) >> devtools/.env
	echo POSTGRES_HOST=$(POSTGRES_HOST) >> devtools/.env
	echo POSTGRES_PORT=$(POSTGRES_PORT) >> devtools/.env
	echo DB_TYPE=$(DB_TYPE) >> devtools/.env
	echo DB_NAME=$(DB_NAME) >> devtools/.env
endef

help: ## Displays help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} /^[a-z0-9A-Z_-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

.PHONY: deps
deps: ## Ensures fresh go.mod and go.sum.
	@go mod tidy
	@go mod verify

.PHONY: check-npm
check-npm: ## Check if npm is present
	$(call check_npm,'run make docs and commit changes')

.PHONY: local-web-setup
web-setup: ## One time web-setup
	@echo ">> checking npm"
	$(MAKE) check-npm
	@echo ">> npm install"
	@npm install --prefix webui
	@echo ">> Done web-setup"

.PHONY: local-web-start
web-start: ## Start web UI
	@npm run dev --prefix webui

.PHONY: local-server-start
local-server-start: ## Start server
	@go run ./cmd/sqedule-server server --db-type $(DB_TYPE) --db-connection 'dbname=$(DB_NAME) user=$(POSTGRES_USER) password=$(POSTGRES_PASSWORD) host=$(POSTGRES_HOST) port=$(POSTGRES_PORT)'

.PHONY: local-start-all
start-all: ## Start the webUI and Server
	($(MAKE) local-server-start)&
	($(MAKE) local-web-start)

.PHONY: local-db-create
local-db-create: ## Creates default database
	PGPASSWORD=$(POSTGRES_PASSWORD)  psql -h $(POSTGRES_HOST) -p $(POSTGRES_PORT) -U $(POSTGRES_USER) -c 'CREATE DATABASE $(DB_NAME)'

.PHONY: local-db-seed
local-db-seed: ## Seeds the database
	PGPASSWORD=$(POSTGRES_PASSWORD)  psql -h $(POSTGRES_HOST) -p $(POSTGRES_PORT) -U $(POSTGRES_USER) dbname=$(DB_NAME) -f devtools/db-seed.sql

.PHONY: local-db-migrate
local-db-migrate: ## Migrate database
	@go run ./cmd/sqedule-server db migrate --db-type $(DB_TYPE) --db-connection 'dbname=$(DB_NAME) user=$(POSTGRES_USER) password=$(POSTGRES_PASSWORD) host=$(POSTGRES_HOST) port=$(POSTGRES_PORT)'

.PHONY: local-db-reset
local-db-reset: ## Reset database
	@go run ./cmd/sqedule-server db migrate --db-type $(DB_TYPE) --db-connection 'dbname=$(DB_NAME) user=$(POSTGRES_USER) password=$(POSTGRES_PASSWORD) host=$(POSTGRES_HOST) port=$(POSTGRES_PORT)'	 --reset

.PHONY: local-db-init
local-db-init: ## Do all DB init steps, requires database
	$(MAKE) local-db-create
	$(MAKE) local-db-migrate
	$(MAKE) local-db-seed

.PHONY: docker-db-create
docker-db-create: ## Creates default database
	@docker run --network=devtools_sqedule -e PGPASSWORD=$(POSTGRES_PASSWORD)  postgres:latest psql -h postgres -p $(POSTGRES_PORT) -U $(POSTGRES_USER) -c 'CREATE DATABASE $(DB_NAME)'

.PHONY: docker-db-seed
docker-db-seed: ## Seeds the database
	@docker run --network=devtools_sqedule -v $(shell pwd)/devtools:/tmp/ -e PGPASSWORD=$(POSTGRES_PASSWORD)  postgres:latest psql -h postgres -p $(POSTGRES_PORT) -U $(POSTGRES_USER) dbname=$(DB_NAME) -f tmp/db-seed.sql

.PHONY: docker-db-migrate
docker-db-migrate: ## Migrate database via docker
	@docker run --network=devtools_sqedule sqedule:latest db migrate --db-type $(DB_TYPE) --db-connection 'dbname=$(DB_NAME) user=$(POSTGRES_USER) password=$(POSTGRES_PASSWORD) host=postgres port=$(POSTGRES_PORT)'

.PHONY: docker-database
docker-database: ## Start a postgres & pgadmin docker env
	@docker-compose -f devtools/docker-compose-db.yml up -d

.PHONY: docker-database-down
docker-database-down: ## Stop the postgres & pgadmin docker env
	@docker-compose -f devtools/docker-compose-db.yml down

.PHONY: docker-build-sqedule
docker-build-sqedule: ## Build docker image for sqedule
	@docker build -t sqedule .

.PHONY: docker-build-webui
docker-build-webui: ## Build docker image for webUI
	@docker build -t webui webui

.PHONY: docker-full
docker-full: ## Start a postgres, pgadmin, webUI & sqedule server docker env
	$(call make_env,'make env var file')
	@docker-compose --project-directory devtools -f devtools/docker-compose-full.yml --env-file devtools/.env up -d
	@rm devtools/.env

.PHONY: docker-full-down
docker-full-down: ## Stop the postgres, pgadmin, webUI & sqedule server docker env
	@docker-compose -f devtools/docker-compose-full.yml down

.PHONY: docker-db-init
docker-db-init: ## Do all DB init steps, requires either docker-full or docker-database
	$(MAKE) docker-db-create
	$(MAKE) docker-db-migrate
	$(MAKE) docker-db-seed

.PHONY: docker-full-init
docker-full-init: ## Build docker containers, start all applications and prepare database
	$(MAKE) docker-build-webui
	$(MAKE) docker-build-sqedule
	$(MAKE) docker-full
	$(MAKE) docker-db-init
