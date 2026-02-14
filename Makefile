.PHONY: build run clean test docker-build docker-run docker-compose-up docker-compose-down setup-precommit install-hooks verify test-integration-setup test-integration-teardown

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=netbird-api-exporter
DOCKER_IMAGE=netbird-api-exporter

# Build the binary
build:
	$(GOBUILD) -o $(BINARY_NAME) -v ./

# Run the application
run: build
	./$(BINARY_NAME)

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)

# Run tests
test:
	$(GOTEST) -v ./...

# Run unit tests only
test-unit:
	./scripts/run-tests.sh unit

# Run integration tests (auto-starts local NetBird if NETBIRD_API_TOKEN not set)
test-integration:
	./scripts/run-tests.sh integration

# Start self-hosted NetBird for integration testing
test-integration-setup:
	./scripts/setup-test-netbird.sh

# Stop self-hosted NetBird after integration testing
test-integration-teardown:
	./scripts/teardown-test-netbird.sh

# Run performance tests
test-performance:
	./scripts/run-tests.sh performance

# Run benchmark tests
test-benchmark:
	./scripts/run-tests.sh benchmark

# Run all tests with coverage
test-all:
	./scripts/run-tests.sh all

# Generate comprehensive coverage report
coverage:
	./scripts/coverage.sh generate

# Generate coverage report for unit tests only
coverage-unit:
	./scripts/coverage.sh --unit-only

# Generate coverage report for integration tests only (requires NETBIRD_API_TOKEN)
coverage-integration:
	./scripts/coverage.sh --integration-only

# Check coverage thresholds against existing coverage data
coverage-check:
	./scripts/coverage.sh check

# Clean coverage reports and artifacts
coverage-clean:
	./scripts/coverage.sh clean

# Generate coverage and enforce minimum thresholds (for CI/CD)
coverage-ci:
	./scripts/coverage.sh generate --threshold 80

# Clean test artifacts
test-clean:
	./scripts/run-tests.sh clean

# Download dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Build Docker image
docker-build:
	docker build -t $(DOCKER_IMAGE) .

# Run with Docker
docker-run: docker-build
	docker run -d \
		--name $(DOCKER_IMAGE) \
		-p 8080:8080 \
		-e NETBIRD_API_TOKEN=${NETBIRD_API_TOKEN} \
		$(DOCKER_IMAGE)

# Stop and remove Docker container
docker-stop:
	docker stop $(DOCKER_IMAGE) || true
	docker rm $(DOCKER_IMAGE) || true

# Start with Docker Compose
docker-compose-up:
	docker compose up -d

# Stop Docker Compose
docker-compose-down:
	docker compose down

# View logs
docker-compose-logs:
	docker compose logs -f netbird-api-exporter

# Setup development environment (install air and create .env if needed)
dev-setup:
	@echo "Setting up development environment..."
	@if ! command -v $(shell go env GOPATH)/bin/air >/dev/null 2>&1; then \
		echo "Installing air for hot reload..."; \
		go install github.com/air-verse/air@latest; \
	else \
		echo "✅ air is already installed"; \
	fi
	@if [ ! -f .env ]; then \
		echo "Creating .env file from env.example..."; \
		cp env.example .env; \
		echo "⚠️  Please edit .env file with your NetBird API token"; \
		echo "📝 You can find the .env file in the project root"; \
	else \
		echo "✅ .env file exists"; \
	fi
	@echo "🚀 Development environment is ready! Run 'make dev' to start."

# Development mode with live reload (requires air: go install github.com/air-verse/air@latest)
dev: dev-setup
	@if [ -f .env ]; then \
		echo "Loading environment variables from .env file..."; \
		bash -c 'set -a && source .env && set +a && GO111MODULE=on $(shell go env GOPATH)/bin/air'; \
	else \
		echo "❌ .env file not found. Run 'make dev-setup' first."; \
		exit 1; \
	fi

