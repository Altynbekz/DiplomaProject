SHELL := /bin/sh
export GOCACHE ?= /tmp/diploma-go-cache
.PHONY: setup dev dev-local migrate-up generate fmt fmt-check lint test test-integration test-e2e check generate-check tools-check
setup:
	python3 scripts/setup-local.py
	cd frontend && npm ci

dev:
	python3 scripts/setup-local.py --compose
	docker compose up --build --wait

dev-local: setup
	./scripts/dev-local.sh

migrate-up:
	@test -n "$$MIGRATION_DATABASE_URL" || (echo 'Set MIGRATION_DATABASE_URL for the migration owner' >&2; exit 1)
	goose -dir backend/db/migrations postgres "$$MIGRATION_DATABASE_URL" up

tools-check:
	@test "$$(sqlc version)" = "v1.31.1"
	@test "$$(goose -version)" = "goose version: v3.30.1"

generate: tools-check
	cd backend && sqlc generate -f db/sqlc.yaml
	cd frontend && npm run generate

generate-check:
	python3 scripts/generate-check.py

fmt:
	cd backend && gofmt -w cmd internal
	cd frontend && npm run fmt

fmt-check:
	@test -z "$$(cd backend && gofmt -l cmd internal)"
	cd frontend && npm run fmt:check

lint:
	cd backend && go vet ./...
	cd frontend && npm run lint

test:
	cd backend && go test ./...
	cd frontend && npm test

test-integration:
	./scripts/test-integration.sh

test-e2e:
	./scripts/test-e2e.sh

check: fmt-check lint generate-check test
