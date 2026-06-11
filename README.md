<h1 align="center">Gitmit</h1>

<p align="center">
  <strong>🧠 Smart, AI-Enhanced Git Commit Message Generator</strong>
</p>

<p align="center">
  <a href="https://golang.org"><img src="https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go" alt="Go Version"></a>
  <a href="https://opensource.org/licenses/MIT"><img src="https://img.shields.io/badge/License-MIT-yellow.svg" alt="License: MIT"></a>
  <a href="https://goreportcard.com/report/github.com/andev0x/gitmit"><img src="https://goreportcard.com/badge/github.com/andev0x/gitmit" alt="Go Report Card"></a>
  <a href="https://github.com/andev0x/gitmit/releases"><img src="https://img.shields.io/github/v/release/andev0x/gitmit" alt="Latest Release"></a>
</p>

---

**Gitmit** is a high-performance CLI tool that takes the guesswork out of commit messages. It analyzes your staged changes and generates professional, context-aware messages following the [Conventional Commits](https://www.conventionalcommits.org/) specification.

Whether you prefer deterministic heuristic rules or powerful local LLMs, Gitmit provides a seamless, offline-first experience to keep your project history clean and meaningful.

## ✨ Why Gitmit?

- **Hybrid Intelligence**: Combines fast, deterministic heuristics with optional Local AI (Ollama) for deep semantic understanding.
- **Privacy First**: Operates 100% locally. No API keys, no data leaving your machine.
- **Project Aware**: Automatically detects your project type (Go, Node.js, Python, Rust, etc.) and tailors suggestions accordingly.
- **Seamless Workflow**: Integrated interactive mode allows you to accept, edit, or regenerate suggestions instantly.

## Screenshots

<p align="left">
  <img src="https://raw.githubusercontent.com/andev0x/description-image-archive/refs/heads/main/gitmit/gitmit.png" width="400" />
</p>

## 🚀 Quick Start

### Installation

#### Using Go:
```bash
go install github.com/andev0x/gitmit@latest
```

#### From Source:
```bash
git clone https://github.com/andev0x/gitmit.git
cd gitmit
make build
sudo make install
```

### Basic Usage

1. **Stage your changes:**
   ```bash
   git add .
   ```

2. **Run Gitmit:**
   ```bash
   gitmit
   ```

3. **Profit!** Accept the suggestion or interactively refine it.

## 🛠 Features

### 🔍 Intelligent Analysis
Uses `git status --porcelain` and `git diff` to understand exactly what changed. It's not just looking at filenames; it understands additions, deletions, and modifications at a structural level.

### 🤖 Hybrid Intelligence (Local AI)
Optionally use models like `qwen2.5-coder` via [Ollama](https://ollama.com) for human-like commit quality. If the AI is unavailable, Gitmit instantly falls back to its robust heuristic engine.

### 📋 Conventional Commit Standards
Supports `feat`, `fix`, `refactor`, `chore`, `test`, `docs`, `style`, `perf`, `ci`, `build`, `security`, and more, ensuring your team stays aligned with industry standards.

### ⚖️ Weighted Scoring Matrix
An advanced algorithm that aggregates signals from:
- **Branch Names**: Extracts intent from `feature/auth` or `fix/bug-123`.
- **Keywords**: Weighted analysis of code changes.
- **Diff Stats**: Understands the ratio of added/deleted lines to distinguish between features and refactors.

### 📦 Language & Ecosystem Awareness
- **Symbol Extraction**: Detects function, class, and variable names in Go, JS/TS, Python, and Java.
- **Dependency Watcher**: Identifies when you add or update libraries in `go.mod`, `package.json`, `requirements.txt`, etc.

## 🧠 Local AI Setup

To leverage the power of LLMs offline:

1. **Install Ollama** from [ollama.com](https://ollama.com).
2. **Pull a model**: `ollama pull qwen2.5-coder:7b`
3. **Initialize Gitmit config**: `gitmit init`
4. **Enable AI** in `.gitmit.json`:
   ```json
   {
     "engine": "ollama",
     "ollama": {
       "model": "qwen2.5-coder:7b"
     }
   }
   ```

## ⌨️ Command Reference

| Command | Description |
|---------|-------------|
| `gitmit` | Analyze changes and suggest a message interactively. |
| `gitmit init` | Create a local `.gitmit.json` configuration. |
| `gitmit init --global` | Create a global `~/.gitmit.json` configuration. |
| `gitmit propose --auto` | Automatically commit with the best suggestion. |
| `gitmit propose -s` | Show multiple ranked suggestions. |
| `gitmit --version` | Show version information. |

### Interactive Actions:
- `y`: **Accept** and commit.
- `n`: **Exit** without committing.
- `e`: **Edit** the message manually.
- `r`: **Regenerate** a new suggestion.
- `a`: **Upgrade** to AI suggestion on-the-fly.

## ⚙️ Configuration

Gitmit works out of the box with zero configuration. For advanced users, it offers deep customization via a tiered config system (Local → Global → Default).

```bash
gitmit init --global
```

Key configuration options include `topicMappings`, `keywordWeights`, and `diffStatThreshold`. For a full deep dive, see [docs/CONFIGURATION.md](docs/CONFIGURATION.md).

## 🤝 Contributing

Contributions are what make the open-source community such an amazing place to learn, inspire, and create. Any contributions you make are **greatly appreciated**.

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Commit your Changes (`gitmit`)
4. Push to the Branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

Check out our [CONTRIBUTING.md](CONTRIBUTING.md) for more details.

## ⚖️ License

Distributed under the MIT License. See `LICENSE` for more information.

---

<p align="center">
  Built with ❤️ by <a href="https://github.com/andev0x">andev0x</a>
</p>
