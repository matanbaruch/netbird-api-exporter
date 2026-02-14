# NetBird API Exporter - Development Rules

This document provides AI agent guidelines for working with this codebase. It is consumed by AI assistants like Claude, Cursor, Copilot, and others.

## Project Overview

NetBird API Exporter is a Prometheus exporter that collects metrics from the NetBird API, providing insights into peers, groups, users, DNS settings, and networks. The main entry point is [main.go](main.go) which sets up the HTTP server and registers the exporters.

## Project Structure

### Core Components
- **Main Application**: [main.go](main.go) - HTTP server setup, middleware, and routing
- **Exporters Package**: [pkg/exporters/](pkg/exporters) - Prometheus exporters for different NetBird resources
- **NetBird Client**: [pkg/netbird/](pkg/netbird) - API client for NetBird services
- **Utilities**: [pkg/utils/](pkg/utils) - Common utilities and helpers
- **Configuration**: [env.example](env.example) - Environment variable examples

### Documentation & Configuration
- **Main Documentation**: [README.md](README.md) - Project overview and usage
- **Architecture**: [ARCHITECTURE.md](ARCHITECTURE.md) - Technical architecture details
- **Contributing**: [CONTRIBUTING.md](CONTRIBUTING.md) - Development guidelines
- **Changelog**: [CHANGELOG.md](CHANGELOG.md) - Release history and changes

### Deployment & Operations
- **Docker**: [Dockerfile](Dockerfile) and [docker-compose.yml](docker-compose.yml)
- **Helm Charts**: [charts/netbird-api-exporter/](charts/netbird-api-exporter)
- **Systemd Service**: [netbird-exporter.service](netbird-exporter.service)
- **Prometheus Config**: [prometheus.yml.example](prometheus.yml.example)

## Changelog Management (CRITICAL)

### Always Update Changelog

**Every meaningful change MUST include a changelog entry.** Use the changelog script located at [scripts/update-changelog.sh](scripts/update-changelog.sh):

```bash
./scripts/update-changelog.sh [type] [description]
```

### Change Types

- **breaking**: Breaking changes requiring user action (include migration instructions)
- **feature**: New functionality and enhancements (use present tense, e.g., "Add support for...")
- **bugfix**: Bug fixes and corrections (reference issue numbers, e.g., "#123")
- **security**: Security-related changes and vulnerability fixes
- **deprecated**: Features that will be removed in future versions
- **removed**: Features that have been removed

### Examples

```bash
# New features
./scripts/update-changelog.sh feature "Add support for custom metrics endpoint"

# Bug fixes
./scripts/update-changelog.sh bugfix "Fix memory leak in DNS exporter (#123)"

# Breaking changes
./scripts/update-changelog.sh breaking "Remove deprecated --old-flag parameter"

# Security fixes
./scripts/update-changelog.sh security "Fix potential path traversal vulnerability"
```

### Workflow
1. Make code changes
2. Run changelog script with appropriate type and description
3. Review changelog with `git diff CHANGELOG.md`
4. Commit both code changes and changelog update together

## Code Style & Standards

### Go Conventions
- Follow standard Go conventions and best practices
- Use meaningful variable and function names
- Add comments for complex logic
- Ensure proper error handling with structured logging using logrus
- Write tests for new functionality

### Logging Standards
- Use logrus for structured logging
- Include relevant fields in log entries (see [main.go](main.go) for examples)
- Use appropriate log levels: Debug, Info, Warn, Error, Fatal
- Debug logging should be comprehensive for HTTP requests/responses

### Error Handling
- Always handle errors gracefully
- Use structured logging for error reporting
- Provide meaningful error messages
- Follow the error handling patterns in existing exporters

## Environment Variables

Key configuration through environment variables (see [env.example](env.example)):
- `NETBIRD_API_URL`: NetBird API endpoint (default: http://localhost:8081 for tests, https://api.netbird.io for production)
- `NETBIRD_API_TOKEN`: API token for authentication (auto-generated from self-hosted instance in CI)
- `LISTEN_ADDRESS`: HTTP server listen address (default: :8080)
- `METRICS_PATH`: Metrics endpoint path (default: /metrics)
- `LOG_LEVEL`: Logging level (default: info)

## Testing & Quality

- Write unit tests for new functionality
- Follow the existing test patterns in the codebase
- Use proper error assertions
- Ensure tests are deterministic and fast
- Integration tests use a self-hosted NetBird instance (see [tests/](tests/) and [scripts/setup-test-netbird.sh](scripts/setup-test-netbird.sh))
- Run `make test-integration-setup` to start a local NetBird server, then `make test-integration` to run tests

## Git Commit Messages

Use conventional commit format:
```
type(scope): description

- Detailed changes
- Reference issue numbers when applicable

Fixes #42
```

Types: feat, fix, docs, style, refactor, test, chore

## Development Workflow

1. **Setup**: Use [docker-compose.yml](docker-compose.yml) for local development
2. **Code Changes**: Follow the patterns in existing exporters
3. **Changelog**: Always update [CHANGELOG.md](CHANGELOG.md) using the script
4. **Testing**: Run tests and verify functionality
5. **Documentation**: Update relevant docs if needed
6. **Commit**: Include both code and changelog changes

## API Integration

When working with NetBird API:
- Use the client patterns in [pkg/netbird/](pkg/netbird)
- Follow the exporter patterns in [pkg/exporters/](pkg/exporters)
- Implement proper rate limiting and error handling
- Add debug logging for API calls
- Handle authentication and API versioning correctly

## Prometheus Metrics

When adding new metrics:
- Follow Prometheus naming conventions
- Use appropriate metric types (Counter, Gauge, Histogram, Summary)
- Include relevant labels for dimensionality
- Document metric purpose and usage
- Follow existing patterns in the exporters

## Security Considerations

- Never log sensitive information (API tokens, credentials)
- Validate all inputs
- Use secure defaults
- Follow the security patterns in [Dockerfile](Dockerfile)
- Run as non-root user (nobody) in containers

