# 🧠 Gitmit

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Report Card](https://goreportcard.com/badge/github.com/gitmit/gitmit)](https://goreportcard.com/report/github.com/gitmit/gitmit)

A lightweight CLI tool that analyzes your staged changes and suggests professional commit messages following the Conventional Commits format — without relying on AI.

## 🔍 Why Gitmit?

Ever stared at your terminal, wondering what to write for a commit message?

With **Gitmit**, just focus on the code — and let the tool suggest clean, readable commit messages based on what actually changed.

## ✨ Features

- **🎯 Smart Analysis**: Analyzes `git status` and `git diff` to understand your changes
- **📝 Conventional Commits**: Follows industry-standard commit message format
- **⚡ Lightning Fast**: Local analysis with zero external dependencies
- **🔒 Privacy First**: Complete offline operation, no data leaves your machine
- **🎨 Interactive Mode**: Accept, reject, or customize suggestions
- **🛠️ Zero Configuration**: Works out of the box in any git repository
- **🌍 Cross-Platform**: Runs on Linux, macOS, and Windows

## 🚀 Installation

### Using Go Install (Recommended)

```bash
go install github.com/gitmit/gitmit@latest
```

### From Source

```bash
git clone https://github.com/gitmit/gitmit.git
cd gitmit
go build -o gitmit
sudo mv gitmit /usr/local/bin/
```

### Binary Releases

Download the latest binary from the [releases page](https://github.com/gitmit/gitmit/releases).

## 📖 Usage

### Basic Usage

```bash
# Stage your changes first
git add .

# Run gitmit
gitmit

# Interactive prompts will guide you:
# 🧠 Gitmit - Smart Git Commit
# 
# 💡 Suggested commit message:
#    feat(api): add user authentication endpoint
# 
# Accept this message? (y/n/e to edit):
```

### Command Line Options

```bash
gitmit --help              # Show help message
gitmit --version           # Show version number
gitmit --dry-run           # Show suggestion without committing
gitmit --verbose           # Show detailed analysis
```

### Examples

```bash
# Basic usage
gitmit

# Dry run to see suggestion only
gitmit --dry-run

# Verbose mode with detailed analysis
gitmit --verbose
```

## 🎯 Commit Types

Gitmit automatically suggests appropriate commit types based on your changes:

| Type | Description | Example |
|------|-------------|---------|
| `feat` | New features or functionality | `feat(auth): add OAuth2 integration` |
| `fix` | Bug fixes | `fix(api): resolve null pointer exception` |
| `refactor` | Code refactoring | `refactor(db): optimize query performance` |
| `chore` | Maintenance tasks | `chore(deps): update dependencies` |
| `test` | Adding or updating tests | `test(auth): add unit tests for login` |
| `docs` | Documentation changes | `docs: update installation guide` |
| `style` | Code style changes | `style: fix linting issues` |
| `perf` | Performance improvements | `perf(api): optimize database queries` |
| `ci` | CI/CD changes | `ci: add automated testing workflow` |
| `build` | Build system changes | `build: update webpack configuration` |

## 🧠 Smart Analysis

Gitmit analyzes your changes using multiple signals:

### 📁 File Operations
- **Added files** → Usually `feat`
- **Modified files** → Context-dependent (`feat`, `fix`, `refactor`)
- **Deleted files** → Usually `chore`
- **Renamed files** → Usually `refactor`

### 📋 File Patterns
- Test files (`*_test.go`, `*.test.js`) → `test`
- Documentation (`*.md`, `README.*`) → `docs`
- Package files (`go.mod`, `package.json`) → `chore`
- Config files (`*.config.*`, `Dockerfile`) → `chore`

### 🔍 Diff Content Analysis
- Function additions → "added functions"
- Import changes → "updated imports"
- Logging additions → "added logging"
- Error handling → "error handling"
- Database operations → "database changes"
- API endpoints → "api endpoints"

### 🎯 Scope Detection
Automatically detects scopes from:
- Directory names (`src/api` → `api`)
- File patterns (`test/` → `test`)
- Special files (`package.json` → `deps`)

## 🎨 Example Outputs

```bash
feat(api): add user authentication endpoint
fix(ui): resolve button hover state issue
refactor(auth): optimize login flow
chore(deps): update Go modules
test(api): add integration tests for user service
docs: update contributing guidelines
style(lint): fix formatting issues
perf(db): optimize query performance
ci: add automated deployment workflow
build: update Docker configuration
```

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

### Development Setup

```bash
# Clone the repository
git clone https://github.com/gitmit/gitmit.git
cd gitmit

# Install dependencies
go mod download

# Run tests
go test ./...

# Build the project
go build -o gitmit

# Run locally
./gitmit --help
```

### Project Structure

```
gitmit/
├── cmd/                 # CLI commands and root command
├── internal/
│   ├── analyzer/       # Git analysis logic
│   ├── generator/      # Message generation logic
│   └── prompt/         # Interactive prompts
├── main.go             # Application entry point
├── go.mod              # Go module definition
└── README.md           # This file
```

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Inspired by [Conventional Commits](https://www.conventionalcommits.org/)
- Built with [Cobra CLI](https://github.com/spf13/cobra)
- Colored output by [Fatih Color](https://github.com/fatih/color)

## 📊 Roadmap

- [ ] Configuration file support
- [ ] Custom commit type definitions
- [ ] Integration with popular editors (VS Code, Vim)
- [ ] Commit message templates
- [ ] Multi-language diff analysis improvements
- [ ] Git hooks integration

---

**Made with ❤️ by the open source community**

If you find Gitmit useful, please consider giving it a ⭐ on GitHub!