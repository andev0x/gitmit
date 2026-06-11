# Gitmit Optimization Guide

## Overview

This document describes the intelligent optimizations implemented in Gitmit to provide smart, context-aware commit message suggestions based purely on git changes analysis and template matching - all working locally without any internet connection or AI.

## Key Optimizations

### 1. Intelligent Git Diff Analysis

The analyzer now performs deep pattern detection on git diffs:

#### Code Structure Detection
- **Function Detection**: Identifies new or modified function declarations
- **Struct Detection**: Recognizes struct definitions and changes
- **Method Detection**: Detects method additions and modifications
- **Import Analysis**: Tracks import changes to understand dependencies

#### Pattern Recognition
The system detects these change patterns:
- `error-handling`: Addition of error handling code
- `test-addition`: New test functions
- `import-changes`: Import modifications
- `documentation`: Comment additions (3+ comments)
- `refactoring`: Large-scale code changes (10+ adds/removes)
- `configuration`: Config file changes
- `api-changes`: HTTP/router/handler modifications
- `database`: SQL/GORM/query changes
- `performance`: Concurrency/optimization code
- `security`: Authentication/encryption changes

#### Enhanced Action Determination
The analyzer now considers:
- Security updates (security, vulnerability keywords)
- Performance improvements (optimize, cache, goroutine)
- Style changes (formatting, whitespace)
- Bug fixes (fix, bug, issue, resolve keywords)
- Test updates (test files)

### 2. Context-Aware Template Scoring

Templates are now scored based on multiple factors:

#### Scoring Criteria
- **Base Score**: 1.0 for all templates
- **Pattern Matching**: +2.0 for templates matching detected patterns
- **Structure Detection**: +1.5 for using detected functions/structs
- **Purpose Relevance**: +1.0 for meaningful purpose placeholders
- **File Type Context**: +0.5-1.5 based on file extensions
- **Major Changes**: +1.0 for restructure/refactor templates
- **Generic Penalty**: -0.5 for generic templates when specific info exists

#### Template Selection Process
1. Score all available templates
2. Sort by score (highest first)
3. Select highest-scored template not in recent history
4. Fall back to highest-scored if all are recent duplicates

### 3. Enhanced Template Library

The template library has been expanded with:

#### New Categories
- `handler`: For request handler changes
- `middleware`: For middleware modifications
- `service`: For service layer changes
- `util`: For utility functions
- `parser`: For parsing logic
- `analyzer`: For analysis code

#### New Action Types
- `SECURITY`: Security-focused commits
- `PERF`: Performance improvements
- `STYLE`: Code formatting changes
- `TEST`: Test modifications

#### More Variations
Each category now includes 3-6 template variations for better context matching.

### 4. Multi-File Pattern Detection

The system now analyzes patterns across multiple files:

#### Detected Patterns
- **feature-addition**: 3+ new files (60% of changes)
- **bug-fix-cascade**: 3+ modifications with "fix" keywords
- **refactor-sweep**: Mix of additions, modifications, and deletions (4+ files)
- **test-suite-update**: 70%+ test files changed
- **config-update**: 70%+ config files changed
- **api-redesign**: 3+ API/handler files modified
- **database-migration**: 2+ migration/schema files changed

#### Automatic Adjustments
When multi-file patterns are detected, the system automatically:
- Adjusts commit action type
- Updates purpose description
- Improves message context

### 5. Intelligent Scope Detection

Smart scope determination based on:

#### Single Topic
If all changes are in the same topic, use that topic as scope.

#### Single Directory
If changes span topics but share a directory, use directory as scope.

#### Multiple Related Topics
For 2-3 related topics, create combined scope (e.g., "api,service,handler").

#### Many Topics
For many topics, use most common topic or "core".

### 6. Enhanced Item Detection

Item selection now prioritizes:
1. Detected function names
2. Detected struct names
3. Detected method names
4. Original file-based item

This ensures commit messages reference actual code elements.

## How It Works

### Analysis Flow

```
Git Changes → Parser → Analyzer → Templater → Formatter → Commit Message
     ↓           ↓         ↓          ↓           ↓
  Diff Data   Changes   Patterns   Scoring    Final
              + Stats   + Context  + Select   Format
```

### Example: Adding a New Function

