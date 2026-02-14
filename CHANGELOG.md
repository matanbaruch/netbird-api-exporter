# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]


### Bugfix
- Fix unit tests making real API calls to api.netbird.io causing 30s timeouts
Files modified in this change:
- Modified: pkg/exporters/dns_test.go
- Modified: pkg/exporters/groups_test.go
- Modified: pkg/exporters/networks_test.go
- Modified: pkg/exporters/peers_test.go
- Modified: pkg/exporters/users_test.go

## [0.1.74] - 2026-02-14


### Features
- Replace cloud API token dependency with self-hosted NetBird for integration testing, enabling all contributors to run the full test suite without secrets
Files modified in this change:
- Deleted: .github/environments/external-pr-testing.yml
- Modified: .github/workflows/coverage.yml
- Deleted: .github/workflows/external-pr.yml
- Modified: .github/workflows/pr-build.yml
- Deleted: .github/workflows/pr-tests.yml
- Deleted: .github/workflows/privileged-tests.yml
- Modified: .github/workflows/test.yml
- Modified: AGENTS.md
- Modified: Makefile
- Modified: docs/security/README.md
- Deleted: docs/security/approval-verification.md
- Deleted: docs/security/external-pr-testing.md
- Modified: pkg/integration_test.go
- Modified: scripts/run-tests.sh
- New: scripts/setup-test-netbird.sh
- New: scripts/teardown-test-netbird.sh
- New: tests/

## [0.1.73] - 2026-01-31

## [0.1.72] - 2026-01-08

## [0.1.71] - 2026-01-08

## [0.1.70] - 2026-01-03

## [0.1.69] - 2025-12-11

## [0.1.68] - 2025-12-11

## [0.1.67] - 2025-12-11

## [0.1.66] - 2025-12-11

## [0.1.65] - 2025-12-11

## [0.1.64] - 2025-12-11

## [0.1.63] - 2025-12-11

## [0.1.62] - 2025-12-11

## [0.1.61] - 2025-12-11

## [0.1.60] - 2025-12-11

## [0.1.59] - 2025-12-01

## [0.1.58] - 2025-11-03

## [0.1.57] - 2025-10-17

## [0.1.56] - 2025-09-10

## [0.1.55] - 2025-09-09

## [0.1.54] - 2025-09-09

## [0.1.53] - 2025-08-29

## [0.1.52] - 2025-08-06

## [0.1.51] - 2025-07-20

## [0.1.50] - 2025-07-13

## [0.1.49] - 2025-07-05


### Bugfix
- Fix GitHub Actions workflow syntax errors in external PR handling
Files modified in this change:
- Modified: .github/workflows/external-pr.yml

## [0.1.48] - 2025-07-02

### Features

- Add secure external PR testing workflow with comment-based approval system (/approve command)
  Files modified in this change:
- Modified: .github/workflows/coverage.yml
- Modified: .github/workflows/test.yml
- Modified: CHANGELOG.md
- Modified: README.md
- Modified: charts/netbird-api-exporter/Chart.yaml
- New: .github/workflows/external-pr.yml
- New: docs/security/

## [0.1.47] - 2025-06-24

## [0.1.46] - 2025-06-24

### Bugfix

- Fix import paths to match updated module name in go.mod

## [0.1.45] - 2025-06-24

## [0.1.44] - 2025-06-24

### Features

- Add netbird_peer_connection_status_by_name metric to track individual peer connection status with peer name, ID, and connection state labels
  Files modified in this change:
- Modified: pkg/exporters/peers.go

## [0.1.43] - 2025-06-23

## [0.1.42] - 2025-06-23

## [0.1.41] - 2025-06-07

### Features

- Add Helm chart testing with kubectl using Azure k8s-deploy action
  Files modified in this change:
- Modified: .github/workflows/test.yml

## [0.1.40] - 2025-06-05

## [0.1.39] - 2025-06-05

### Bugfix

- Fix Prometheus metric naming to follow official guidelines - remove \_total suffix from gauge metrics
  Files modified in this change:
- Modified: pkg/exporters/dns.go
- Modified: pkg/exporters/dns_test.go
- Modified: pkg/exporters/groups.go
- Modified: pkg/exporters/groups_test.go
- Modified: pkg/exporters/networks.go
- Modified: pkg/exporters/networks_test.go
- Modified: pkg/exporters/peers.go
- Modified: pkg/exporters/peers_test.go
- Modified: pkg/exporters/users.go
- Modified: pkg/exporters/users_test.go

## [0.1.38] - 2025-06-02

## [0.1.37] - 2025-06-02

### Features

- Enhance PR build workflow to use real NETBIRD_API_TOKEN secret for comprehensive integration testing
  Files modified in this change:
- Modified: CHANGELOG.md
- Modified: charts/netbird-api-exporter/Chart.yaml
- New: .github/workflows/pr-build.yml
- Add PR build workflow for Docker images and Helm chart validation
  Files modified in this change:
