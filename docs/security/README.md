# Security Documentation

This directory contains security-related documentation for the NetBird API Exporter project.

## Security Overview

### Integration Testing

The project uses a self-hosted NetBird instance for integration testing. CI workflows automatically spin up a local NetBird server container, create an API token, and run tests against it. No external secrets are required.

### Repository Secrets

The project uses the following secrets:

- **`CODECOV_TOKEN`** - For code coverage reporting
- **`GITHUB_TOKEN`** - Automatic GitHub token for workflows

### Security Features

- **No secret dependencies for testing** - Self-hosted NetBird eliminates the need for shared API tokens
- **All contributors can run full tests** - External contributors from forks have full test access
- **Audit trail** of all test runs via GitHub Actions logs

## Reporting Security Issues

If you discover a security vulnerability, please:

1. **Do NOT open a public issue**
2. **Email the maintainers** (see [SECURITY.md](../../SECURITY.md))
3. **Provide details** about the vulnerability
4. **Allow time** for investigation and fix

## Additional Resources

- [Main Security Policy](../../SECURITY.md)
- [Contributing Guidelines](../../CONTRIBUTING.md)
- [Code of Conduct](../../CODE_OF_CONDUCT.md)

---

**Note:** This documentation is actively maintained. If you find outdated information or have suggestions for improvement, please open an issue or submit a PR.
