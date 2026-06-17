.PHONY: test lint run build clean

# Run all tests
test:
	cd backend && go test -race ./...
	cd frontend && npm test -- --run

# Run tests with coverage report
test-coverage:
	cd backend && go test -race -coverprofile=coverage.out ./...
	cd backend && go tool cover -html=coverage.out

# Run linter
lint:
	cd backend && golangci-lint run
	cd frontend && npm run lint

# Start full stack locally (requires Docker)
run:
	docker compose up

# Start backend only (no Docker — requires local temporalite)
run-backend:
	cd backend && go run cmd/server/main.go

# Run E2E tests (stack must be running)
e2e:
	cd frontend && npx playwright test

# Clean build artifacts
clean:
	rm -f backend/coverage.out
	rm -rf frontend/.next frontend/test-results