- Modified: charts/netbird-api-exporter/Chart.yaml
- New: .github/workflows/pr-build.yml

- Add values.schema.json to Helm chart for configuration validation and documentation
  Files modified in this change:
- New: charts/netbird-api-exporter/values.schema.json

## [0.1.36] - 2025-06-02

### Bugfix

- Fix linting errors for unchecked error return values in resp.Body.Close() calls
  Files modified in this change:
- Modified: pkg/integration_test.go
- Fix linting errors (errcheck) in test files for unchecked error return values
  Files modified in this change:
- Modified: pkg/netbird/client_integration_test.go
- Modified: pkg/utils/config_test.go
- Fix unit tests job running integration tests in GitHub workflow
  Files modified in this change:
- Modified: scripts/run-tests.sh
- Fix linting issues in performance tests including errcheck, gosec, and ineffassign violations

### Features

- Add comprehensive GitHub Actions test workflow with matrix testing
  Files modified in this change:
- Modified: .github/workflows/lint.yml
- Modified: CHANGELOG.md
- Modified: Makefile
- Modified: charts/netbird-api-exporter/Chart.yaml
- New: .github/workflows/test.yml
- New: coverage.html
- New: pkg/exporters/performance_test.go
- New: pkg/integration_test.go
- New: pkg/netbird/client_integration_test.go
- New: pkg/utils/config_test.go
- New: scripts/run-tests.sh
- Add comprehensive test suite including integration and performance tests
  Files modified in this change:
- Modified: CHANGELOG.md
- Modified: Makefile
- Modified: charts/netbird-api-exporter/Chart.yaml
- New: coverage.html
- New: pkg/exporters/performance_test.go
- New: pkg/integration_test.go
- New: pkg/netbird/client_integration_test.go
- New: pkg/utils/config_test.go
- New: scripts/run-tests.sh
- Add comprehensive test suite including integration tests and performance tests
  Files modified in this change:
- Modified: Makefile
- New: pkg/exporters/performance_test.go
- New: pkg/integration_test.go
- New: pkg/netbird/client_integration_test.go
- New: pkg/utils/config_test.go
- New: scripts/run-tests.sh

## [0.1.35] - 2025-06-02

## [0.1.34] - 2025-06-02

## [0.1.33] - 2025-06-02

## [0.1.32] - 2025-06-01

## [0.1.31] - 2025-06-01

### Bugfix

- Fix update-changelog.sh to properly update Helm chart artifacthub.io/changes with correct format and handle empty descriptions
  Files modified in this change:
- Added: .github/dependabot.yml
- Modified: scripts/update-changelog.sh

## [0.1.30] - 2025-06-01

## [0.1.29] - 2025-06-01

### Bugfix

- Fix ArtifactHub.io annotations by properly quoting strings and correcting alternative name
  Files modified in this change:
- Modified: charts/netbird-api-exporter/Chart.yaml

## [0.1.28] - 2025-06-01

### Features

- Refactor GitHub issue templates with improved structure and NetBird-specific context
  Files modified in this change:
- Deleted: .github/ISSUE_TEMPLATE.md
- Modified: .github/ISSUE_TEMPLATE/bug_report.md
- Deleted: .github/ISSUE_TEMPLATE/custom.md
- Modified: .github/ISSUE_TEMPLATE/feature_request.md
- New: .github/ISSUE_TEMPLATE/config.yml
- New: .github/ISSUE_TEMPLATE/documentation.md

## [0.1.27] - 2025-06-01

### Features

- Enhanced update-changelog.sh script to also update Helm chart artifacthub.io/changes annotation and include uncommitted files summary
  Files modified in this change:
- Modified: .github/workflows/release.yml
- Modified: ARCHITECTURE.md
- Modified: CHANGELOG.md
- Modified: README.md
- Modified: SECURITY.md
- Modified: charts/netbird-api-exporter/Chart.yaml
- Modified: scripts/update-changelog.sh

### Bugfix

- Fix Helm chart name to resolve packaging filename mismatch
  Files modified in this change:
- Modified: charts/netbird-api-exporter/Chart.yaml

### Security

- Add build provenance attestations for Docker images and Go binaries

## [0.1.26] - 2025-06-01

### Features

- Add comprehensive GitHub badges including build status, Go version, Prometheus, Docker, Kubernetes, license, and distribution metrics
  Files modified in this change:
- Modified: README.md

## [0.1.25] - 2025-06-01

### Features

- Add icon to helm chart for better visual identification in Artifact Hub
  Files modified in this change:
- Modified: charts/netbird-api-exporter/Chart.yaml

### Bugfix

- Fix Helm chart name to resolve packaging filename mismatch
  Files modified in this change:
- Modified: charts/netbird-api-exporter/Chart.yaml

## [0.1.24] - 2025-06-01

### Security

- Add build provenance attestations for Docker images and Go binaries

## [0.1.23] - 2025-06-01

### Bugfix

