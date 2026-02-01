# irg - Interactive Ripgrep

[![Go Report Card](https://goreportcard.com/badge/github.com/William9923/irg)](https://goreportcard.com/report/github.com/William9923/irg)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/github/go-mod/go-version/william-nobara/irg)](https://golang.org/dl/)
[![Release](https://img.shields.io/github/v/release/william-nobara/irg?include_prereleases)](https://github.com/William9923/irg/releases)

A terminal UI for interactive grep search with real-time results and live file preview. Inspired by ijq (interactive jq), irg provides a responsive interface for searching through codebases using ripgrep's powerful search engine.

> **⚠️ Alpha Release**: irg is currently in alpha (v0.0.x). While functional, expect some rough edges and missing features. Feedback and contributions are welcome!

## 🎬 Demo

```
┌─────────────────────────────────────────────────────────────────────────┐
│ irg - Interactive Ripgrep                                               │
├─────────────────────────────────────────────────────────────────────────┤
│ Pattern: func.*Error                         │ Path:                    │
│                                             │                           │
├─────────────────────────────────────────────┼───────────────────────────┤
│ [1] main.go:45                              │ 43: // validateInput      │
│     func validateInput() error {            │ 44: // checks if the      │ 
│                                             │ 45: func validateInput()  │
│ [2] search/ripgrep.go:123                   │ >>> error {               │
│     func parseError(data []byte) error {    │ 46:   if input == "" {    │
│                                             │ 47:     return fmt.Error  │
│ [3] ui/model.go:234                         │                           │
│     func handleKeyError() tea.Model {       │                           │
│                                             │                           │
│ [4] internal/search/ripgrep.go:89           │                           │
│     func streamResultsWithError() error {   │                           │
├─────────────────────────────────────────────┼───────────────────────────┤
│ 🔍 Smart Case • ↑↓ Navigate • Tab Switch • Ctrl+T Toggle • Enter Open   │
└─────────────────────────────────────────────────────────────────────────┘
```

**Key Features Demonstrated:**
- **Left pane**: Real-time search results with file paths and line numbers
- **Right pane**: File context showing 5 lines around the match  
- **Highlighted match**: The actual matching line is visually emphasized
- **Status bar**: Shows current search mode and available keybindings
- **Dual inputs**: Separate fields for pattern and optional path scoping

## ✨ Features

### 🚀 Performance & Responsiveness
- **Real-time search results** as you type with 200ms debounce
- **Streaming results** with smart batching (every 50ms or 100 matches)
- **Performance optimized** with results capped at 10,000 matches
- **Powered by ripgrep** for blazing-fast text search

### 🎨 User Interface  
- **Split-pane design**: Results list (left) + file preview (right)
- **Context preview**: Shows 5 lines above and below each match
- **Syntax highlighting**: Automatic language detection and syntax highlighting in preview pane
- **Match highlighting**: Visual emphasis on matching lines in the preview
- **Dual input fields**: Separate pattern and path scoping
- **Status indicators**: Current search mode and available shortcuts

### 🔧 Smart Search Features
- **Smart-case search**: Case-insensitive for lowercase, sensitive for mixed case
- **Case sensitivity toggle**: Cycle through Smart → Sensitive → Insensitive  
- **Path scoping**: Limit search to specific directories or file patterns
- **Regex support**: Full regex pattern matching via ripgrep

### ⚡ Workflow Integration
- **Editor integration**: Press Enter to open files at the exact match line
- **Keyboard-driven**: Complete functionality without mouse
- **Responsive design**: Adapts to terminal size changes
- **Clean exit**: Graceful shutdown with Ctrl+C

## Requirements

- **ripgrep (rg)**: Must be installed and available in PATH
- **Go 1.23.4+**: For building from source

## Installation

### Prerequisites

Before installing irg, ensure you have:

- **ripgrep (rg)**: Must be installed and available in your PATH
  ```bash
  # Ubuntu/Debian
  sudo apt install ripgrep
  
  # macOS
  brew install ripgrep
  
  # Windows (via Chocolatey)
  choco install ripgrep
  
  # Or download from: https://github.com/BurntSushi/ripgrep/releases
  ```

- **Go 1.23.4+**: Required for building from source or using `go install`

### Method 1: Download precompiled binary (Recommended)

Download the latest release from GitHub:

```bash
# Linux (x86_64)
curl -L https://github.com/William9923/irg/releases/download/v0.0.1/irg-v0.0.1-linux-amd64.tar.gz | tar -xz
sudo mv irg /usr/local/bin/

# macOS (x86_64)
curl -L https://github.com/William9923/irg/releases/download/v0.0.1/irg-v0.0.1-darwin-amd64.tar.gz | tar -xz
sudo mv irg /usr/local/bin/

# macOS (ARM64/M1+)
curl -L https://github.com/William9923/irg/releases/download/v0.0.1/irg-v0.0.1-darwin-arm64.tar.gz | tar -xz
sudo mv irg /usr/local/bin/

# Windows (PowerShell)
Invoke-WebRequest -Uri "https://github.com/William9923/irg/releases/download/v0.0.1/irg-v0.0.1-windows-amd64.zip" -OutFile "irg.zip"
Expand-Archive -Path "irg.zip" -DestinationPath "."
# Move irg.exe to a directory in your PATH
```

Or visit the [releases page](https://github.com/William9923/irg/releases/tag/v0.0.1) to download manually.

### Method 2: Via go install

```bash
# Install latest release
go install github.com/William9923/irg@latest

# Or install specific version
go install github.com/William9923/irg@v0.0.1
```

Ensure your Go bin directory is in your PATH:

```bash
# Add this to your ~/.bashrc, ~/.zshrc, or equivalent
export PATH=$PATH:$(go env GOPATH)/bin

# Or check where Go installs binaries
go env GOPATH
```

### Method 3: Building from source

```bash
git clone https://github.com/William9923/irg.git
cd irg
go build -o irg .

# Optional: Move to your PATH
sudo mv irg /usr/local/bin/
# Or for user-only install:
mv irg ~/.local/bin/  # Ensure ~/.local/bin is in your PATH
```

### Verify Installation

After installation, verify everything works:

```bash
# Check irg is accessible
which irg

# Check ripgrep is accessible
which rg

# Test irg (should start the TUI)
irg --help
```

## Usage

Start irg in your project directory:

```bash
irg
```

### Command Line Options

- `--case=MODE`: Set case sensitivity mode
  - `smart` (default): Case-insensitive unless uppercase is used
  - `sensitive`: Always case-sensitive
  - `insensitive`: Always case-insensitive
- `--type=TYPE`: Include only files of type (e.g., `--type=go`)
- `--type-not=TYPE`: Exclude files of type (e.g., `--type-not=test`)

Example:
```bash
irg --case=sensitive    # Force case-sensitive search
irg --case=insensitive  # Force case-insensitive search
irg --type=go --type=rust "func" # Search only in Go and Rust files
```

### Keybindings

- **Tab**: Cycle between pattern input, path input, and type filter
- **Up/Down** or **Ctrl+P/Ctrl+N**: Navigate through results (or dropdown when visible)
- **Enter**: Open selected result in your default editor (or select from dropdown when visible)
- **PgUp/PgDn**: Jump 10 results at a time
- **Ctrl+T**: Toggle case sensitivity mode (Smart → Sensitive → Insensitive → Smart)
- **Esc**: Close dropdown or clear type input
- **Ctrl+C**: Quit (press twice quickly)

### Basic Workflow

**1. Start irg in your project:**
```bash
cd your-project
irg
```

**2. Type your search pattern:**
```
Pattern: [function.*main_____] ← your cursor here
```
Results appear in real-time as you type!

**3. Navigate through results:**
```
Results:                          Preview:
[1] main.go:12               →    10: package main
    function main() {             11: 
[2] src/app.go:45                 12: function main() {
    function mainLoop() {         13:   fmt.Println("Hello")
                                  14: }
```

**4. Optional path scoping:**
```
Pattern: error                Path: [src/______]
```
Only searches within the `src/` directory.

**5. Open files in your editor:**
Press `Enter` on any result to open the file at that line in your default editor.

### Search Options

irg supports three case sensitivity modes:

- **Smart** (default): Case-insensitive search for lowercase patterns, case-sensitive for patterns with uppercase letters
- **Sensitive**: Always case-sensitive search  
- **Insensitive**: Always case-insensitive search

### Example Use Cases

**🔍 Find function definitions:**
```bash
Pattern: "func.*Search"
Results: All functions containing "Search" in their name
```

**📁 Search in specific directories:**
```bash
Pattern: "TODO"          Path: "src/components/"
Results: All TODO comments in component files
```

**🐛 Debug error handling:**
```bash
Pattern: "error.*return"
Results: All error handling patterns in your codebase
```

**📝 Find configuration:**
```bash
Pattern: "config\."      Path: "*.go"
Results: All config usage in Go files
```

## How It Works

### Architecture

irg is built on a clean separation of concerns:

```
┌─────────────────────────────────────────┐
│          Bubble Tea TUI (ui/)           │
│  ┌─────────────────┐  ┌──────────────┐ │
│  │ Pattern Input   │  │  Path Input  │ │
│  ├─────────────────┤  ├──────────────┤ │
│  │ Results View    │  │ Preview View │ │
│  │ (split pane)    │  │ (context)    │ │
│  └─────────────────┘  └──────────────┘ │
└─────────────┬───────────────────────────┘
              │
              │ debounced (200ms)
              │
              ▼
┌─────────────────────────────────────────┐
│        Ripgrep Backend (search/)        │
│  - JSON output parsing                  │
│  - Streaming results                    │
│  - Smart-case search                    │
│  - 1,000 matches per search             │
└─────────────────────────────────────────┘
```

### Components

- **main.go**: Entry point that validates ripgrep installation and launches the Bubble Tea program
- **internal/search/ripgrep.go**: Wraps ripgrep with JSON output parsing and streaming result delivery
- **internal/ui/model.go**: Bubble Tea Model implementing the TUI with debouncing, dual inputs, and split viewports
- **internal/highlight/**: Syntax highlighting engine with automatic language detection
- **internal/editor/**: External editor integration supporting vim, VS Code, and more

### Key Design Decisions

**Debouncing**: Input is debounced for 200ms to balance responsiveness with performance. This prevents excessive search launches while maintaining an interactive feel.

**Streaming Results**: Results are streamed from ripgrep and batched every 50ms or every 100 matches, whichever comes first. This provides real-time feedback without overwhelming the UI.

**Result Capping**: While ripgrep searches are limited to 1,000 matches per call, the UI caps results at 10,000 to maintain performance on large result sets.

**Context Preview**: The preview pane loads file context asynchronously, showing 5 lines above and below the match to provide useful context without overwhelming the display.

## 🚀 Roadmap

### v0.1.0 - Enhanced Search
- [x] Syntax highlighting in preview (✅ Implemented)
- [x] Editor integration with line number support (✅ Implemented)
- [x] File type filtering (--type flag) (✅ Implemented)
- [ ] Regex/literal mode toggle
- [ ] Improved error handling

### v0.2.0 - User Experience  
- [ ] Configuration file support for themes and settings
- [ ] Search history persistence
- [ ] Custom key bindings
- [ ] Toggle syntax highlighting on/off
- [ ] Multiple theme support

### v1.0.0 - Stable Release
- [x] Basic test suite (✅ Implemented for highlight package)
- [ ] Comprehensive test coverage for all packages
- [ ] Performance optimizations
- [ ] Plugin system
- [ ] Advanced filtering options

See [CHANGELOG.md](CHANGELOG.md) for detailed release notes.

## Dependencies

- **github.com/charmbracelet/bubbletea** (v1.2.4): Framework for building terminal user interfaces
- **github.com/charmbracelet/bubbles** (v0.20.0): Component library for Bubble Tea (textinput, viewport)
- **github.com/charmbracelet/lipgloss** (v1.0.0): Styling library for terminal UI
- **github.com/alecthomas/chroma/v2** (v2.23.1): Syntax highlighting engine

## Contributing

Contributions are welcome! Feel free to:

- Report bugs or suggest features via [GitHub Issues](https://github.com/William9923/irg/issues)
- Submit pull requests for improvements
- Share feedback and usage examples

### Development

```bash
git clone https://github.com/William9923/irg.git
cd irg
go mod download
go build -o irg .
```

See [AGENTS.md](AGENTS.md) for detailed development guidelines.

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Acknowledgments

- Built with [Charm's Bubble Tea](https://github.com/charmbracelet/bubbletea) TUI framework
- Powered by [ripgrep](https://github.com/BurntSushi/ripgrep) for fast text search
- Inspired by [ijq](https://sr.ht/~gpanders/ijq/) for interactive JSON querying
