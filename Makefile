.PHONY: generate test test-contract lint lint-backend lint-frontend run build clean hooks temporal-dev temporal-down temporal-smoke

# Temporal dev-server service name in docker-compose.yml.
TEMPORAL_SERVICE ?= temporal
TEMPORAL_TASK_QUEUE ?= temflowral

# Point this clone at versioned hooks under .githooks/ (run once after clone)
hooks:
	git config core.hooksPath .githooks
	@chmod +x .githooks/*
	@echo "Git hooks enabled (core.hooksPath=.githooks)"

# Run all tests
test:
	cd backend && go test -race ./...
	cd frontend && npm test -- --run

# Regenerate contract-derived source code (Go server + TS client)
generate:
	cd backend && go generate ./...
	cd frontend && npm run generate

# Run tests with coverage report
test-coverage:
	cd backend && go test -race -coverprofile=coverage.out ./...
	cd backend && go tool cover -html=coverage.out

# Run all linters
lint: lint-backend lint-frontend

# Use the CI-pinned golangci-lint version; no separate install required.
lint-backend:
	./scripts/run-golangci-lint.sh

lint-frontend:
	cd frontend && npm run lint

# Start full stack locally (requires Docker)
run:
	docker compose up

# Start a local Temporal dev server via docker-compose (no local install).
# Serves gRPC on :7233 and the Web UI on http://localhost:8233.
temporal-dev:
	docker compose up $(TEMPORAL_SERVICE)

# Stop and remove the Temporal dev server container.
temporal-down:
	docker compose down

# Execute the registered smoke workflow using the CLI inside the dev container.
# Requires `make temporal-dev` and `make run-backend` to be running.
temporal-smoke:
	docker compose exec $(TEMPORAL_SERVICE) temporal workflow execute \
		--workflow-id "temflowral-smoke-$$(date +%s)" \
		--type temflowral.noop \
		--task-queue $(TEMPORAL_TASK_QUEUE) \
		--input '"hello"'

# Start backend only (connects to the Temporal dev server above)
run-backend:
	cd backend && go run cmd/server/main.go

# Run E2E tests (Playwright starts Prism + Next.js by default)
e2e:
	cd frontend && npm run e2e

# Run contract conformance (validate responses vs api/openapi.yaml).
# Defaults to the Prism mock; set API_BASE_URL=http://localhost:8080 for the
# live backend.
test-contract:
	cd frontend && npm run test:contract

# Clean build artifacts
clean:
	rm -f backend/coverage.out
	rm -rf frontend/.next frontend/test-results frontend/playwright-report frontend/blob-report