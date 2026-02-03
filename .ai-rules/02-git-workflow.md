# Git Workflow Rules

## Commit Guidelines

### Atomic Commits

**One Logical Change Per Commit**
- Each commit should represent a single, complete change
- Should be reversible without breaking the codebase
- Easy to review and understand

**Examples**
```bash
# Good - atomic
git commit -m "feat: add GET endpoint for KV retrieval"
git commit -m "test: add unit tests for GET handler"
git commit -m "docs: update API specification with GET endpoint"

# Bad - multiple changes
git commit -m "add GET endpoint, fix bug, update docs"
```

### Conventional Commits

**Format**
```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

**Types**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `test`: Adding or updating tests
- `refactor`: Code restructuring without behavior change
- `perf`: Performance improvements
- `style`: Code style changes (formatting, no logic change)
- `chore`: Maintenance tasks (dependencies, build)
- `ci`: CI/CD changes

**Scope** (optional)
- `handlers`: HTTP handlers
- `database`: Database layer
- `config`: Configuration
- `api`: API changes
- `kv`: KV interface

**Examples**
```bash
# Feature
feat(handlers): implement batch delete endpoint

# Bug fix
fix(database): resolve BBolt file locking issue

# Documentation
docs(api): add examples for batch operations

# Test
test(handlers): add integration tests for namespace management

# Refactor
refactor(kv): extract validation logic to helper function

# Performance
perf(handlers): optimize batch operation memory usage
```

### Commit Message Best Practices

**Subject Line**
- Max 72 characters
- Imperative mood: "add" not "added" or "adds"
- No period at the end
- Be specific and descriptive

**Body** (optional, for complex changes)
```
feat(handlers): implement batch set operation

Add support for setting multiple key-value pairs in a single request.
This reduces network overhead and improves performance for bulk operations.

- Support up to 1000 operations per batch
- Return detailed results for each operation
- Handle partial failures gracefully
```

**Footer** (for breaking changes or issue references)
```
feat(api): change error response format

BREAKING CHANGE: Error responses now use "code" instead of "error_code"

Closes #123
```

## Git Workflow

### Branch Strategy

**Main Branches**
- `main`: Production-ready code
- `dev`: Development branch (current work)

**Feature Branches**
- Create from `dev`
- Name: `feature/description` or `fix/description`
- Example: `feature/prometheus-metrics`

**Workflow**
```bash
# Start feature
git checkout dev
git pull origin dev
git checkout -b feature/new-feature

# Work and commit
git add .
git commit -m "feat: add new feature"

# Keep updated
git fetch origin
git rebase origin/dev

# Push
git push origin feature/new-feature

# Create PR to dev
```

### Before Committing

**Checklist**
1. [ ] Code compiles: `go build ./...`
2. [ ] Tests pass: `go test ./...`
3. [ ] Linting clean: `golangci-lint run`
4. [ ] Tests added for new code
5. [ ] Documentation updated
6. [ ] No secrets in code
7. [ ] Commit message follows conventions

**Commands**
```bash
# Check status
git status

# Stage files
git add <file>
git add .  # or all files

# Commit
git commit -m "type(scope): description"

# Verify
git log --oneline -1
```

### Git Commands for Commander

**Check Changes**
```bash
# See what changed
git status
git diff

# See staged changes
git diff --cached
```

**Commit Process**
```bash
# Stage specific files
git add internal/handlers/kv.go
git add internal/handlers/kv_test.go

# Commit
git commit -m "feat(handlers): add KV CRUD handlers

- Implement GET, POST, DELETE, HEAD endpoints
- Add parameter validation
- Include comprehensive error handling
- Add unit tests with 80%+ coverage"

# Push
git push origin feature/kv-crud
```

**Amend Last Commit** (if not pushed)
```bash
# Fix typo or add forgotten file
git add forgotten-file.go
git commit --amend --no-edit

# Change commit message
git commit --amend -m "better message"
```

**Undo Changes**
```bash
# Unstage file
git reset HEAD <file>

# Discard changes
git checkout -- <file>

# Undo last commit (keep changes)
git reset --soft HEAD~1

# Undo last commit (discard changes) - DANGEROUS
git reset --hard HEAD~1
```

## Commit Frequency

### When to Commit

**Commit After**
- Implementing a complete function
- Fixing a bug
- Adding tests for a feature
- Updating documentation
- Completing a logical unit of work

**Don't Commit**
- Broken code (unless marked WIP)
- Incomplete features (unless on feature branch)
- Generated files (binaries, coverage reports)
- Sensitive data (.env files)

**Example Flow**
```bash
# 1. Implement feature
git add internal/handlers/batch.go
git commit -m "feat(handlers): implement batch set handler"

# 2. Add tests
git add internal/handlers/batch_test.go
git commit -m "test(handlers): add batch set handler tests"

# 3. Update docs
git add docs/api-specification.yaml
git commit -m "docs(api): add batch set endpoint to specification"
```

## Commander-Specific Rules

### Commit Message Examples from Project

```bash
# From Phase 1
git commit -m "feat: implement KV CRUD API endpoints for /api/v1

- Implement GET /api/v1/kv/{namespace}/{collection}/{key} to retrieve values
- Implement POST /api/v1/kv/{namespace}/{collection}/{key} to set values
- Implement DELETE /api/v1/kv/{namespace}/{collection}/{key} to remove keys
- Implement HEAD /api/v1/kv/{namespace}/{collection}/{key} to check key existence
- Add request/response structures with proper error handling
- Add comprehensive unit tests for all CRUD operations
- Validate input parameters and normalize namespace
- Return standardized JSON responses with timestamps
- Achieve 81.8% test coverage for handlers package"
```

### Multi-Line Messages

**When to Use**
- Implementing multiple related changes
- Need to explain rationale
- Breaking changes
- Complex refactoring

**Format**
```bash
git commit -m "feat(database): add Redis backend support

Implement Redis as an alternative KV storage backend alongside BBolt and MongoDB.

Key features:
- Connection pooling with configurable size
- Key format: namespace:collection:key
- Automatic JSON serialization
- Context-aware operations with timeout support

Performance improvements:
- 10x faster than MongoDB for simple key-value operations
- Sub-millisecond response times for cached data

Configuration:
- REDIS_URI environment variable
- Supports authentication and TLS

Closes #45"
```

## Git Hooks (Recommended)

### Pre-commit Hook
```bash
#!/bin/sh
# .git/hooks/pre-commit

# Run tests
go test ./... || exit 1

# Run linter
golangci-lint run || exit 1

# Check for secrets
if git diff --cached | grep -i "password\|secret\|token\|api_key"; then
    echo "⚠️  Warning: Possible secret in commit"
    exit 1
fi
```

### Commit Message Hook
```bash
#!/bin/sh
# .git/hooks/commit-msg

# Check commit message format
commit_msg=$(cat "$1")
if ! echo "$commit_msg" | grep -qE "^(feat|fix|docs|test|refactor|perf|style|chore|ci)(\(.+\))?: .+"; then
    echo "❌ Invalid commit message format"
    echo "Use: type(scope): description"
    exit 1
fi
```

## References

- [Conventional Commits](https://www.conventionalcommits.org/)
- [Git Best Practices](https://git-scm.com/book/en/v2/Distributed-Git-Contributing-to-a-Project)
- [Atomic Commits](https://www.aleksandrhovhannisyan.com/blog/atomic-git-commits/)
