# Interactive Mode Guide

## Overview

Gitmit now features an enhanced interactive mode with intelligent commit message generation that works completely offline. No AI or APIs needed!

## How to Use

### 1. Stage Your Changes

```bash
git add .
```

### 2. Run Gitmit

```bash
gitmit
```

### 3. Choose Your Action

You'll see an interactive prompt with four options:

```
💡 Suggested commit message:
feat(api): implement user authentication strategy

Actions:
  y - Accept and commit
  n - Reject and exit
  e - Edit message manually
  r - Regenerate different suggestion

Choice [y/n/e/r]:
```

## Available Actions

### ✅ `y` - Accept and Commit

Press `y` (or just press Enter) to accept the suggested commit message and commit your changes immediately.

```
Choice [y/n/e/r]: y
✅ Changes committed successfully.
```

### ❌ `n` - Reject and Exit

Press `n` to reject the suggestion and exit without committing. Useful when you want to make more changes first.

```
Choice [y/n/e/r]: n
❌ Commit cancelled.
```

### ✏️ `e` - Edit Message Manually

Press `e` to edit the commit message manually. You can modify it to better fit your needs.

```
Choice [y/n/e/r]: e

📝 Edit the commit message:
Current: feat(api): implement user authentication strategy
New message: feat(api): add OAuth2 authentication with JWT tokens

✓ Updated commit message:
feat(api): add OAuth2 authentication with JWT tokens

Choice [y/n/e/r]:
```

After editing, you'll be presented with the options again.

### 🔄 `r` - Regenerate Different Suggestion

Press `r` to generate a completely different commit message. The system uses intelligent variation algorithms to provide diverse alternatives.

```
Choice [y/n/e/r]: r

💡 Alternative suggestion #1:
feat(api): integrate OAuth provider for secure access

Choice [y/n/e/r]: r

💡 Alternative suggestion #2:
feat(auth): add role-based access control for authentication

Choice [y/n/e/r]: y
✅ Changes committed successfully.
```

**Features of Regeneration:**
- Generates up to 10 different alternatives
- Uses similarity detection to ensure diversity
- Avoids showing the same message twice
- Maintains context relevance
- Smart template selection based on your changes

## Advanced Features

### Multiple Suggestions Mode

Get multiple ranked suggestions at once:

```bash
gitmit propose -s
```

Output:
```
💡 Ranked Suggestions:
1. feat(api): implement user authentication strategy (recommended)
2. feat(api): add token-based access via middleware
3. feat(auth): integrate OAuth provider for secure access
4. feat(api): expose new endpoint for authentication
5. feat(auth): implement MFA/2FA support for security
```

### Context Analysis

See what Gitmit detected in your changes:

```bash
gitmit propose --context
```

Output:
```
📊 Analysis Context:
Action: feat
Topic:  api
Item:   handler
Purpose: authentication
Scope:  auth
Files:  +127 -15
Types:  [go]
```

### Auto-Commit Mode

Skip the interactive prompt and commit automatically:

```bash
gitmit propose --auto
```

## How It Works

### 1. Pattern Detection

Gitmit analyzes your code changes and detects patterns like:

- Error handling improvements
- Test additions/updates
- API endpoint changes
- Database schema modifications
- Security enhancements
- Performance optimizations
- Configuration updates
- Logging enhancements
- Middleware changes
- CLI command additions
- Interface implementations
- Type definitions
- And more...

### 2. Context Extraction

The analyzer extracts context from your changes:

- File types and extensions
- Directory structure
- Function names
- Struct definitions
- Method signatures
- Import statements
- Line additions/deletions
- Change magnitude

### 3. Smart Template Selection

Using a weighted scoring algorithm, Gitmit:

- Matches templates to detected patterns (+2.0 points)
- Rewards relevant placeholders (+1.0-3.0 points)
- Considers file type context (+0.5-1.5 points)
- Applies special case bonuses (+2.5 points)
- Penalizes generic templates (-1.0 points)
- Adds small randomness for variety (+0.0-0.5 points)

