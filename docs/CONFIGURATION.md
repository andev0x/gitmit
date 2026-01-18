# Gitmit Configuration Guide

Gitmit supports a flexible configuration system that allows you to customize commit message generation, keyword scoring, and project-specific rules.

## Configuration Hierarchy

Gitmit uses a three-tier configuration hierarchy:

1. **Local** (`.gitmit.json`) - Project-specific settings in the current directory
2. **Global** (`~/.gitmit.json`) - User-wide settings in your home directory  
3. **Default** (Embedded) - Built-in defaults

Settings from higher priority configs (local) override lower priority ones (global, then default).

## Quick Start

### Initialize Configuration

Create a configuration file with language-specific defaults:

```bash
# Create local config in current directory
gitmit init

# Create global config in home directory
gitmit init --global
```

The `init` command automatically detects your project type and generates appropriate keyword mappings.

## Configuration File Structure

```json
{
  "projectType": "go",
  "diffStatThreshold": 0.5,
  "topicMappings": {
    "internal/api": "api",
    "internal/database": "db",
    "internal/auth": "auth"
  },
  "keywordMappings": {
    "authentication": "auth",
    "database": "db"
  },
  "keywords": {
    "feat": {
      "func": 3,
      "class": 2,
      "new": 2
    },
    "fix": {
      "bug": 3,
      "error": 2,
      "if err != nil": 3
    }
  },
  "templates": {}
}
```

## Configuration Options

### Project Type

**`projectType`** (string)

Specifies the programming language/framework for your project. Gitmit uses this to apply language-specific keyword detection and symbol extraction.

**Supported values:**
- `go` - Go projects (detects `go.mod`)
- `nodejs` - Node.js/JavaScript projects (detects `package.json`)
- `python` - Python projects (detects `requirements.txt`, `setup.py`, `pyproject.toml`)
- `java` - Java projects (detects `pom.xml`, `build.gradle`)
- `ruby` - Ruby projects (detects `Gemfile`)
- `rust` - Rust projects (detects `Cargo.toml`)
- `php` - PHP projects (detects `composer.json`)
- `generic` - Default for unrecognized projects

**Auto-detection:** If not specified, Gitmit automatically detects the project type by checking for characteristic files.

**Example:**
```json
{
  "projectType": "nodejs"
}
```

### Diff Stat Threshold

**`diffStatThreshold`** (float, default: 0.5)

Controls the threshold for the diff stat analysis algorithm. This ratio determines when to prioritize different commit types based on added vs deleted lines.

**How it works:**
- If `deletedRatio > threshold + 0.2`, suggests `refactor` (cleanup)
- If `addedRatio > threshold + 0.2` with 50+ lines added, suggests `feat` (new feature)
- Balanced changes (both ratios > 0.3) suggest `refactor` (modification)

**Example:**
```json
{
  "diffStatThreshold": 0.6
}
```

### Topic Mappings

**`topicMappings`** (object)

Maps file path patterns to topic/scope names. When files in these paths are changed, Gitmit uses the mapped topic in commit messages.

**Example:**
```json
{
  "topicMappings": {
    "internal/api": "api",
    "internal/database": "db",
    "internal/auth": "auth",
    "cmd": "cli",
    "pkg": "core",
    "docs": "docs"
  }
}
```

**Result:** Changes in `internal/api/handler.go` → `feat(api): ...`

### Keyword Mappings

**`keywordMappings`** (object)

Provides aliases for keywords found in diffs. Simplifies complex terms into shorter scope names.

**Example:**
```json
{
  "keywordMappings": {
    "authentication": "auth",
    "authorization": "auth",
    "database": "db",
    "configuration": "config"
  }
}
```

### Keyword Scoring

**`keywords`** (object)

Defines action-specific keywords and their weights for the keyword scoring algorithm. Gitmit analyzes `git diff --cached` content, counts keyword occurrences, and calculates scores to determine the most appropriate commit type.

**Format:**
```json
{
  "keywords": {
    "<action>": {
      "<keyword>": <weight>
    }
  }
}
```

**Algorithm:**
For each action, score = Σ(keyword_occurrences × weight). The action with the highest score is selected.

**Example:**
```json
{
  "keywords": {
    "feat": {
      "func": 3,
      "class": 2,
      "new": 2,
      "add": 2,
      "implement": 2
    },
    "fix": {
      "bug": 3,
      "fix": 3,
      "error": 2,
      "if err != nil": 3,
      "try": 1,
      "catch": 1
    },
    "refactor": {
      "refactor": 3,
      "restructure": 2,
      "rename": 2
    },
    "test": {
      "test": 3,
      "Test": 3,
      "assert": 2,
      "expect": 2
    },
    "docs": {
      "docs": 3,
      "documentation": 3,
      "comment": 2
    }
  }
}
```

**Language-specific keywords** are automatically added based on `projectType`:

**Go:**
```json
{
  "feat": {
    "func": 3,
    "type": 2,
    "struct": 2,
    "interface": 2
  },
  "fix": {
    "if err != nil": 3,
    "panic": 2
  }
}
```

**Node.js:**
```json
{
  "feat": {
    "function": 3,
    "class": 2,
    "export": 2
  },
  "fix": {
    "try": 2,
    "catch": 2,
    "throw": 2
  }
}
```

**Python:**
```json
{
  "feat": {
    "def": 3,
    "class": 2,
    "async def": 3
  },
  "fix": {
    "except": 2,
    "raise": 2
  }
}
```

### Custom Templates

**`templates`** (object)

Reserved for future use. Will allow custom commit message templates.

## Advanced Features

### Automatic Project Profiling

Gitmit automatically detects your project type by checking for characteristic files:

