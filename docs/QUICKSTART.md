# Quick Start Guide - Optimized Gitmit

## What's New

Gitmit has been optimized with intelligent commit message generation that analyzes your code changes and suggests contextually accurate commit messages - all working **100% locally** without internet or AI.

## Installation

```bash
# Build from source
go build -o bin/gitmit

# Or use existing binary
cd bin
./gitmit --version
```

## Basic Usage

### 1. Stage Your Changes
```bash
git add .
# or
git add specific-file.go
```

### 2. Generate Commit Message
```bash
./bin/gitmit propose
```

Example output:
```
feat(parser): add ParseCommitHistory for commit history parsing

Copy the message above and use it to commit.
```

### 3. Commit Automatically (Optional)
```bash
./bin/gitmit propose --auto
```

## Smart Features

### Function Detection
When you add a new function:
```go
// internal/parser/git.go
+func ParseCommitHistory() ([]Commit, error) {
```

**Result**: `feat(parser): add ParseCommitHistory for commit history parsing`

### Multi-File Intelligence
When fixing a bug across multiple files:
```bash
git add internal/handler/user.go internal/service/user.go
./bin/gitmit propose
```

**Result**: `fix(handler,service): resolve issue across multiple components`

### Pattern Recognition
The tool automatically detects:
- New functions/structs/methods
- Error handling additions
- Test updates
- API changes
- Database modifications
- Performance optimizations
- Security updates
- Documentation changes

## Command Options

```bash
# Preview without committing
./bin/gitmit propose --dry-run

# Auto-commit with generated message
./bin/gitmit propose --auto

# Summary only (no color)
./bin/gitmit propose --summary
```

## Customization

### Project-Specific Configuration

Create `.commit_suggest.json` in your project root:

```json
{
  "topicMappings": {
    "controllers": "api",
    "models": "database",
    "views": "ui"
  },
  "keywordMappings": {
    "authenticate": "user authentication",
    "authorize": "access control"
  }
}
```

### Custom Templates

Place `templates.json` next to the executable:

```json
{
  "A": {
    "mymodule": [
      "feat(mymodule): add {item} for {purpose}",
      "feat(mymodule): implement {item}"
    ]
  },
  "M": {
    "mymodule": [
      "fix(mymodule): resolve {purpose} in {item}",
      "refactor(mymodule): improve {item}"
    ]
  }
}
```

## Examples

### Example 1: New Feature
```bash
# Add new authentication handler
git add internal/handler/auth.go
./bin/gitmit propose
```
**Output**: `feat(handler): add AuthHandler for user authentication`

### Example 2: Bug Fix
```bash
# Fix validation bug
git add internal/validator/user.go
./bin/gitmit propose
```
**Output**: `fix(validator): correct issue related to UserValidator`

### Example 3: Refactoring
```bash
# Refactor multiple files
git add internal/parser/*.go
./bin/gitmit propose
```
**Output**: `refactor(parser): improve code organization`

### Example 4: Tests
```bash
# Add test suite
git add internal/*_test.go
./bin/gitmit propose
```
**Output**: `test(core): add tests for new functionality`

### Example 5: Documentation
```bash
# Update README
git add README.md GUIDE.md
./bin/gitmit propose
```
**Output**: `docs: update documentation`

## How It Works

### Intelligence Behind the Scenes

1. **Analyzes Git Diff**
   - Extracts function/struct/method names
   - Detects code patterns (error handling, tests, API changes)
   - Identifies file types and purposes

2. **Scores Templates**
   - Matches patterns to templates
   - Considers code structures detected
   - Evaluates file context
   - Selects best-fitting template

3. **Generates Message**
   - Replaces placeholders with actual values
   - Applies intelligent scope detection
   - Formats according to Conventional Commits

### What Makes It Smart

- ✅ Extracts actual function/struct names from code
- ✅ Detects 15+ change patterns automatically
- ✅ Recognizes multi-file patterns (features, bug fixes, refactors)
- ✅ Intelligent scope based on directory structure
- ✅ Context-aware template selection with scoring
- ✅ Avoids duplicate messages from history
- ✅ 100% local, no network or AI required

## Tips for Best Results

### 1. Use Meaningful Names
```go
// Good - results in clear commit messages
func ValidateUserCredentials() error

// Less clear
func validate() error
```

### 2. Stage Related Changes Together
```bash
# Good - detects feature addition pattern
git add internal/handler/user.go internal/service/user.go

# Less context
git add internal/handler/user.go
./bin/gitmit propose
git add internal/service/user.go
./bin/gitmit propose
```

### 3. Add Context in Code
```go
// The analyzer detects documentation patterns
// Add user authentication endpoint
func HandleLogin() {
    // implementation
}
```

### 4. Organize Files Clearly
```
internal/
  ├── handler/    → topic: handler
  ├── service/    → topic: service
  └── repository/ → topic: repository
```

## Commit Types Supported

| Type | When Used | Example |
|------|-----------|---------|
| feat | New features | `feat(api): add UserHandler` |
| fix | Bug fixes | `fix(auth): correct token validation` |
| refactor | Code improvements | `refactor(parser): improve logic` |
| perf | Performance | `perf(db): optimize query` |
| security | Security fixes | `security(auth): fix vulnerability` |
| test | Tests | `test(api): add test coverage` |
| docs | Documentation | `docs: update README` |
| style | Formatting | `style(core): format code` |
| chore | Maintenance | `chore(config): update settings` |
| ci | CI/CD | `ci: update workflow` |

## Troubleshooting

### Getting Generic Messages?
- Ensure meaningful function/struct names in code
- Add clear keywords (fix, bug, optimize, etc.)
- Stage related files together
- Check `.commit_suggest.json` for custom mappings

### Wrong Topic Detected?
- Use consistent directory structure
- Add topic mapping in `.commit_suggest.json`
- Keep related files in same directories

### Incorrect Action Type?
- Use clear keywords: "fix", "bug", "optimize", "security"
- Name test files with `_test.go` suffix
- Group related changes in commits

## Learn More

- **[OPTIMIZATION.md](OPTIMIZATION.md)** - Deep dive into optimizations
- **[TEMPLATE_REFERENCE.md](TEMPLATE_REFERENCE.md)** - Template placeholder guide
- **[OPTIMIZATION_SUMMARY.md](OPTIMIZATION_SUMMARY.md)** - Quick optimization overview
- **[README.md](README.md)** - Original documentation
- **[CONTRIBUTING.md](CONTRIBUTING.md)** - Contribution guidelines

## Performance

- ⚡ Analysis: Milliseconds
- 💾 Memory: Minimal footprint
- 🌐 Network: Zero (100% offline)
- 🤖 AI: None (pure algorithms)
- 📦 Dependencies: Minimal (cobra, color)

## Why It's Better

### Traditional Approach
- Generic messages: "Update files"
- Manual thinking required
- Inconsistent style
- Time-consuming

### Optimized Gitmit
- Specific messages: "feat(parser): add ParseCommitHistory for commit history"
- Automatic analysis
- Conventional Commits standard
- Instant suggestions

## Support

For issues or questions:
1. Check documentation files
2. Review examples in OPTIMIZATION.md
3. Verify configuration in `.commit_suggest.json`
4. Check template customization in TEMPLATE_REFERENCE.md

---

**Built for developers who want smart commit messages without AI or internet dependencies.**
