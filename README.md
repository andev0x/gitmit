<div align="center">
  <img src="assets/p1.png" alt="Gitmit" width="600"/>

  [![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org)
  [![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
  [![Go Report Card](https://goreportcard.com/badge/github.com/andev0x/gitmit)](https://goreportcard.com/report/github.com/andev0x/gitmit)
</div>

# Gitmit

A lightweight CLI tool that analyzes your staged changes and generates professional git commit messages following the [Conventional Commits](https://www.conventionalcommits.org/) specification.

## Features

- **Intelligent Analysis** - Analyzes git status and diff to understand your changes
- **Conventional Commits** - Follows the Conventional Commits specification for standardized messages
- **Multiple Commit Types** - Supports feat, fix, refactor, chore, test, docs, style, perf, ci, build, security, config, deploy, revert, and wip
- **Interactive Mode** - Customize commit messages with interactive prompts
- **Quick Commits** - Fast commits without interaction
- **Smart Analysis** - Advanced commit history analysis and insights
- **Amend Support** - Easily amend previous commits with smart suggestions
- **Custom Messages** - Use custom messages with scope and breaking change support
- **Zero Configuration** - Works out of the box with sensible defaults
- **Offline First** - Complete offline operation, no external dependencies


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

### Command-Line Options

```bash
# Show suggested message without committing
gitmit --dry-run

# Display detailed analysis of changes
gitmit --verbose

# Commit immediately without interactive prompts
gitmit --quick

# Use OpenAI for enhanced message generation
gitmit --openai

# Amend the previous commit
gitmit --amend

# Force interactive mode
gitmit --interactive

# Use a custom commit message
gitmit --message "your message"

# Specify a commit scope
gitmit --scope "api"

# Mark as breaking change
gitmit --breaking
```

### Subcommands

#### Analyze Commit History

```bash
gitmit analyze
```

Provides insights on:
- Commit patterns and trends
- Most active files and directories
- Commit type distribution
- Development velocity

#### Smart Suggestions

```bash
gitmit smart
```

Offers:
- Multiple commit suggestions with confidence levels
- Context-aware reasoning
- File operation analysis
- Scope detection
- Breaking change identification

#### Propose from Diff

```bash
git diff --cached | gitmit propose
```

Generate a commit message from a diff passed via stdin.

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

### Feature Addition

```bash
git add new-feature.js
gitmit
# Generates: feat: add new-feature.js
```

### Bug Fix

```bash
git add bug-fix.js
gitmit
# Generates: fix: resolve issue in bug-fix.js
```

### Documentation Update

```bash
git add README.md
gitmit
# Generates: docs: update README
```

### Quick Commit

```bash
git add .
gitmit --quick
# Commits immediately without prompts
```

### Custom Message with Scope

```bash
git add .
gitmit --message "improve performance" --scope "api" --breaking
# Generates: feat(api)!: improve performance
```

### Amend Previous Commit

```bash
git add additional-changes.js
gitmit --amend
# Amends the last commit with new changes
```

## Configuration

Gitmit works out of the box without any configuration. No configuration files are required.

### Optional: OpenAI Integration

To use OpenAI for enhanced commit message generation, set your API key:

```bash
export OPENAI_API_KEY="your-api-key"
gitmit --openai
```

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
