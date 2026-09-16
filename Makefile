.DEFAULT_GOAL := help

help: ## list targets
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-10s %s\n", $$1, $$2}'

build: ## build both binaries into bin/
	go build -o bin/api ./cmd/api

test: ## unit tests
	go test -race ./...

cover: ## coverage per package
	go test -cover ./...

lint: ## vet and formatting check
	go vet ./... && test -z "$$(gofmt -l .)"

COMPOSE_TEST := docker compose -f docker-compose.yml -f docker-compose.test.yml
TEST_DATABASE_URL ?= postgres://hospital:hospital@localhost:5433/hospital_test?sslmode=disable

test-db: ## repository tests against postgres from compose (host port 5433)
	$(COMPOSE_TEST) up -d --wait postgres
	TEST_DATABASE_URL=$(TEST_DATABASE_URL) go test -race -count=1 ./internal/repository/...
