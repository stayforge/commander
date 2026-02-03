# AI Development Rules

This directory contains detailed development rules for Commander project, organized by topic to optimize token usage.

## Structure

```
.clinerules              # Main index with quick reference
.ai-rules/
├── README.md            # This file
├── 01-code-style.md     # Go best practices
├── 02-git-workflow.md   # Commit conventions
├── 03-testing.md        # Testing patterns
├── 04-api-design.md     # REST API design
├── 05-database.md       # Database patterns
├── 06-documentation.md  # Documentation standards
├── 07-performance.md    # Performance optimization
└── 08-security.md       # Security practices
```

## Usage

### For AI Assistants

When working on Commander:

1. **Always read** `.clinerules` first (main index)
2. **Load specific rules** as needed:
   - Writing code? → `01-code-style.md`
   - Committing? → `02-git-workflow.md`
   - Adding tests? → `03-testing.md`
   - API work? → `04-api-design.md`
   - Database? → `05-database.md`
   - Documentation? → `06-documentation.md`
   - Performance? → `07-performance.md`
   - Security? → `08-security.md`

3. **Follow universal rules** in `.clinerules` at all times

### For Developers

Browse these files to understand:
- Project standards and conventions
- Best practices and patterns
- Testing requirements
- Documentation expectations
- Performance targets
- Security guidelines

## Rule Categories

### 1. Code Style (01-code-style.md)
- Go naming conventions
- File organization
- Function guidelines
- Error handling patterns
- Comments and formatting
- Project-specific patterns

### 2. Git Workflow (02-git-workflow.md)
- Atomic commit strategy
- Conventional commit format
- Branching strategy
- Commit message examples
- Pre-commit checklist

### 3. Testing (03-testing.md)
- Coverage requirements (85%+)
- Table-driven test pattern
- MockKV usage
- Handler testing
- Benchmarking
- Integration tests

### 4. API Design (04-api-design.md)
- RESTful principles
- URL structure
- Request/response format
- Status codes
- Error handling
- Gin handler pattern

### 5. Database (05-database.md)
- KV interface implementation
- BBolt, Redis, MongoDB patterns
- Factory pattern
- Data organization
- Context usage
- Transaction handling

### 6. Documentation (06-documentation.md)
- Code documentation (godoc)
- API documentation (OpenAPI)
- README structure
- Code examples
- Changelog format
- TODO comments

### 7. Performance (07-performance.md)
- Edge device optimization
- Binary size reduction
- Memory management
- I/O optimization
- Profiling techniques
- Load testing

### 8. Security (08-security.md)
- Input validation
- Error message safety
- Secrets management
- Rate limiting
- Authentication patterns
- Security checklist

## Design Philosophy

### Token Efficiency

Instead of one large file:
- **Modular**: Load only what you need
- **Focused**: Each file covers one topic
- **Indexed**: Quick reference in main file
- **Searchable**: Clear structure

### Comprehensive Coverage

All aspects covered:
- ✅ Code quality
- ✅ Git workflow
- ✅ Testing
- ✅ API design
- ✅ Database patterns
- ✅ Documentation
- ✅ Performance
- ✅ Security

### Practical Examples

Every rule includes:
- Clear explanation
- Code examples
- Good/bad patterns
- Real-world scenarios
- Commander-specific guidance

## Quick Reference

### Before Writing Code
1. Read `.clinerules`
2. Load relevant rule file(s)
3. Follow established patterns
4. Write tests

### Before Committing
1. Check `02-git-workflow.md`
2. Run: `go test ./...`
3. Run: `golangci-lint run`
4. Use conventional commit format
5. One logical change per commit

### Before Documentation
1. Check `06-documentation.md`
2. Update code comments
3. Update API spec
4. Add examples
5. Test examples

### Before Deployment
1. Check `08-security.md`
2. Review security checklist
3. Run performance tests
4. Update documentation

## Maintenance

### Adding Rules

When adding new rules:
1. Choose appropriate category
2. Follow existing format
3. Include examples
4. Update this README
5. Update `.clinerules` index

### Updating Rules

When updating:
1. Keep examples current
2. Test code examples
3. Maintain consistency
4. Update version in `.clinerules`

## Statistics

- **Total Files**: 9 (1 index + 8 rule files)
- **Total Lines**: ~3,800
- **Average per File**: ~470 lines
- **Topics Covered**: 8 categories
- **Code Examples**: 100+
- **Best Practices**: 200+

## Benefits

### For AI Assistants
- Load only relevant rules
- Reduce token usage
- Focus on specific task
- Consistent behavior

### For Developers
- Clear standards
- Easy reference
- Comprehensive coverage
- Real examples

### For Project
- Consistent quality
- Faster onboarding
- Better code review
- Maintainable codebase

## Version

**Version**: 1.0.0  
**Last Updated**: 2026-02-03  
**Status**: Active

## Related Documentation

- **Main README**: `../README.md`
- **Project Plan**: `../docs/PROJECT_MANAGEMENT_PLAN.md`
- **API Docs**: `../docs/api-specification.yaml`
- **Phase Report**: `../docs/PHASE1_COMPLETION.md`

---

**Note**: These rules are living documents. Update them as the project evolves.
