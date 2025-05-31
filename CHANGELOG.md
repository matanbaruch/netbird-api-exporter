# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.13] - 2025-05-31

### Bugfix

- Fix changelog update script breaking the changelog structure and creating duplicate sections

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
- Adjusted checkout settings for better version control

[0.1.10]: https://github.com/matanbaruch/netbird-api-exporter/compare/v0.1.9...v0.1.10
[0.1.11]: https://github.com/matanbaruch/netbird-api-exporter/compare/v0.1.10...v0.1.11
[0.1.12]: https://github.com/matanbaruch/netbird-api-exporter/compare/v0.1.11...v0.1.12
[0.1.13]: https://github.com/matanbaruch/netbird-api-exporter/compare/v0.1.12...v0.1.13
[Unreleased]: https://github.com/matanbaruch/netbird-api-exporter/compare/v0.1.13...HEAD
[0.1.9]: https://github.com/matanbaruch/netbird-api-exporter/compare/v0.1.8...v0.1.9
[0.1.8]: https://github.com/matanbaruch/netbird-api-exporter/compare/v0.1.7...v0.1.8
[0.1.7]: https://github.com/matanbaruch/netbird-api-exporter/compare/v0.1.6...v0.1.7
[0.1.6]: https://github.com/matanbaruch/netbird-api-exporter/compare/v0.1.5...v0.1.6
[0.1.5]: https://github.com/matanbaruch/netbird-api-exporter/compare/v0.1.4...v0.1.5
[0.1.4]: https://github.com/matanbaruch/netbird-api-exporter/compare/v0.1.3...v0.1.4
[0.1.3]: https://github.com/matanbaruch/netbird-api-exporter/compare/v0.1.2...v0.1.3
[0.1.2]: https://github.com/matanbaruch/netbird-api-exporter/compare/v0.1.1...v0.1.2
[0.1.1]: https://github.com/matanbaruch/netbird-api-exporter/releases/tag/v0.1.1
