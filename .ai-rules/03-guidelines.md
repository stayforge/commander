# Development Guidelines

## Git Workflow

**Commit format**: `type(scope): description`
- Types: feat, fix, docs, test, refactor, perf, chore, ci
- Scopes: handlers, database, config, api, kv
- Subject: max 72 chars, imperative mood, no period

**Before committing**:
1. `go build ./...`
2. `go test ./...`
3. `golangci-lint run`
4. No secrets in code

**Branch strategy**: `main` (production) ← `dev` (development) ← `feature/*` or `fix/*`

## Security

- Secrets via env vars only (never hardcode)
- Generic error messages to API clients; log details server-side
- Validate all user input (params, body, query)
- BBolt file permissions: 0600
- TLS in production

## Performance (Edge Devices)

- Target: <50ms p99, <20MB binary, <100MB RAM, <1s startup
- Pre-allocate slices: `make([]T, 0, len(items))`
- Batch BBolt writes in single transaction
- Set timeouts on all operations: `context.WithTimeout`
- Build optimized: `go build -ldflags="-s -w" -trimpath`

## Documentation

- Exported functions need godoc comments
- Update `docs/api-specification.yaml` when changing endpoints
- Use TODO format: `// TODO: description` or `// TODO(name): description`
