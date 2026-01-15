# Template Placeholders Reference

## Available Placeholders

Templates in `templates.json` support the following placeholders:

### Core Placeholders

| Placeholder | Description | Example Value | Source |
|------------|-------------|---------------|---------|
| `{topic}` | Module or directory name | `parser`, `api`, `auth` | File path analysis |
| `{item}` | Specific code element | `ParseCommitHistory`, `UserValidator` | Function/struct/method detection or filename |
| `{purpose}` | Inferred intent | `authentication`, `database query`, `validation` | Keyword analysis from diff |
| `{source}` | Original file name (renames) | `old_parser.go` | Git rename detection |
| `{target}` | New file name (renames) | `new_parser.go` | Git rename detection |

## Placeholder Resolution

### {topic} Resolution Priority
1. Custom mapping from `.commit_suggest.json`
2. Second-level directory (e.g., `internal/parser` → `parser`)
3. First-level directory
4. Default: `core`

### {item} Resolution Priority
1. Detected function names
2. Detected struct names
3. Detected method names
4. Filename without extension

### {purpose} Detection

The system detects purpose from these keywords:

| Keyword in Diff | Purpose |
|----------------|---------|
| login | authentication |
| validate | validation |
| query | database query |
| cache | caching |
| refactor | code restructuring |
| logging, log | logging |
| docs | documentation |
| middleware | middleware |
| test | testing |
| config | configuration |
| ci | ci/cd |
| sql, gorm | database logic |
| feat | new feature |
| bug, fix | bug fix |
| cleanup | code cleanup |
| perf | performance improvement |
| security | security update |
| dep | dependency update |
| build | build system |
| style | code style |

**Default**: `general update`

## Template Examples

### Good Templates

```json
{
  "feat(api): add {item} handler",
  "feat(auth): implement {item} for {purpose}",
  "fix(db): correct {purpose} in {item}",
  "refactor({topic}): improve {item} implementation",
  "test({topic}): add tests for {item}",
  "perf({topic}): optimize {item} performance"
}
```

### Template Scoring

Templates are scored based on:

1. **Pattern Match** (+2.0)
   - Template keywords match detected patterns
   - Example: "api" in template when `api-changes` detected

2. **Structure Usage** (+1.5)
   - Uses `{item}` when functions/structs detected
   - Ensures actual code names in commit messages

3. **Purpose Relevance** (+1.0)
   - Uses `{purpose}` when meaningful purpose detected
   - Not when purpose is "general update"

4. **File Type Context** (+0.5 to +1.5)
   - `.go` + "func" in template: +0.5
   - `.json/.yaml` + "config": +1.0
   - `.md` + "docs": +1.5

5. **Major Change Bonus** (+1.0)
   - "restructure", "refactor", "major" keywords
   - When `IsMajor` flag set (500+ line changes)

6. **Generic Penalty** (-0.5)
   - "general" keyword when specific patterns exist
   - Encourages specific messages

## Creating Custom Templates

### Project-Specific Templates

Create `templates.json` next to the executable:

```json
{
  "A": {
    "mymodule": [
      "feat(mymodule): add {item} for {purpose}",
      "feat(mymodule): implement {item} handler",
      "feat(mymodule): create {item} service"
    ]
  },
  "M": {
    "mymodule": [
      "fix(mymodule): resolve {purpose} in {item}",
      "refactor(mymodule): improve {item} logic",
      "perf(mymodule): optimize {item}"
    ]
  }
}
```

### Custom Mappings

Create `.commit_suggest.json` in project root:

```json
{
  "topicMappings": {
    "controllers": "api",
    "models": "database",
    "views": "ui",
    "routes": "api",
    "migrations": "db"
  },
  "keywordMappings": {
    "authenticate": "user authentication",
    "authorize": "access control",
    "sanitize": "input validation",
    "encrypt": "data encryption",
    "serialize": "data serialization"
  }
}
```

## Template Action Types

### Standard Actions

| Action | Template Key | Use Case | Example |
|--------|-------------|----------|---------|
| Add | `A` | New files | `feat(api): add UserHandler` |
| Modify | `M` | Changed files | `fix(auth): correct token validation` |
| Delete | `D` | Removed files | `chore(api): remove deprecated endpoint` |
| Rename | `R` | Renamed files | `refactor(core): rename oldFile to newFile` |

### Special Actions

| Action | Template Key | Use Case | Example |
|--------|-------------|----------|---------|
| Security | `SECURITY` | Security fixes | `security(auth): fix vulnerability in token validation` |
| Performance | `PERF` | Optimizations | `perf(db): optimize query performance` |
| Style | `STYLE` | Formatting | `style(core): format code for consistency` |
| Test | `TEST` | Test changes | `test(api): add tests for UserHandler` |
| Docs | `DOC` | Documentation | `docs(api): update API documentation` |
| Misc | `MISC` | Miscellaneous | `chore: general maintenance` |

## Pattern-Based Selection

When specific patterns are detected, certain templates get priority:

### Pattern → Template Preference

| Detected Pattern | Preferred Template Keywords | Score Bonus |
|-----------------|---------------------------|-------------|
| error-handling | fix, improve | +2.0 |
| test-addition | test, coverage | +2.0 |
| documentation | docs, clarify | +2.0 |
| api-changes | api, endpoint, handler | +2.0 |
| database | db, query, migration | +2.0 |
| security | security, auth, token | +2.0 |
| performance | perf, optimize, cache | +2.0 |
| refactoring | refactor, improve, clean | +2.0 |
| configuration | config, settings | +2.0 |

## Best Practices

### 1. Use Descriptive Placeholders
```json
// Good
"feat(api): implement {item} for {purpose}"

// Less descriptive
"feat: add {item}"
```

### 2. Provide Multiple Variations
```json
"api": [
  "feat(api): add {item} endpoint",
  "feat(api): implement {item} handler",
  "feat(api): create {item} route"
]
```

### 3. Match Your Project Structure
```json
// For microservices
"service": ["feat(service): add {item} microservice"]

// For monoliths
"module": ["feat(module): add {item} feature"]
```

### 4. Be Specific
```json
// Better
"feat(auth): implement {item} for secure authentication"

// Generic
"feat: add new functionality"
```

## Troubleshooting

### Placeholder Not Replaced
**Issue**: `{item}` shows as literal in message

**Solutions**:
- Ensure placeholder exactly matches: `{item}` not `{Item}` or `{ item }`
- Check that analyzer detects code structures
- Verify file has actual code changes (not just whitespace)

### Wrong Topic Selected
**Issue**: Topic doesn't match your project structure

**Solutions**:
- Add custom mapping in `.commit_suggest.json`
- Organize files in clear directory structure
- Check that files are in subdirectories (not root)

### Generic Purpose
**Issue**: Purpose is always "general update"

**Solutions**:
- Use meaningful keywords in code changes
- Add custom keywords in `.commit_suggest.json`
- Include comments explaining changes

---

For complete optimization details, see [OPTIMIZATION.md](OPTIMIZATION.md).