### 4. Intelligent Variation

When you press 'r' to regenerate:

1. **Filters used messages** - Never shows the same suggestion twice
2. **Calculates similarity** - Uses word overlap and character matching
3. **Applies diversity bonus** - Rewards dissimilar messages
4. **Maintains relevance** - Still uses context-based scoring
5. **Provides alternatives** - Up to 10 unique regenerations

The similarity algorithm combines:
- Word-level comparison (60% weight) - Jaccard similarity of word sets
- Character-level comparison (40% weight) - Position-based matching

## Tips for Best Results

### 1. Make Focused Commits

Stage related changes together for better context detection:

```bash
# Good - focused change
git add internal/api/auth.go
gitmit

# Better - related files
git add internal/api/auth.go internal/api/middleware.go
gitmit
```

### 2. Use Descriptive File Names

File names help Gitmit understand context:
- `auth.go` → Suggests authentication-related messages
- `handler.go` → Suggests API/handler messages
- `_test.go` → Suggests test-related messages

### 3. Write Clear Code

Gitmit detects:
- Function names (e.g., `func ValidateUser`)
- Struct names (e.g., `type UserAuth struct`)
- Method names (e.g., `func (u *User) Authenticate()`)

Clear names lead to better suggestions.

### 4. Try Regeneration

Don't settle for the first suggestion! Press 'r' a few times to see alternatives. Each regeneration provides a fresh perspective while maintaining context relevance.

### 5. Use Edit When Needed

The edit option (`e`) is perfect for:
- Adding specific ticket numbers
- Including additional context
- Fine-tuning the wording
- Adding scope details

## Troubleshooting

### No Staged Changes

```
⚠️ no staged changes
```

**Solution:** Stage your changes first with `git add`

### Generic Suggestions

If suggestions seem too generic:
1. Try regenerating with 'r'
2. Use `--context` to see what was detected
3. Check if file names are descriptive
4. Ensure changes have clear function/struct names

### Maximum Regenerations Reached

```
⚠ Maximum regeneration attempts reached.
```

**Solution:** You've regenerated 10 times. Either:
- Accept one of the suggestions and edit it (`e`)
- Exit and refine your changes (`n`)
- Use a custom message

## Examples by Change Type

### Adding a New Feature

```
Changes: +50 lines in internal/api/users.go
Detected: New function CreateUser
Pattern: API changes

Suggestion: feat(api): implement CreateUser functionality
```

### Fixing a Bug

```
Changes: ±20 lines in internal/service/payment.go
Detected: Error handling added
Pattern: Error handling

Suggestion: fix(service): resolve error handling in payment processing
```

### Refactoring Code

```
Changes: +100, -95 lines across 3 files
Detected: Multiple functions moved
Pattern: Refactoring

Suggestion: refactor(core): restructure payment service modules
```

### Adding Tests

```
Changes: +150 lines in internal/api/users_test.go
Detected: func TestCreateUser
Pattern: Test addition

Suggestion: test(api): add unit tests for CreateUser
```

### Updating Documentation

```
Changes: +30 lines in README.md
Detected: Markdown file
Pattern: Documentation

Suggestion: docs: update README with authentication guide
```

## Performance

All operations are performed locally with zero network overhead:

- **Analysis:** ~10-50ms depending on change size
- **Template selection:** ~1-5ms
- **Variation generation:** ~2-10ms
- **No API calls or AI inference required**

Perfect for offline development, CI/CD pipelines, and fast workflows.

## Summary

The new interactive mode provides a professional, intelligent commit message experience without requiring AI or external services. Key benefits:

✅ **Smart suggestions** based on code analysis
✅ **Multiple alternatives** with diversity algorithms
✅ **Manual editing** when you need control
✅ **Zero configuration** - works immediately
✅ **Completely offline** - no API keys needed
✅ **Fast execution** - millisecond response times
✅ **Context-aware** - understands your changes

Happy committing! 🎉
