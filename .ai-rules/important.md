# Important Rules

These rules MUST be followed at all times. No exceptions.

## 1. English Only in Codebase

All code, comments, documentation, commit messages, error messages, and log messages MUST be written in **English only**. No other languages allowed in any file tracked by git.

Agent conversation with humans can use any language.

## 2. Never Commit Secrets

Never hardcode or commit secrets, credentials, API keys, or connection strings. Use environment variables via `.env` (which is gitignored).

## 3. Tests Required

All new code must have tests. Run before committing:
```bash
go test ./...
golangci-lint run
```

## 4. Conventional Commits

```
type(scope): description
```
Types: feat, fix, docs, test, refactor, perf, chore, ci

One logical change per commit.

## 5. Never Expose Internal Errors

API error responses must use generic messages. Log details server-side only.
