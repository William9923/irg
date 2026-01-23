# igrep - Interactive Grep

[![Go Report Card](https://goreportcard.com/badge/github.com/william-nobara/igrep)](https://goreportcard.com/report/github.com/william-nobara/igrep)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/github/go-mod/go-version/william-nobara/igrep)](https://golang.org/dl/)
[![Release](https://img.shields.io/github/v/release/william-nobara/igrep?include_prereleases)](https://github.com/william-nobara/igrep/releases)

A terminal UI for interactive grep search with real-time results and live file preview. Inspired by ijq (interactive jq), igrep provides a responsive interface for searching through codebases using ripgrep's powerful search engine.

> **⚠️ Alpha Release**: igrep is currently in alpha (v0.0.x). While functional, expect some rough edges and missing features. Feedback and contributions are welcome!

## 🎬 Demo

```
┌─────────────────────────────────────────────────────────────────────────┐
│ igrep - Interactive Grep                                                │
├─────────────────────────────────────────────────────────────────────────┤
│ Pattern: func.*Error                         │ Path:                    │
│                                             │                          │
├─────────────────────────────────────────────┼──────────────────────────┤
│ [1] main.go:45                              │ 43: // validateInput     │
│     func validateInput() error {            │ 44: // checks if the     │ 
│                                             │ 45: func validateInput() │
│ [2] search/ripgrep.go:123                   │ >>> error {              │
│     func parseError(data []byte) error {    │ 46:   if input == "" {   │
│                                             │ 47:     return fmt.Error │
│ [3] ui/model.go:234                         │                          │
│     func handleKeyError() tea.Model {       │                          │
│                                             │                          │
│ [4] internal/search/ripgrep.go:89           │                          │
│     func streamResultsWithError() error {   │                          │
├─────────────────────────────────────────────┼──────────────────────────┤
│ 🔍 Smart Case • ↑↓ Navigate • Tab Switch • Ctrl+T Toggle • Enter Open  │
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
- **Match highlighting** in the preview pane for easy identification
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

Before installing igrep, ensure you have:

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

- **Go 1.23.4+**: Required only if building from source

### Method 1: Via go install (Recommended)

```bash
# Install latest release
go install github.com/william-nobara/igrep@latest

# Or install specific version
go install github.com/william-nobara/igrep@v0.0.1
```

Ensure your Go bin directory is in your PATH:

```bash
# Add this to your ~/.bashrc, ~/.zshrc, or equivalent
export PATH=$PATH:$(go env GOPATH)/bin

# Or check where Go installs binaries
go env GOPATH
```

### Method 2: Building from source

```bash
git clone https://github.com/william-nobara/igrep.git
cd igrep
go build -o igrep .

# Optional: Move to your PATH
sudo mv igrep /usr/local/bin/
# Or for user-only install:
mv igrep ~/.local/bin/  # Ensure ~/.local/bin is in your PATH
```

### Verify Installation

After installation, verify everything works:

```bash
# Check igrep is accessible
which igrep

# Check ripgrep is accessible
which rg

# Test igrep (should start the TUI)
igrep --help
```

## Usage

Start igrep in your project directory:

```bash
igrep
```

### Command Line Options

- `--case=MODE`: Set case sensitivity mode
  - `smart` (default): Case-insensitive unless uppercase is used
  - `sensitive`: Always case-sensitive
  - `insensitive`: Always case-insensitive

Example:
```bash
igrep --case=sensitive    # Force case-sensitive search
igrep --case=insensitive  # Force case-insensitive search
```

### Keybindings

- **Tab**: Switch between pattern input and path input
- **Up/Down** or **Ctrl+P/Ctrl+N**: Navigate through results
- **PgUp/PgDn**: Jump 10 results at a time
- **Ctrl+T**: Toggle case sensitivity mode (Smart → Sensitive → Insensitive → Smart)
- **Enter**: Open selected result in your default editor
- **Ctrl+C**: Quit (press twice quickly)

### Basic Workflow

**1. Start igrep in your project:**
```bash
cd your-project
igrep
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

igrep supports three case sensitivity modes:

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

igrep is built on a clean separation of concerns:

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

### Key Design Decisions

**Debouncing**: Input is debounced for 200ms to balance responsiveness with performance. This prevents excessive search launches while maintaining an interactive feel.

**Streaming Results**: Results are streamed from ripgrep and batched every 50ms or every 100 matches, whichever comes first. This provides real-time feedback without overwhelming the UI.

**Result Capping**: While ripgrep searches are limited to 1,000 matches per call, the UI caps results at 10,000 to maintain performance on large result sets.

**Context Preview**: The preview pane loads file context asynchronously, showing 5 lines above and below the match to provide useful context without overwhelming the display.

## 🚀 Roadmap

### v0.1.0 - Enhanced Search
- [ ] Regex/literal mode toggle
- [ ] File type filtering (--type flag)
- [ ] Match highlighting within lines
- [ ] Improved error handling

### v0.2.0 - User Experience  
- [ ] Configuration file support
- [ ] Search history persistence
- [ ] Syntax highlighting in preview
- [ ] Custom key bindings

### v1.0.0 - Stable Release
- [ ] Comprehensive test suite
- [ ] Performance optimizations
- [ ] Plugin system
- [ ] Advanced filtering options

See [CHANGELOG.md](CHANGELOG.md) for detailed release notes.

## Dependencies

- **github.com/charmbracelet/bubbletea** (v1.2.4): Framework for building terminal user interfaces
- **github.com/charmbracelet/bubbles** (v0.20.0): Component library for Bubble Tea (textinput, viewport)
- **github.com/charmbracelet/lipgloss** (v1.0.0): Styling library for terminal UI

## Contributing

Contributions are welcome! Feel free to:

- Report bugs or suggest features via [GitHub Issues](https://github.com/william-nobara/igrep/issues)
- Submit pull requests for improvements
- Share feedback and usage examples

### Development

```bash
git clone https://github.com/william-nobara/igrep.git
cd igrep
go mod download
go build -o igrep .
```

See [AGENTS.md](AGENTS.md) for detailed development guidelines.

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Acknowledgments

- Built with [Charm's Bubble Tea](https://github.com/charmbracelet/bubbletea) TUI framework
- Powered by [ripgrep](https://github.com/BurntSushi/ripgrep) for fast text search
- Inspired by [ijq](https://sr.ht/~gpanders/ijq/) for interactive JSON querying
