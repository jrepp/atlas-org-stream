# Atlassian Organization Tool

## Overview
A Go CLI tool that queries Atlassian organization and team structure APIs and outputs NDJSON metadata for consumption by other tools.

## Recent Changes
- Enhanced tool with JSON schema output as first NDJSON element
- Added comprehensive statistics tracking throughout processing
- Implemented periodic and final NDJSON statistics output
- Added human-readable statistics summary to stderr
- Created GitHub Actions workflows for CI/CD, PR validation, and releases
- Improved error handling with graceful finalization

## Project Architecture
- Single Go binary with minimal dependencies
- Uses standard library for HTTP requests and JSON handling
- Modular design with separate structs for different API entities
- Configurable output (stdout or file) for pipeline integration
- Comprehensive statistics tracking with NDJSON output
- JSON schema validation for all output data

## GitHub Workflows
- **CI/CD Pipeline** (.github/workflows/ci.yml): Automated testing across multiple Go versions, security scanning, and artifact building
- **Pull Request Validation** (.github/workflows/pr.yml): Code formatting, linting, and automated validation for PRs
- **Release Management** (.github/workflows/release.yml): Automated releases with multi-platform binaries when tags are pushed

## Development Workflow
1. Create feature branches from `main` or `develop`
2. Make changes and push to branch
3. Open pull request - automatic validation runs
4. After review and approval, merge to main
5. Tag releases trigger automated binary builds and GitHub releases

## Usage
The tool supports two authentication modes:

### Teams API Only (Basic Auth)
```bash
go run main.go -username <email> -api-token <api-token> -org-id <org-id> [-output <file>]
```

### Admin + Teams API (OAuth + Basic Auth)
```bash
go run main.go -access-token <oauth-token> -username <email> -api-token <api-token> -org-id <org-id> [-output <file>]
```

Required parameters:
- org-id: Organization ID to query
- Either access-token (for admin API) OR username+api-token (for teams API)

Optional parameters:
- admin-base-url: Admin API base URL (default: https://api.atlassian.com)
- output: Output file path (default: stdout)

## Output Format
The tool outputs NDJSON (newline-delimited JSON) with each line containing:
- type: "organization" or "team"
- timestamp: ISO 8601 timestamp
- data: The actual entity data with all metadata