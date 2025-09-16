# Atlassian Organization Tool

[![CI/CD Pipeline](https://github.com/jrepp/atlas-org-stream/workflows/CI%2FCD%20Pipeline/badge.svg)](https://github.com/jrepp/atlas-org-stream/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/jrepp/atlas-org-stream)](https://goreportcard.com/report/github.com/jrepp/atlas-org-stream)
[![Go Version](https://img.shields.io/github/go-mod/go-version/jrepp/atlas-org-stream)](https://golang.org/)
[![Release](https://img.shields.io/github/release/jrepp/atlas-org-stream.svg)](https://github.com/jrepp/atlas-org-stream/releases/latest)
[![License](https://img.shields.io/github/license/jrepp/atlas-org-stream)](LICENSE)

A high-performance Go CLI tool that queries Atlassian organization and team structure APIs, outputting comprehensive metadata as streaming NDJSON for consumption by other tools and pipelines.

## 🚀 Features

- **Dual Authentication**: Supports both OAuth 2.0 (admin APIs) and Basic Auth (teams APIs)
- **Streaming Output**: NDJSON format with JSON schema validation
- **Comprehensive Statistics**: Real-time processing metrics with periodic updates
- **Robust Error Handling**: Exponential backoff, retry logic, and rate limiting support
- **Production Ready**: Extensive logging, error recovery, and pipeline integration

## 📦 Installation

### Download Pre-built Binaries

Download the latest release for your platform from the [releases page](https://github.com/jrepp/atlas-org-stream/releases/latest):

```bash
# Linux/macOS
curl -L https://github.com/jrepp/atlas-org-stream/releases/latest/download/atlassian-org-tool-linux-amd64 -o atlassian-org-tool
chmod +x atlassian-org-tool

# Windows (PowerShell)
Invoke-WebRequest -Uri "https://github.com/jrepp/atlas-org-stream/releases/latest/download/atlassian-org-tool-windows-amd64.exe" -OutFile "atlassian-org-tool.exe"
```

### Build from Source

```bash
git clone https://github.com/jrepp/atlas-org-stream.git
cd atlas-org-stream
make build

# Or using Go directly
go build -o atlassian-org-tool .
```

### Install with Go

```bash
go install github.com/jrepp/atlas-org-stream@latest
```

## 🔧 Quick Start

### Prerequisites

You'll need Atlassian API credentials. Choose one of these authentication methods:

1. **Teams API Only** (Basic Auth): Username + API Token
2. **Admin + Teams API** (OAuth + Basic Auth): OAuth Access Token + Username + API Token

### Basic Usage

```bash
# Teams API only (Basic Auth)
./atlassian-org-tool \
  -username your-email@company.com \
  -api-token your-api-token \
  -org-id your-org-id

# Admin + Teams API (OAuth + Basic Auth)  
./atlassian-org-tool \
  -access-token your-oauth-token \
  -username your-email@company.com \
  -api-token your-api-token \
  -org-id your-org-id

# Save to file
./atlassian-org-tool \
  -username your-email@company.com \
  -api-token your-api-token \
  -org-id your-org-id \
  -output organization-data.ndjson
```

## 📊 Output Format

The tool outputs NDJSON (newline-delimited JSON) with the following structure:

```json
{"type":"schema","timestamp":"2024-01-01T00:00:00Z","data":{...}}
{"type":"organization","timestamp":"2024-01-01T00:00:01Z","data":{...}}
{"type":"team","timestamp":"2024-01-01T00:00:02Z","data":{...}}
{"type":"statistics","timestamp":"2024-01-01T00:00:03Z","data":{...}}
```

### Data Types

- **`schema`**: JSON schema describing all subsequent data structures
- **`organization`**: Organization metadata (requires OAuth token)
- **`team`**: Team data with member information
- **`statistics`**: Processing statistics (periodic and final)

## 🛠️ Development

### Prerequisites

- Go 1.21 or later
- Make (optional, but recommended)

### Setup

```bash
git clone https://github.com/jrepp/atlas-org-stream.git
cd atlas-org-stream

# Install dependencies
make deps

# Run tests
make test

# Build
make build
```

### Available Make Targets

```bash
# Building
make build              # Build for current platform
make build-all          # Build for all platforms (Linux, Windows, macOS)
make build-linux        # Build for Linux
make build-windows      # Build for Windows
make build-darwin       # Build for macOS

# Testing
make test               # Run unit tests
make test-coverage      # Run tests with coverage report
make test-race          # Run tests with race detector
make test-bench         # Run benchmark tests

# Code Quality
make fmt                # Format code
make fmt-check          # Check code formatting
make vet                # Run go vet
make lint               # Run golangci-lint
make staticcheck        # Run staticcheck
make security           # Run security scan

# Development
make run                # Build and run with --help
make install            # Install to $GOPATH/bin
make clean              # Clean build artifacts

# CI/CD
make ci                 # Run all CI checks
```

### Project Structure

```
├── main.go                    # CLI entry point
├── internal/                  # Internal packages
│   ├── types/                 # Data structures and types
│   ├── client/                # HTTP client and API methods
│   ├── stats/                 # Statistics tracking
│   └── schema/                # JSON schema generation
├── .github/workflows/         # GitHub Actions workflows
├── Makefile                   # Build automation
└── README.md                  # This file
```

### Running Tests

```bash
# All tests
make test

# With coverage
make test-coverage
open coverage.html

# Race detection
make test-race

# Individual packages
go test -v ./internal/types/
go test -v ./internal/client/
```

## 🔐 Authentication Setup

### API Token (Basic Auth)

1. Go to [Atlassian Account Settings](https://id.atlassian.com/manage-profile/security/api-tokens)
2. Click "Create API token"
3. Copy the generated token

### OAuth Access Token (Admin API)

1. Create an OAuth app in your Atlassian organization
2. Obtain an access token with `read:organization` scope
3. Use this token for organization-level queries

## 📈 Usage Examples

### Pipeline Integration

```bash
# Process data with jq
./atlassian-org-tool -username user@company.com -api-token token -org-id org123 | \
  jq -r 'select(.type == "team") | .data.displayName'

# Save statistics only
./atlassian-org-tool -username user@company.com -api-token token -org-id org123 | \
  jq -r 'select(.type == "statistics")'

# Count team members
./atlassian-org-tool -username user@company.com -api-token token -org-id org123 | \
  jq -r 'select(.type == "team") | .data.members | length' | \
  awk '{sum += $1} END {print "Total members:", sum}'
```

### Error Handling

The tool provides comprehensive error reporting:

```bash
# Stderr shows human-readable progress and statistics
./atlassian-org-tool -username user@company.com -api-token token -org-id org123 2>progress.log

# Final statistics are always output, even on failure
echo $?  # Exit code: 0 = success, 1 = failure
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes and add tests
4. Run quality checks (`make ci`)
5. Commit your changes (`git commit -m 'Add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

### Code Quality Standards

- All code must pass `make ci` (formatting, linting, tests, security)
- Maintain test coverage above 80%
- Follow [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Add tests for new functionality

## 📋 Command Reference

### Required Parameters

- `-org-id string`: Atlassian organization ID

### Authentication Parameters

Must provide **one** of these combinations:

- `-username` + `-api-token`: For teams API access
- `-access-token` + `-username` + `-api-token`: For admin + teams API access

### Optional Parameters

- `-admin-base-url string`: Admin API base URL (default: `https://api.atlassian.com`)
- `-output string`: Output file path (default: stdout)

### Examples

```bash
# Minimal usage (teams only)
./atlassian-org-tool -username user@company.com -api-token abc123 -org-id org456

# Full access (admin + teams)
./atlassian-org-tool \
  -access-token oauth_token_here \
  -username user@company.com \
  -api-token abc123 \
  -org-id org456 \
  -output complete-org-data.ndjson

# Custom admin URL
./atlassian-org-tool \
  -admin-base-url https://custom.atlassian.net \
  -username user@company.com \
  -api-token abc123 \
  -org-id org456
```

## 🐛 Troubleshooting

### Common Issues

**Authentication Errors (401)**
- Verify your API token is valid and not expired
- Ensure your username/email is correct
- Check that you have appropriate permissions

**Rate Limiting (429)**
- The tool automatically handles rate limiting with exponential backoff
- For large organizations, consider running during off-peak hours

**Network Timeouts**
- The tool uses 30-second timeouts with automatic retries
- Check your network connection and firewall settings

### Debug Output

Enable verbose logging by redirecting stderr:

```bash
./atlassian-org-tool [args] 2>debug.log
tail -f debug.log
```

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🔗 Links

- [GitHub Repository](https://github.com/jrepp/atlas-org-stream)
- [Issues](https://github.com/jrepp/atlas-org-stream/issues)
- [Releases](https://github.com/jrepp/atlas-org-stream/releases)
- [Atlassian API Documentation](https://developer.atlassian.com/cloud/admin/about/)