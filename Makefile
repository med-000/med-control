SHELL := /bin/sh

GO_CACHE ?= /tmp/overview-go-build-cache
GO_ENV := GOCACHE=$(GO_CACHE)

.PHONY: help
help:
	@printf '%s\n' 'Available targets:'
	@printf '%s\n' '  make backend       Run backend server'
	@printf '%s\n' '  make infra         Run Notion sync worker'
	@printf '%s\n' '  make infra-raw     Print raw Notion rows'
	@printf '%s\n' '  make infra-tasks   Print mapped tasks sent to backend'
	@printf '%s\n' '  make test          Run all Go tests'
	@printf '%s\n' '  make docker-build  Build Docker images'
	@printf '%s\n' '  make up            Start Docker Compose'
	@printf '%s\n' '  make down          Stop Docker Compose'

.PHONY: backend
backend:
	cd backend && $(GO_ENV) go run ./cmd/app

.PHONY: infra
infra:
	cd infra && $(GO_ENV) go run ./cmd/app

.PHONY: infra-raw
infra-raw:
	cd infra && $(GO_ENV) go run ./cmd/raw

.PHONY: infra-tasks
infra-tasks:
	cd infra && $(GO_ENV) go run ./cmd/tasks

.PHONY: test
test:
	cd shared && $(GO_ENV) go test ./...
	cd infra && $(GO_ENV) go test ./...
	cd backend && $(GO_ENV) go test ./...

.PHONY: docker-build
docker-build:
	docker compose build

.PHONY: up
up:
	docker compose up --build

.PHONY: down
down:
	docker compose down