- Fix GitHub Actions workflow by installing ORAS before using it in release pipeline
- Fix GitHub Actions release workflow Helm packaging error with 'latest' version

## [0.1.22] - 2025-06-01

### Features

- Only commit version bump after successful Docker and Helm builds
- Only push git tags after successful Docker and Helm builds

## [0.1.21] - 2025-06-01

### Features

- Enhanced release automation and build process improvements

## [0.1.20] - 2025-06-01

### Features

- Build process and CI/CD improvements

## [0.1.19] - 2025-06-01

### Features

- Release pipeline enhancements

## [0.1.18] - 2025-06-01

### Features

- Add publishing to artifacthub.io for the OCI package of the helm chart

## [0.1.17] - 2025-05-31

### Features

- Add comprehensive Grafana dashboard with documentation
- Add comprehensive SECURITY.md with vulnerability reporting procedures and security guidelines

## [0.1.16] - 2025-05-31

### Bugfix

- Fix bundler version compatibility for Dependabot (update from v1.17.2 to v2.5.0)

## [0.1.15] - 2025-05-31

### Features

- Enhanced development environment and tooling improvements

## [0.1.14] - 2025-05-31

### Features

- Add comprehensive test coverage for NetBird client and types
- Migrate from legacy .cursorrules to modern .cursor/rules MDC format

### Bugfix

- Fix changelog update script breaking the changelog structure and creating duplicate sections

## [0.1.13] - 2025-05-31

### Features

- Development workflow improvements

## [0.1.12] - 2025-05-31

### Features

- Improve development workflow with automatic environment setup

## [0.1.11] - 2025-05-31

### Features

- Enhanced release process with automated changelog generation
- Add changelog file for better release tracking

## [0.1.10] - 2025-05-31

### Features

- Enhanced release process with automated changelog generation
- Add changelog file for better release tracking

<!--
When adding entries to the changelog, use the following guidelines:

### BREAKING CHANGES
- List any breaking changes that require user action
- Include migration instructions if needed

### Features
- New functionality and enhancements
- Use present tense (e.g., "Add support for...")

### Bugfix
- Bug fixes and corrections
- Reference issue numbers if applicable (e.g., "Fix memory leak in exporter (#123)")

### Security
- Security-related changes
- Vulnerability fixes

### Deprecated
- Features that will be removed in future versions

### Removed
- Features that have been removed
-->

## [0.1.9] - 2024-12-16

### Bugfix

- Refactor logging in main.go and networks.go for consistency
- Adjusted formatting in debug logging middleware
- Updated log fields in NetworksExporter to enhance readability and maintain uniformity in log output

## [0.1.8] - 2024-12-16

### Features

- Implement debug logging middleware for HTTP requests and responses
- Enhanced logging in exporters (DNS, Groups, Networks, Peers, Users) to capture API fetch details and metrics updates

## [0.1.7] - 2024-12-16

### Bugfix

- Update installation instructions in README and helm.md to reflect OCI registry usage
- Removed the need for adding a Helm repository
- Added examples for installing the NetBird API Exporter directly from the OCI registry
- Added configurations for ExternalSecret and production setups

## [0.1.6] - 2024-12-16

### Features

- Enhanced Docker security and permissions
- Added wget to Dockerfile dependencies
- Created app directory with proper permissions for the nobody user
- Adjusted binary copy command to set ownership and executable permissions

### Bugfix

- Update README.md documentation

## [0.1.5] - 2024-12-16

### Bugfix

- Add Jekyll build artifacts to gitignore
- Fix GitHub Pages deployment with custom blue color scheme for Just the Docs theme
- Fix Liquid template syntax errors in docker.md by escaping Docker template syntax
- Remove conflicting just-the-docs gem from Gemfile to avoid conflicts with github-pages
- Add jekyll-remote-theme plugin for remote theme support
- Comment out missing logo reference
- Add Gemfile.lock for reproducible builds

## [0.1.4] - 2024-12-16

### Features

- Enhanced GitHub Actions workflows by adding paths-ignore for documentation files
- Transition to Just the Docs theme for documentation
- Improved navigation and refined installation guides
- Introduced new sections for better organization and accessibility
- Removed deprecated navigation data file

## [0.1.3] - 2024-12-16

### Features

- Transition to ReadTheDocs theme for documentation
- Enhanced navigation and improved installation guides
- Added detailed sections for Docker, Docker Compose, and security considerations
- Streamlined content organization and updated links for better accessibility

## [0.1.2] - 2024-12-16

### Features

- Refactored documentation structure and enhanced installation guides
- Added comprehensive installation methods including Docker, Docker Compose, Helm, and systemd
- Introduced new configuration files and workflows for GitHub Pages deployment
- Improved overall clarity and accessibility of the documentation

## [0.1.1] - 2024-12-16

### Bugfix

- Updated README.md to remove contributing section
- Enhanced GitHub Actions workflow by adding permissions for write access
