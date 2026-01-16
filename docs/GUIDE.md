# Gitmit User Guide

## 🚀 Installation

### Using Go Install (Recommended)
```bash
go install github.com/andev0x/gitmit@latest
```

### From Source
```bash
git clone https://github.com/andev0x/gitmit.git
cd gitmit
go build -o gitmit
sudo mv gitmit /usr/local/bin/
```

### Binary Releases
Download the latest binary from the [releases page](https://github.com/andev0x/gitmit/releases).

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

## 🎨 Interactive Interface

### How It Works
When you run `gitmit`, it provides an interactive prompt:

```bash
🧠 Gitmit - Smart Git Commit
💡 Suggested commit message:
   feat(api): add user authentication endpoint
Accept this message? (y/n/e to edit):
```

- **`y` or Enter**: Accept the suggested message.
- **`n`**: Cancel the commit.
- **`e`**: Edit the message interactively.

### Editing Messages
If you choose to edit, you'll see a prompt to enter a custom message:
```bash
Enter your commit message: [default: feat(api): add user authentication endpoint]
```
Type your message and press Enter to confirm.

### Example Workflow
1. Stage your changes:
   ```bash
   git add .
   ```
2. Run `gitmit`:
   ```bash
   gitmit
   ```
3. Follow the interactive prompts to commit your changes.

## 🛠️ Troubleshooting

### Common Issues

#### Missing Dependencies
If you encounter errors related to missing dependencies, ensure you have Go 1.21 or higher installed. Run the following command to install dependencies:
```bash
go mod download
```

#### Incorrect Go Version
Check your Go version using:
```bash
go version
```
If your version is below 1.21, update Go from the [official website](https://golang.org/dl/).

#### Binary Not Found
If `gitmit` is not found after installation, ensure your `$GOPATH/bin` is in your `PATH`. Add the following to your shell configuration file:
```bash
export PATH=$PATH:$(go env GOPATH)/bin
```

### Debugging
Run `gitmit --verbose` to see detailed analysis and debug information.
