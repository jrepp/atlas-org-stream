# Agent Workflow Guide

This document provides a concise workflow for AI agents working on this project to minimize token consumption while maintaining code quality.

## Quick Context

**Project**: Go CLI tool for Atlassian organization/team data extraction  
**Output**: NDJSON streaming format with schema validation  
**Structure**: Clean internal packages (`types`, `client`, `stats`, `schema`)

## Essential Commands

```bash
# Build & Test
make build test

# Code Quality (required before commit)
make fmt vet lint

# Full CI check
make ci
```

## Development Workflow

### 1. Making Changes
- **Read first**: Use `read` tool to understand existing code structure
- **Small changes**: Edit existing files, avoid creating new ones unless necessary
- **Follow patterns**: Maintain existing code patterns and structure

### 2. Testing
```bash
# Quick test
make test

# Full validation
make ci
```

### 3. Code Quality
Always run before committing:
```bash
make fmt      # Format code
make vet      # Static analysis
make lint     # Linting
```

## Key Files

- `main.go` - CLI entry point (186 lines, imports internal packages)
- `internal/types/` - Data structures
- `internal/client/` - HTTP client with retry logic
- `internal/stats/` - Statistics tracking
- `internal/schema/` - JSON schema generation
- `Makefile` - Build automation with all targets

## Package Structure Rules

- **internal/types**: All structs, no business logic
- **internal/client**: HTTP operations, authentication, retries
- **internal/stats**: Statistics tracking with atomic operations
- **internal/schema**: JSON schema generation only
- **main.go**: CLI coordination, minimal business logic

## Common Tasks

### Adding New Functionality
1. Identify correct internal package
2. Add to existing file if related, new file if distinct
3. Add tests in `*_test.go` files
4. Update main.go imports if needed

### Fixing Bugs
1. Identify affected package
2. Write test to reproduce issue
3. Fix code
4. Verify `make ci` passes

### Performance Issues
- Check retry logic in `internal/client`
- Verify statistics tracking efficiency
- Review HTTP client timeouts

## Testing Strategy

- **Unit tests**: All internal packages have `*_test.go`
- **Mock servers**: Use `httptest` for HTTP client testing
- **Coverage**: Target >80% coverage
- **Race detection**: `make test-race`

## GitHub Workflows

- **CI**: Runs on push/PR with multi-Go version testing
- **PR Validation**: Format check, lint, test
- **Release**: Auto-release on tag push with multi-platform binaries

## Error Patterns

- Use `fmt.Errorf` with `%w` for error wrapping
- Log to stderr for user feedback
- Return structured errors from internal packages
- Handle retries in client package only

## Common Anti-Patterns to Avoid

- ❌ Adding business logic to main.go
- ❌ Direct HTTP calls outside client package
- ❌ Manual statistics tracking outside stats package
- ❌ Creating files in root directory
- ❌ Modifying GitHub workflows without testing

## Quick Reference

```bash
# Essential files to check when debugging
./main.go                          # CLI logic
./internal/client/atlassian.go     # HTTP operations
./internal/stats/tracker.go        # Statistics
./Makefile                         # Build targets

# Test specific package
go test -v ./internal/client/

# Check what changed
git status
git diff

# Verify build
make build && ./build/atlassian-org-tool --help
```

## Token Optimization Tips

- Use `search_codebase` with specific queries instead of reading multiple files
- Read files with `limit` parameter for large files
- Use `grep` for finding specific patterns
- Batch related operations in single responses
- Focus on specific packages when making changes