# Format code
fmt:
	$(GOCMD) fmt ./...

# Lint code (requires golangci-lint)
lint:
	@echo "Running golangci-lint..."
	GO111MODULE=on golangci-lint run
	@echo "Running go vet..."
	GO111MODULE=on go vet ./...
	@echo "Running go fmt check..."
	@test -z "$$(GO111MODULE=on gofmt -l .)" || (echo "Code is not formatted. Run 'make fmt' to fix." && exit 1)
	@echo "All linting checks passed!"

# Security scan (requires gosec)
security:
	gosec ./...

# Build for multiple platforms
build-all:
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o bin/$(BINARY_NAME)-linux-amd64
	GOOS=linux GOARCH=arm64 $(GOBUILD) -o bin/$(BINARY_NAME)-linux-arm64
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -o bin/$(BINARY_NAME)-darwin-amd64
	GOOS=darwin GOARCH=arm64 $(GOBUILD) -o bin/$(BINARY_NAME)-darwin-arm64
	GOOS=windows GOARCH=amd64 $(GOBUILD) -o bin/$(BINARY_NAME)-windows-amd64.exe

# Run all checks (tests, linting, formatting, coverage)
check: test lint coverage-check
	@echo "All checks passed!"

# Set up pre-commit hooks
setup-precommit:
	@echo "Setting up pre-commit hooks..."
	@./scripts/setup-precommit.sh

# Install pre-commit hooks (simple version)
install-hooks:
	@echo "Installing simple git pre-commit hook..."
	@cp scripts/pre-commit .git/hooks/pre-commit
	@chmod +x .git/hooks/pre-commit
	@echo "✅ Pre-commit hook installed!"

# Verify development environment setup
verify:
	@echo "Verifying development environment..."
	@./scripts/verify-setup.sh

# Help
help:
	@echo "Available targets:"
	@echo "  build           - Build the binary"
	@echo "  run             - Build and run the application"
	@echo "  clean           - Clean build artifacts"
	@echo "  test            - Run tests"
	@echo "  test-unit       - Run unit tests only"
	@echo "  test-integration - Run integration tests (auto-starts local NetBird)"
	@echo "  test-integration-setup - Start self-hosted NetBird for testing"
	@echo "  test-integration-teardown - Stop self-hosted NetBird"
	@echo "  test-performance - Run performance tests"
	@echo "  test-benchmark  - Run benchmark tests"
	@echo "  test-all        - Run all tests with coverage"
	@echo "  test-clean      - Clean test artifacts"
	@echo "  coverage        - Generate comprehensive coverage report"
	@echo "  coverage-unit   - Generate coverage report for unit tests only"
	@echo "  coverage-integration - Generate coverage for integration tests"
	@echo "  coverage-check  - Check coverage thresholds"
	@echo "  coverage-clean  - Clean coverage reports and artifacts"
	@echo "  coverage-ci     - Generate coverage with CI/CD thresholds"
	@echo "  deps            - Download and tidy dependencies"
	@echo "  docker-build    - Build Docker image"
	@echo "  docker-run      - Run with Docker"
	@echo "  docker-stop     - Stop Docker container"
	@echo "  docker-compose-up   - Start with Docker Compose"
	@echo "  docker-compose-down - Stop Docker Compose"
	@echo "  docker-compose-logs - View container logs"
	@echo "  dev-setup       - Setup development environment (install air, create .env)"
	@echo "  dev             - Development mode with live reload"
	@echo "  fmt             - Format code"
	@echo "  lint            - Lint code (golangci-lint, go vet, format check)"
	@echo "  check           - Run all checks (tests + linting)"
	@echo "  setup-precommit - Interactive setup for pre-commit hooks"
	@echo "  install-hooks   - Install simple git pre-commit hook"
	@echo "  verify          - Verify development environment setup"
	@echo "  security        - Run security scan"
	@echo "  build-all       - Build for multiple platforms"
	@echo "  help            - Show this help"
