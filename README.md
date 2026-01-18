<div align="center">
  <img src="assets/p1.png" alt="Gitmit" width="600"/>

  [![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org)
  [![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
  [![Go Report Card](https://goreportcard.com/badge/github.com/andev0x/gitmit)](https://goreportcard.com/report/github.com/andev0x/gitmit)
</div>

# Gitmit

A lightweight CLI tool that analyzes your staged changes and generates professional git commit messages following the [Conventional Commits](https://www.conventionalcommits.org/) specification.

## Features

- **Intelligent Analysis** - Analyzes git status and diff to understand your changes using advanced pattern detection
- **Conventional Commits** - Follows the Conventional Commits specification for standardized messages
- **Configuration Hierarchy** - Local (`.gitmit.json`) → Global (`~/.gitmit.json`) → Default (Embedded) config support
- **Automatic Project Profiling** - Detects project type (Go, Node.js, Python, Java, etc.) from characteristic files
- **Keyword Scoring Algorithm** - Analyzes git diff content and scores keywords to determine the best commit type
- **Symbol Extraction** - Uses language-aware regex to extract function, class, and variable names
- **Git Porcelain Status** - Leverages `git status --porcelain` for accurate file state detection
- **Diff Stat Analysis** - Infers intent based on added vs deleted lines ratio
- **Commit History Context** - Maintains consistency by learning from recent commit messages
- **Interactive Mode** - Enhanced interactive prompts with y/n/e/r options (yes/no/edit/regenerate)
- **Smart Regeneration** - Generate alternative commit messages with diverse suggestions
- **Context-Aware Scoring** - Weighted algorithm for intelligent template selection
- **Pattern Detection** - Detects error handling, tests, API changes, database operations, and more
- **Multiple Commit Types** - Supports feat, fix, refactor, chore, test, docs, style, perf, ci, build, security, and more
- **Zero Configuration** - Works out of the box with sensible defaults
- **Offline First** - Complete offline operation, no AI or external dependencies required
- **History Tracking** - Learns from your commit history to avoid repetitive suggestions


## Installation

### From Releases

Download the latest release for your platform from the [releases page](https://github.com/andev0x/gitmit/releases).

### Build from Source

Clone the repository and build:

```bash
git clone https://github.com/andev0x/gitmit.git
cd gitmit
go build -o gitmit
```

Or with build confirmation:

```bash
go build -o bin/gitmit 2>&1 && echo "✓ Build successful" && ./bin/gitmit --version
```

### Installation to PATH

#### Linux

```bash
sudo mv gitmit /usr/local/bin
```

#### macOS

First, determine your Go binary directory:

```bash
which go
```

Then move the executable to your Go bin directory:

```bash
sudo mv bin/gitmit $(go env GOPATH)/bin/gitmit
```

Alternatively, add the directory containing `gitmit` to your shell's `PATH` environment variable.


## Quick Start

Stage your changes and generate a commit message:

```bash
git add .
gitmit
```

Gitmit will analyze your changes and suggest a professional commit message following Conventional Commits format.

## Usage

### Interactive Mode (Default)

When you run `gitmit`, it will analyze your changes and present you with an interactive prompt:

```bash
git add .
gitmit

💡 Suggested commit message:
feat(api): implement user authentication strategy

Actions:
  y - Accept and commit
  n - Reject and exit
  e - Edit message manually
  r - Regenerate different suggestion

Choice [y/n/e/r]:
```

**Interactive Options:**
- **`y`** (or press Enter) - Accept the suggestion and commit
- **`n`** - Reject and exit without committing
- **`e`** - Edit the message manually with your own text
- **`r`** - Regenerate a completely different suggestion using intelligent variation algorithms

### Command-Line Options

```bash
# Show suggested message without committing
gitmit --dry-run

# Get multiple ranked suggestions
gitmit --suggestions

# Show analysis context (what was detected)
gitmit --context

# Auto-commit with best suggestion (skip interactive)
gitmit --auto

# Enable debug mode
gitmit --debug
```

### Subcommands

#### Initialize Configuration

```bash
# Create local .gitmit.json in current directory
gitmit init

# Create global ~/.gitmit.json in home directory
gitmit init --global
```

The `init` command automatically detects your project type and generates a configuration file with:
- Language-specific keyword mappings
- Project-appropriate topic mappings
- Customizable keyword scoring weights
- Diff stat analysis thresholds

See [CONFIGURATION.md](CONFIGURATION.md) for detailed configuration options.

#### Propose (Default Command)

```bash
gitmit propose           # Analyze and suggest commit message
gitmit propose -i        # Interactive mode with multiple suggestions
gitmit propose -s        # Show multiple ranked suggestions
gitmit propose --context # Show what was analyzed
gitmit propose --auto    # Auto-commit with best suggestion
```

If no subcommand is provided, `gitmit` defaults to `propose`.

## How It Works

Gitmit uses intelligent offline algorithms to analyze your changes:

1. **Automatic Project Profiling** - Detects project type by checking for:
   - `go.mod` (Go)
   - `package.json` (Node.js)
   - `requirements.txt` (Python)
   - And more (Java, Ruby, Rust, PHP)

2. **Git Porcelain Status** - Uses `git status --porcelain` to read file states:
   - A (Added) → prioritizes `feat` templates
   - M (Modified) → analyzes for `fix`, `refactor`, or `feat`
   - D (Deleted) → suggests `chore` or `refactor`
   - R (Renamed) → suggests `refactor`

3. **Keyword Scoring Algorithm** - Analyzes `git diff --cached` content:
   - Counts keyword occurrences
   - Multiplies by configured weights
   - Selects action with highest score
   - Example: `+ func` (weight: 3) + `+ class` (weight: 2) = 5 points for `feat`

4. **Symbol Extraction via Regex** - Language-aware pattern matching:
   - Go: Functions (`func Name(`), structs (`type Name struct`)
   - JavaScript: Functions, arrow functions, classes
   - Python: Functions (`def name(`), classes (`class Name`)
   - Fills `{item}` placeholder automatically

5. **Path-based Topic Detection** - Uses `filepath.Dir` logic:
   - Custom topic mappings from config
   - Prioritizes `internal/` or `pkg/` subdirectories
   - Falls back to most specific directory name

6. **Diff Stat Analysis** - Analyzes line change ratios:
   - Deleted lines > 70% → suggests `refactor` (cleanup)
   - Added lines > 70% with 50+ lines → suggests `feat` (new feature)
   - Balanced changes → suggests `refactor` (modification)

7. **Commit History Context** - Maintains consistency:
   - Retrieves most recent commit message
   - Extracts scope from `type(scope): message` format
   - Prioritizes same scope for next commit

8. **Pattern Detection** - Identifies code patterns like:
   - Error handling improvements
   - Test additions
   - API/endpoint changes
   - Database operations
   - Security enhancements
   - Performance optimizations
   - Configuration updates
   - And 15+ other patterns

9. **Context Analysis** - Examines:
   - File types and extensions
   - Directory structure
   - Function/struct/method changes
   - Line additions and deletions
   - Multi-file patterns

10. **Weighted Scoring** - Selects templates using:
    - Placeholder availability (item, purpose, topic)
    - Pattern matching bonuses
    - File type context
    - Special case detection
    - Diversity algorithms for variations

## Commit Types

Gitmit supports the following commit types (automatically detected):

| Type | Description |
|------|-------------|
| **feat** | New features |
| **fix** | Bug fixes |
| **refactor** | Code refactoring |
| **chore** | Maintenance tasks |
| **test** | Adding or updating tests |
| **docs** | Documentation changes |
| **style** | Code style changes (formatting, whitespace) |
| **perf** | Performance improvements |
| **ci** | CI/CD configuration changes |
| **build** | Build system changes |
| **security** | Security improvements |
| **config** | Configuration changes |
| **deploy** | Deployment changes |
| **revert** | Reverting previous commits |
| **wip** | Work in progress |

## Examples

### Basic Interactive Usage

```bash
# Stage your changes
git add internal/api/handler.go

# Run gitmit
gitmit

# Output:
💡 Suggested commit message:
feat(api): implement authentication middleware

Actions:
  y - Accept and commit
  n - Reject and exit
  e - Edit message manually
  r - Regenerate different suggestion

Choice [y/n/e/r]: r

# After pressing 'r':
💡 Alternative suggestion #1:
feat(api): add role-based access control for authentication

Choice [y/n/e/r]: y

✅ Changes committed successfully.
```

### Multiple Suggestions Mode

```bash
git add .
gitmit propose -s

# Output:
💡 Ranked Suggestions:
1. feat(api): implement user authentication strategy (recommended)
2. feat(api): add token-based access via middleware
3. feat(auth): integrate OAuth provider for secure access
4. feat(api): expose new endpoint for authentication
5. feat(auth): implement MFA/2FA support for security
```

### Context Analysis

```bash
gitmit propose --context

# Output:
📊 Analysis Context:
Action: feat
Topic:  api
Item:   handler
Purpose: authentication
Scope:  auth
Files:  +127 -15
Types:  [go]

💡 Suggested commit message:
feat(auth): implement handler authentication strategy
```

### Edit Mode

```bash
gitmit

# Output:
💡 Suggested commit message:
feat(api): add new endpoint

Choice [y/n/e/r]: e

📝 Edit the commit message:
Current: feat(api): add new endpoint
New message: feat(api): add user registration endpoint with validation

✓ Updated commit message:
feat(api): add user registration endpoint with validation

Choice [y/n/e/r]: y
✅ Changes committed successfully.
```

## Configuration

Gitmit works out of the box without any configuration, but you can customize its behavior using a configuration file.

### Configuration Hierarchy

1. **Local** (`.gitmit.json`) - Project-specific settings in current directory
2. **Global** (`~/.gitmit.json`) - User-wide settings in home directory
3. **Default** (Embedded) - Built-in defaults

Settings from higher priority configs override lower priority ones.

### Quick Start

```bash
# Create local config with auto-detected project type
gitmit init

# Create global config
gitmit init --global
```

The `init` command automatically:
- Detects your project type (Go, Node.js, Python, etc.)
- Generates language-specific keyword mappings
- Creates customizable topic mappings
- Sets up keyword scoring weights

### Configuration Options

**Core Features:**
- **Project Type Detection** - Automatically identifies language/framework
- **Keyword Scoring** - Define action-specific keywords and weights
- **Topic Mappings** - Map file paths to commit scopes
- **Diff Stat Threshold** - Control added/deleted line ratio analysis
- **Custom Templates** - Define your own commit message patterns (coming soon)

**Example `.gitmit.json`:**
```json
{
  "projectType": "go",
  "diffStatThreshold": 0.5,
  "topicMappings": {
    "internal/api": "api",
    "internal/database": "db"
  },
  "keywords": {
    "feat": {
      "func": 3,
      "class": 2
    },
    "fix": {
      "bug": 3,
      "error": 2
    }
  }
}
```

For detailed configuration documentation, see [CONFIGURATION.md](CONFIGURATION.md).

### Intelligence Built-In

All intelligence is built-in using:

- **Template-based generation** with 100+ curated commit message templates
- **Pattern matching algorithms** for context detection
- **Weighted scoring system** for template selection
- **Similarity detection** for diverse variations
- **Commit history tracking** to avoid repetition
- **Language-aware symbol extraction** via regex
- **Keyword scoring** based on git diff analysis
- **Diff stat analysis** for intent inference

No AI, APIs, or external services required. Everything runs locally and offline.

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines on how to contribute to this project.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Roadmap

- [ ] Git hooks integration
- [ ] Team commit templates
- [ ] Commit message validation
- [ ] Integration with issue trackers
- [ ] Multi-language support
- [ ] Commit message templates
- [ ] Branch-based suggestions
- [ ] Commit message history learning