- `go.mod` → Go
- `package.json` → Node.js
- `requirements.txt`, `setup.py`, `pyproject.toml` → Python
- `pom.xml`, `build.gradle` → Java
- `Gemfile` → Ruby
- `Cargo.toml` → Rust
- `composer.json` → PHP

This enables language-specific keyword sets and symbol extraction without manual configuration.

### Symbol Extraction via Regex

Gitmit uses language-aware regex patterns to extract function, class, and method names from diffs:

**Go:**
- Functions: `func FunctionName(`
- Methods: `func (r Receiver) MethodName(`
- Structs: `type StructName struct`

**JavaScript/TypeScript:**
- Functions: `function functionName(`, arrow functions
- Classes: `class ClassName`, `export class ClassName`

**Python:**
- Functions: `def function_name(`, `async def function_name(`
- Classes: `class ClassName`

**Java/C/C++:**
- Methods: `public/private/protected Type methodName(`
- Classes: `public/private/protected class ClassName`

Extracted symbols are used as `{item}` placeholders in commit messages.

### Path-based Topic Detection

Gitmit uses `filepath.Dir` logic to determine topics from directory structure:

1. Checks custom `topicMappings` first
2. Prioritizes `internal/` or `pkg/` subdirectories
3. Falls back to the most specific non-generic directory name
4. Uses "core" if no specific topic found

**Example:**
- `internal/auth/handler.go` → topic: `auth`
- `pkg/database/queries.go` → topic: `database`
- `cmd/server/main.go` → topic: `server`

### Git Porcelain Status

Gitmit uses `git status --porcelain` for accurate file state detection:

- **A** (Added) → Immediately prioritizes `feat` suggestions
- **M** (Modified) → Analyzes diff for `fix`, `refactor`, or `feat`
- **D** (Deleted) → Suggests `chore` or `refactor`
- **R** (Renamed) → Suggests `refactor`

This eliminates irrelevant suggestions by narrowing to the correct action group.

### Diff Stat Analysis

Analyzes the ratio of added vs deleted lines to infer intent:

```
deletedRatio = totalRemoved / (totalAdded + totalRemoved)
addedRatio = totalAdded / (totalAdded + totalRemoved)
```

**Logic:**
- `deletedRatio > 0.7` → Suggests `refactor` (cleanup/removal)
- `addedRatio > 0.7` with 50+ new lines → Suggests `feat` (new feature)
- Balanced changes → Suggests `refactor` (modification)

### Commit History Context

Retrieves the most recent commit message to maintain consistency:

```bash
git log -1 --pretty=%B
```

Extracts the scope from conventional commit format (`type(scope): message`) and prioritizes the same scope for the next commit.

**Example:**
- Previous commit: `feat(auth): implement OAuth provider`
- Next commit suggestion prioritizes: `feat(auth): ...`

## Examples

### Go Project Configuration

```json
{
  "projectType": "go",
  "diffStatThreshold": 0.5,
  "topicMappings": {
    "internal/api": "api",
    "internal/database": "db",
    "internal/auth": "auth",
    "cmd": "cli",
    "pkg": "core"
  },
  "keywordMappings": {
    "authentication": "auth",
    "database": "db"
  },
  "keywords": {
    "feat": {
      "func": 3,
      "type": 2,
      "struct": 2,
      "interface": 2
    },
    "fix": {
      "if err != nil": 3,
      "error": 2,
      "bug": 3
    },
    "test": {
      "Test": 3,
      "testing.T": 2
    }
  }
}
```

### Node.js/React Project Configuration

```json
{
  "projectType": "nodejs",
  "diffStatThreshold": 0.5,
  "topicMappings": {
    "src/components": "ui",
    "src/api": "api",
    "src/utils": "utils",
    "src/hooks": "hooks"
  },
  "keywordMappings": {
    "component": "ui",
    "endpoint": "api"
  },
  "keywords": {
    "feat": {
      "function": 3,
      "class": 2,
      "const": 1,
      "export": 2,
      "component": 3
    },
    "fix": {
      "bug": 3,
      "fix": 3,
      "try": 2,
      "catch": 2
    }
  }
}
```

### Python Project Configuration

```json
{
  "projectType": "python",
  "diffStatThreshold": 0.5,
  "topicMappings": {
    "src/api": "api",
    "src/models": "models",
    "src/utils": "utils"
  },
  "keywords": {
    "feat": {
      "def": 3,
      "class": 2,
      "async def": 3
    },
    "fix": {
      "except": 2,
      "raise": 2,
      "bug": 3
    },
    "test": {
      "test_": 3,
      "assert": 2
    }
  }
}
```

## Best Practices

1. **Start with defaults**: Run `gitmit init` to generate language-specific defaults
2. **Customize gradually**: Add custom mappings as you identify patterns in your workflow
3. **Team consistency**: Share global config (`~/.gitmit.json`) across team members
4. **Project specificity**: Use local config (`.gitmit.json`) for project-specific rules
5. **Commit the config**: Include `.gitmit.json` in version control for team collaboration

## Troubleshooting

### Config not being loaded
- Check file location (`.gitmit.json` in project root or `~/.gitmit.json` in home)
- Verify JSON syntax with `cat .gitmit.json | jq`

### Wrong project type detected
- Explicitly set `projectType` in config
- Check for conflicting marker files (e.g., both `go.mod` and `package.json`)

### Keywords not affecting suggestions
- Increase keyword weights
- Check that keywords match actual diff content (case-sensitive)
- Use `gitmit propose --debug` to see keyword scores

## Support

For issues, feature requests, or questions about configuration:
- GitHub Issues: https://github.com/andev0x/gitmit/issues
- Documentation: https://github.com/andev0x/gitmit
