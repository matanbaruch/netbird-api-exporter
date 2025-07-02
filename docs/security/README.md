# Security Documentation

This directory contains security-related documentation for the NetBird API Exporter project.

## Documents

### [External PR Testing](external-pr-testing.md)

**For Maintainers and Contributors**

Comprehensive guide on how we securely handle pull requests from external contributors that require access to repository secrets. Covers:

- Security model and safeguards
- Approval process for maintainers
- Testing workflow for contributors
- Troubleshooting and best practices

## Security Overview

### Repository Secrets

The project uses the following secrets for testing:

- **`NETBIRD_API_TOKEN`** - Required for integration tests with real NetBird API
- **`CODECOV_TOKEN`** - For code coverage reporting
- **`GITHUB_TOKEN`** - Automatic GitHub token for workflows

### External Contributions

External contributors (from forks) cannot access repository secrets by default. We've implemented a secure approval system that:

1. **Runs basic tests automatically** (no secrets required)
2. **Requires manual approval** for tests that need secrets
3. **Uses GitHub labels** for approval workflow
4. **Provides clear feedback** to contributors

### Security Features

- **Manual code review required** before secret access
- **Environment protection** for sensitive workflows
- **Audit trail** of all approvals and test runs
- **Clear separation** between safe and sensitive tests

## Quick Start for Maintainers

To approve an external PR for testing:

**Option 1: Comment (Easiest)**

```
/approve
```

**Option 2: Manual Label**
Add the `approved-for-testing` label via GitHub UI.

## Quick Start for Contributors

When you submit a PR from a fork:

1. **Basic tests run immediately** - no waiting needed
2. **You'll see a security check comment** explaining the process
3. **Integration tests require approval** from maintainers
4. **Be patient** - maintainers will review and approve safe PRs

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