**Input**: New function in `internal/parser/git.go`
```go
+func ParseCommitHistory() ([]Commit, error) {
+    // implementation
+}
```

**Analysis**:
- Action: `A` (added)
- Detected Function: `ParseCommitHistory`
- Topic: `parser` (from directory)
- Pattern: `error-handling` (return error)
- Item: `ParseCommitHistory` (from function)

**Template Scoring**:
- "feat(parser): add {item} for {purpose}" → Score: 4.5
- "feat(parser): implement {item}" → Score: 4.0
- "feat: add {item} in {topic}" → Score: 2.5

**Result**: `feat(parser): add ParseCommitHistory for commit history parsing`

### Example: Bug Fix Across Multiple Files

**Input**: 
- Modified `internal/handler/user.go`
- Modified `internal/service/user.go`
- Modified `internal/validator/user.go`

All contain "fix" keywords.

**Analysis**:
- Multi-file pattern: `bug-fix-cascade`
- Action: `fix` (auto-adjusted)
- Scope: `handler,service,validator`
- Purpose: `resolve issue across multiple components` (auto-adjusted)

**Result**: `fix(handler,service,validator): resolve issue across multiple components`

## Configuration

### Custom Topic Mappings

Create `.commit_suggest.json` in your project root:

```json
{
  "topicMappings": {
    "controllers": "api",
    "models": "db",
    "views": "ui"
  },
  "keywordMappings": {
    "authenticate": "authentication",
    "authorize": "authorization",
    "sanitize": "input validation"
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
  }
}
```

## Performance

All optimizations run locally with:
- **No network calls**: 100% offline operation
- **Fast analysis**: Processes changes in milliseconds
- **Low memory**: Efficient pattern matching
- **No AI required**: Pure algorithmic approach

## Best Practices

### 1. Stage Related Changes Together
Group related changes in commits for better pattern detection:
```bash
git add internal/parser/*.go
gitmit propose
```

### 2. Use Meaningful Names
The system extracts function/struct names, so use descriptive names:
```go
// Good: func ValidateUserCredentials()
// Better for commit messages than: func validate()
```

### 3. Add Context Comments
The analyzer detects documentation patterns:
```go
// Adding validation for user input
func ValidateInput(input string) error {
    // implementation
}
```

### 4. Consistent File Organization
Better scope detection with organized structure:
```
internal/
  ├── handler/
  ├── service/
  └── repository/
```

## Testing the Optimizations

### Test Pattern Detection
```bash
# Add error handling
git add file-with-error-handling.go
./bin/gitmit propose --dry-run

# Add tests
git add *_test.go
./bin/gitmit propose --dry-run

# Multi-file refactor
git add internal/*/refactored-files.go
./bin/gitmit propose --dry-run
```

### Test Scope Detection
```bash
# Changes in single module
git add internal/parser/*.go
./bin/gitmit propose --dry-run

# Changes across modules
git add internal/*/*.go
./bin/gitmit propose --dry-run
```

## Future Enhancements

Potential improvements for local-only operation:
- Learn from project-specific commit history
- Build project-specific keyword dictionaries
- Detect naming conventions automatically
- Suggest breaking change indicators
- Integration with git hooks for validation
- Custom pattern definition files

## Troubleshooting

### Generic Messages
**Issue**: Getting generic commit messages
**Solution**: 
- Ensure meaningful function/struct names
- Add more context in code changes
- Stage related files together
- Check `.commit_suggest.json` for custom mappings

### Incorrect Action Type
**Issue**: Wrong commit type (feat vs fix)
**Solution**:
- Use clear keywords in changes ("fix", "bug", "optimize")
- Ensure proper file naming (e.g., `*_test.go` for tests)
- Check multi-file pattern detection

### Poor Scope Detection
**Issue**: Scope doesn't reflect actual changes
**Solution**:
- Organize files in clear directory structure
- Keep related changes in same directories
- Use consistent module naming

## Contributing

When adding new optimizations:
1. Keep everything local (no network calls)
2. Use algorithmic approaches (no AI/LLM)
3. Test with various commit types
4. Document pattern detection logic
5. Add template variations for new patterns

---

For more information, see [README.md](README.md) and [CONTRIBUTING.md](CONTRIBUTING.md).
