# Contributing to NetBird API Exporter

Thank you for your interest in contributing to the NetBird API Exporter! This document provides guidelines and instructions for contributing to this project.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Environment](#development-environment)
- [Making Changes](#making-changes)
- [Changelog Management](#changelog-management)
- [Code Style](#code-style)
- [Testing](#testing)
- [Submitting Changes](#submitting-changes)
- [Issue Reporting](#issue-reporting)
- [Documentation](#documentation)
- [Community](#community)

## Code of Conduct

This project follows a code of conduct to ensure a welcoming environment for all contributors. Please be respectful and professional in all interactions.

## Development Setup

### Prerequisites

- Go 1.23 or later
- Docker and Docker Compose (for containerized development)
- Git
- Make (for using the Makefile commands)
- golangci-lint (for linting)

### Optional Development Tools

- [gosec](https://github.com/securecodewarrior/gosec) for security scanning
- [air](https://github.com/air-verse/air) for live reload during development

## Getting Started

### 1. Fork and Clone

1. Fork the repository on GitHub
2. Clone your fork locally:

```bash
git clone https://github.com/YOUR_USERNAME/netbird-api-exporter.git
cd netbird-api-exporter
```

### 2. Set Up Environment

1. Copy the example environment file:

```bash
cp env.example .env
```

2. Edit `.env` with your NetBird API token:

```env
NETBIRD_API_TOKEN=your_token_here
NETBIRD_API_URL=https://api.netbird.io
LISTEN_ADDRESS=:8080
METRICS_PATH=/metrics
LOG_LEVEL=debug
```

### 3. Install Dependencies

```bash
make deps
```

### 4. Set Up Pre-commit Hooks

```bash
make setup-precommit
```

### 5. Verify Everything Works

```bash
# Build the project
make build

# Run tests
make test

# Run all checks (tests + linting)
make check
```

### 6. Development Mode

For active development with live reload:

```bash
# Install air for live reload
go install github.com/air-verse/air@latest

# Start development mode
make dev
```

## Making Changes

### Branching Strategy

- Create a new branch for each feature or bug fix
- Use descriptive branch names:
  - `feature/add-new-metric`
  - `fix/auth-token-validation`
  - `docs/update-installation-guide`

```bash
git checkout -b feature/your-feature-name
```

### Project Structure

```
├── main.go                 # Application entry point
├── pkg/
│   ├── exporters/         # Prometheus exporters for different APIs
│   ├── netbird/          # NetBird API client
│   └── utils/            # Utility functions
├── charts/               # Helm charts for Kubernetes deployment
├── docs/                 # Documentation
└── tests/               # Test files
```

### Adding New Metrics

When adding new metrics:

1. Add the metric definition in the appropriate exporter file in `pkg/exporters/`
2. Implement the data collection logic
3. Update the documentation in `README.md`
4. Add appropriate tests
5. Consider backwards compatibility

### Code Changes Workflow

1. Make your changes
2. Run formatting: `make fmt`
3. Run linting: `make lint`
4. Run tests: `make test`
5. Test locally: `make run` or `make dev`
6. Update documentation if needed
7. Update changelog if needed (see [Changelog Management](#changelog-management))

## Changelog Management

This project maintains a changelog following the [Keep a Changelog](https://keepachangelog.com/) format. The changelog helps users and contributors understand what has changed between releases.

### When to Update the Changelog

Update the changelog for:

- **BREAKING CHANGES**: Any changes that break backward compatibility
- **Features**: New functionality and enhancements
- **Bugfixes**: Bug fixes and corrections
- **Security**: Security-related changes
- **Deprecated**: Features that will be removed in future versions
- **Removed**: Features that have been removed

### How to Update the Changelog

#### Option 1: Manual Update

Edit `CHANGELOG.md` directly and add your entry under the `[Unreleased]` section in the appropriate category.

#### Option 2: Using the Helper Script

Use the provided script to add entries easily:

```bash
# Add a new feature
./scripts/update-changelog.sh feature "Add support for custom metrics endpoint"

# Add a bug fix
./scripts/update-changelog.sh bugfix "Fix memory leak in DNS exporter (#123)"

# Add a breaking change
./scripts/update-changelog.sh breaking "Remove deprecated --old-flag parameter"

# Available types: breaking, feature, bugfix, security, deprecated, removed
```

### Changelog Format

Entries should be:

- Written in past tense
- Clear and concise
- Include issue/PR numbers when relevant
- Grouped by type (breaking, feature, bugfix, etc.)

Example:

```markdown
## [Unreleased]

### BREAKING CHANGES

- Remove deprecated --old-flag parameter

### Features

- Add support for custom metrics endpoint
- Enhanced logging for better debugging

### Bugfix

- Fix memory leak in DNS exporter (#123)
- Resolve configuration validation issue (#456)
```

### Automated Changelog

The changelog is automatically integrated into releases:

- When a new version is released, the `[Unreleased]` section becomes the new version
- A new empty `[Unreleased]` section is created
- The changelog content is included in GitHub releases

## Code Style

### Go Style Guidelines

- Follow the [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Use `gofmt` for formatting (automated via `make fmt`)
- Write clear, self-documenting code with appropriate comments
- Use meaningful variable and function names

### Specific Guidelines

1. **Error Handling**: Always handle errors appropriately

   ```go
   if err != nil {
       log.WithError(err).Error("Failed to fetch data")
       return err
   }
   ```

2. **Logging**: Use structured logging with logrus

   ```go
   log.WithFields(logrus.Fields{
       "metric": "peers_total",
       "count":  len(peers),
   }).Info("Collected peers metric")
   ```

3. **Metrics**: Follow Prometheus naming conventions
   - Use `snake_case` for metric names
   - Include appropriate labels
   - Add help text for all metrics

### Linting

Run the linter before submitting:

```bash
make lint
```

This runs:

- `golangci-lint`
- `go vet`
- Format checking

## Testing

### Running Tests

```bash
# Run all tests
make test

# Run tests with verbose output
go test -v ./...

# Run tests with coverage
go test -cover ./...
```

### Writing Tests

- Write unit tests for new functionality
- Place test files alongside the code they test
- Use table-driven tests when appropriate
- Mock external dependencies (NetBird API calls)

Example test structure:

```go
func TestPeerExporter_CollectMetrics(t *testing.T) {
    tests := []struct {
        name     string
        peers    []netbird.Peer
        expected int
    }{
        {
            name:     "empty peers",
            peers:    []netbird.Peer{},
            expected: 0,
        },
        // Add more test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

## Submitting Changes

### Pull Request Process

1. **Update Documentation**: Ensure your changes are documented
2. **Test Thoroughly**: Run `make check` to ensure all tests pass
3. **Commit Messages**: Write clear, descriptive commit messages

#### Commit Message Format

```
type: brief description

Longer description if needed, explaining what changed and why.

Fixes #123
```

Types:

- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

#### Pull Request Checklist

- [ ] Tests pass (`make test`)
- [ ] Linting passes (`make lint`)
- [ ] Documentation updated (if applicable)
- [ ] CHANGELOG updated (if applicable)
- [ ] Branch is up to date with main
- [ ] Clear description of changes

### Review Process

1. Submit your pull request
2. Maintainers will review your changes
3. Address any feedback
4. Once approved, your changes will be merged

## Issue Reporting

### Bug Reports

When reporting bugs, please include:

- **Environment**: OS, Go version, deployment method
- **Configuration**: Relevant environment variables (mask sensitive data)
- **Steps to Reproduce**: Clear steps to reproduce the issue
- **Expected Behavior**: What you expected to happen
- **Actual Behavior**: What actually happened
- **Logs**: Relevant log output (set `LOG_LEVEL=debug`)

### Feature Requests

For feature requests:

- **Use Case**: Describe the problem you're trying to solve
- **Proposed Solution**: Your idea for implementing the feature
- **Alternatives**: Other solutions you've considered
- **Additional Context**: Any other relevant information

## Documentation

### Documentation Standards

- Keep documentation up to date with code changes
- Use clear, concise language
- Include examples where helpful
- Follow the existing documentation style

### Documentation Locations

- **README.md**: Main project documentation
- **docs/**: Detailed documentation and guides
- **Code Comments**: Inline documentation for complex logic
- **Helm Chart**: Chart documentation in `charts/netbird-api-exporter/`

## Community

### Getting Help

- **Issues**: For bugs and feature requests
- **Discussions**: For questions and general discussion
- **Documentation**: Check existing docs first

### Maintainers

Current maintainers:

- [@matanbaruch](https://github.com/matanbaruch)

## Development Tips

### Useful Make Commands

```bash
make help                 # Show all available commands
make dev                  # Development mode with live reload
make docker-compose-up    # Run with Docker Compose
make docker-compose-logs  # View container logs
make build-all           # Build for multiple platforms
make security           # Run security scan
```

### Debugging

1. Set `LOG_LEVEL=debug` in your environment
2. Use `make docker-compose-logs` to view logs
3. Test individual components with unit tests

### Performance Considerations

- Consider the impact on NetBird API rate limits
- Use appropriate scrape intervals
- Monitor memory usage for large deployments

## Pre-commit Hooks

We **strongly recommend** setting up pre-commit hooks to automatically run code quality checks before each commit. This helps maintain consistent code quality and catches issues early.

### Quick Setup

```bash
make setup-precommit
```

This interactive script will guide you through choosing between:

- **Simple Git Hook**: Basic bash script with no external dependencies
- **Pre-commit Framework**: Advanced hook management with additional checks

### What the Hooks Do

Pre-commit hooks automatically run:

- `go fmt` - Code formatting
- `golangci-lint` - Comprehensive linting
- `go vet` - Static analysis
- `go test` - Unit tests

### Manual Installation

If you prefer to install the simple hook directly:

```bash
make install-hooks
```

For detailed setup instructions and troubleshooting, see the [Pre-commit Hooks Guide](docs/getting-started/pre-commit-hooks.md).

---

Thank you for contributing to NetBird API Exporter! Your contributions help make NetBird monitoring better for everyone.
