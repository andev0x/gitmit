# Contributing to Gitmit

Thank you for your interest in contributing to Gitmit! This document provides guidelines and information for contributors.

## 🎯 Project Goals

Gitmit aims to be:
- **Fast and lightweight**: Minimal dependencies, quick analysis
- **Privacy-focused**: Complete offline operation
- **Professional**: Industry-standard commit message format
- **User-friendly**: Simple CLI interface with helpful prompts
- **Reliable**: Consistent and predictable behavior

## 🚀 Getting Started

### Prerequisites

- Go 1.21 or higher
- Git
- Basic understanding of Go and git workflows

### Development Setup

1. **Fork and clone the repository**
   ```bash
   git clone https://github.com/andev0x/gitmit.git
   cd gitmit
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Build the project**
   ```bash
   go build -o gitmit
   ```

4. **Run tests**
   ```bash
   go test ./...
   ```

5. **Test the CLI locally**
   ```bash
   ./gitmit --help
   ```

## 📁 Project Structure

```
gitmit/
├── cmd/                    # CLI commands and application entry
│   └── root.go            # Root command and main application logic
├── internal/              # Internal packages (not importable by external projects)
│   ├── analyzer/          # Git repository analysis
│   │   └── git.go        # Git operations and change analysis
│   ├── generator/         # Commit message generation
│   │   └── message.go    # Message generation logic
│   └── prompt/            # User interaction
│       └── interactive.go # Interactive prompts
├── main.go                # Application entry point
├── go.mod                 # Go module definition
├── go.sum                 # Go module checksums
├── README.md              # Project documentation
└── CONTRIBUTING.md        # This file
```

## 🛠️ Development Guidelines

### Code Style

- Follow standard Go conventions and formatting (`gofmt`, `golint`)
- Use meaningful variable and function names
- Add comments for exported functions and complex logic
- Keep functions focused and single-purpose
- Use early returns to reduce nesting

### Commit Messages

Since this is a tool for commit messages, we should practice what we preach! Use conventional commit format:

```
type(scope): description

Examples:
feat(analyzer): add support for Python file detection
fix(prompt): handle EOF gracefully in interactive mode
docs: update installation instructions
test(generator): add unit tests for scope detection
refactor(cmd): simplify command structure
```

### Testing

- Write unit tests for new functionality
- Ensure existing tests pass before submitting PR
- Test edge cases and error conditions
- Use table-driven tests where appropriate

Example test structure:
```go
func TestGenerateMessage(t *testing.T) {
    tests := []struct {
        name     string
        analysis *analyzer.ChangeAnalysis
        expected string
    }{
        {
            name: "simple feature addition",
            analysis: &analyzer.ChangeAnalysis{
                Added: []string{"src/feature.go"},
                // ... other fields
            },
            expected: "feat: add feature.go",
        },
        // ... more test cases
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            generator := New()
            result := generator.GenerateMessage(tt.analysis)
            if result != tt.expected {
                t.Errorf("expected %q, got %q", tt.expected, result)
            }
        })
    }
}
```

## 🐛 Bug Reports

When reporting bugs, please include:

1. **Environment information**:
   - Operating system and version
   - Go version (`go version`)
   - Gitmit version (`gitmit --version`)

2. **Steps to reproduce**:
   - Clear, step-by-step instructions
   - Sample repository state if possible
   - Expected vs actual behavior

3. **Additional context**:
   - Error messages or logs
   - Screenshots if relevant
   - Any workarounds you've found

## 💡 Feature Requests

For new features, please:

1. **Check existing issues** to avoid duplicates
2. **Describe the use case** and problem you're solving
3. **Propose a solution** if you have ideas
4. **Consider backwards compatibility** and project goals

## 🔄 Pull Request Process

1. **Create a feature branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes**
   - Follow the coding guidelines
   - Add tests for new functionality
   - Update documentation if needed

3. **Test your changes**
   ```bash
   go test ./...
   go build -o gitmit
   ./gitmit --help  # Test basic functionality
   ```

4. **Commit your changes**
   ```bash
   git add .
   gitmit  # Use gitmit to create your commit message!
   ```

5. **Push and create PR**
   ```bash
   git push origin feature/your-feature-name
   ```

6. **PR Requirements**:
   - Clear title and description
   - Reference related issues
   - Include test coverage
   - Ensure CI passes

## 🧪 Testing Guidelines

### Unit Tests
- Test individual functions and methods
- Mock external dependencies (git commands)
- Use table-driven tests for multiple scenarios

### Integration Tests
- Test complete workflows
- Use temporary git repositories
- Verify actual git operations

### Manual Testing
Before submitting, manually test:
- Basic functionality (`gitmit` in a real repository)
- Edge cases (empty repository, no staged changes)
- Different file types and change patterns
- Interactive prompts and user input

## 📚 Documentation

When contributing:
- Update README.md for user-facing changes
- Add inline code comments for complex logic
- Update help text and command descriptions
- Consider adding examples for new features

## 🏷️ Release Process

Releases are managed by maintainers:

1. Version bumping follows semantic versioning
2. Release notes highlight new features and breaking changes
3. Binaries are built for multiple platforms
4. Go module is tagged appropriately

## 🤝 Code of Conduct

- Be respectful and inclusive
- Focus on constructive feedback
- Help newcomers and answer questions
- Maintain a professional tone in all interactions

## 💬 Getting Help

- **Issues**: For bugs and feature requests
- **Discussions**: For questions and general discussion
- **Email**: For security issues or private concerns

## 🎉 Recognition

Contributors will be:
- Listed in release notes for significant contributions
- Mentioned in the README contributors section
- Invited to help with project direction and decisions

---

Thank you for contributing to Gitmit! Every contribution, no matter how small, helps make the tool better for everyone.