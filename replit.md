# Atlassian Organization Tool

## Overview
A Go CLI tool that queries Atlassian organization and team structure APIs and outputs NDJSON metadata for consumption by other tools.

## Recent Changes
- Created main.go with complete Atlassian API client implementation
- Added command-line argument parsing for API credentials
- Implemented NDJSON streaming output functionality
- Added organization and team structure walking with member details
- Included rate limiting and error handling

## Project Architecture
- Single Go binary with minimal dependencies
- Uses standard library for HTTP requests and JSON handling
- Modular design with separate structs for different API entities
- Configurable output (stdout or file) for pipeline integration

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