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
