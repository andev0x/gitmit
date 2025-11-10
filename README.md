<div align="center">
  <img src="assets/p1.png" alt="Gitmit" width="600"/>


[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Report Card](https://goreportcard.com/badge/github.com/andev0x/gitmit)](https://goreportcard.com/report/github.com/andev0x/gitmit)

</div>

# Gitmit - Smart Git Commit Message Generator

🧠 A lightweight CLI tool that analyzes your staged changes and suggests professional commit messages following Conventional Commits format.

## Features

- **Intelligent Analysis**: Analyzes git status and diff to understand your changes
- **Conventional Commits**: Follows the [Conventional Commits](https://www.conventionalcommits.org/) specification
- **Multiple Commit Types**: Supports feat, fix, refactor, chore, test, docs, style, perf, ci, build, security, config, deploy, revert, and wip
- **Interactive Mode**: Customize commit messages with an interactive prompt
- **Quick Commits**: Fast commits without interaction
- **Smart Analysis**: Advanced commit history analysis and insights
- **Amend Support**: Easily amend previous commits with smart suggestions
- **Custom Messages**: Use custom messages with scope and breaking change support
- **Zero Configuration**: Works out of the box
- **Lightning Fast**: Complete offline operation


## Installation

Gitmit is designed to run everywhere.

### From Releases

Download the latest release for your platform from the [releases page](https://github.com/andev0x/gitmit/releases).

### From Source

```bash
git clone https://github.com/andev0x/gitmit.git
cd gitmit
go build -o gitmit
```

After building, you can install `gitmit` to your system's PATH.

#### Linux (Arch Linux Example)

To install `gitmit` to `/usr/local/bin`:

```bash
sudo mv gitmit /usr/local/bin
```

#### macOS

To determine where `go` binaries are typically installed on your system, use `which go`. This will help you decide where to move the `gitmit` executable. For example, if `which go` returns `/usr/local/bin/go`, you might move `gitmit` there:

```bash
sudo mv gitmit /Users/username/go/bin
```

Alternatively, you can add the directory containing the `gitmit` executable to your shell's PATH environment variable.


## Usage

### Basic Usage

```bash
# Stage your changes
git add .

# Generate and commit with smart message
gitmit
```

### Command Options

```bash
# Show suggested message without committing
gitmit --dry-run

# Show detailed analysis
gitmit --verbose

# Quick commit without interaction
gitmit --quick

# Use OpenAI for enhanced message generation
gitmit --openai

# Amend the last commit
gitmit --amend

# Force interactive mode
gitmit --interactive

# Custom commit message
gitmit --message "your custom message"

# Specify commit scope
gitmit --scope "api"

# Mark as breaking change
gitmit --breaking
```

### Smart Analysis

```bash
# Analyze commit history and get insights
gitmit analyze

# Get smart commit suggestions
gitmit smart
```

### Propose Mode

```bash
# Propose commit message from diff
git diff --cached | gitmit propose
```

## Commit Types

Gitmit automatically detects and suggests appropriate commit types:

- **feat**: New features
- **fix**: Bug fixes
- **refactor**: Code refactoring
- **chore**: Maintenance tasks
- **test**: Adding or updating tests
- **docs**: Documentation changes
- **style**: Code style changes (formatting, etc.)
- **perf**: Performance improvements
- **ci**: CI/CD changes
- **build**: Build system changes
- **security**: Security improvements
- **config**: Configuration changes
- **deploy**: Deployment changes
- **revert**: Reverting previous commits
- **wip**: Work in progress

## Examples

### Feature Addition
```bash
git add new-feature.js
gitmit
# Suggests: feat: add new-feature.js
```

### Bug Fix
```bash
git add bug-fix.js
gitmit
# Suggests: fix: resolve issue in bug-fix.js
```

### Documentation Update
```bash
git add README.md
gitmit
# Suggests: docs: update README
```

### Quick Commit
```bash
git add .
gitmit --quick
# Commits immediately with auto-generated message
```

### Custom Message with Scope
```bash
git add .
gitmit --message "improve performance" --scope "api" --breaking
# Creates: feat(api)!: improve performance
```

### Amend Previous Commit
```bash
git add additional-changes.js
gitmit --amend
# Amends the last commit with new changes
```

## Smart Analysis Features

### Commit History Analysis
```bash
gitmit analyze
```

Provides insights on:
- Commit patterns and trends
- Most active files and directories
- Commit type distribution
- Development velocity
- Potential improvements

### Smart Suggestions
```bash
gitmit smart
```

Offers:
- Multiple commit suggestions with confidence levels
- Context-aware reasoning
- File operation analysis
- Scope detection
- Breaking change identification

## Configuration

Gitmit works out of the box with zero configuration. However, you can enhance it with:

### OpenAI Integration (pending)
Set your OpenAI API key for enhanced commit message generation:
```bash
export OPENAI_API_KEY="your-api-key"
gitmit --openai
```

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for details.

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
