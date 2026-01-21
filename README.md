# igrep - Interactive Grep

A terminal UI for interactive grep search with real-time results and live file preview. Inspired by ijq (interactive jq), igrep provides a responsive interface for searching through codebases using ripgrep's powerful search engine.

![igrep demo](./demo.gif)

## Features

- Real-time search results as you type
- 200ms debounce for responsive typing experience
- Split-pane UI: results list (left) and file preview (right)
- Context preview showing 5 lines above and below each match
- Match line highlighting in the preview pane
- Path scoping via dedicated input field
- Smart-case search (case-insensitive unless uppercase is used)
- Streaming results with batching for smooth UI updates
- Results capped at 10,000 for performance
- Built on Charm's Bubble Tea framework for elegant TUI

## Requirements

- **ripgrep (rg)**: Must be installed and available in PATH
- **Go 1.23.4+**: For building from source

## Installation

### Via go install

```bash
go install github.com/william-nobara/igrep@latest
```

Ensure your GOPATH/bin is in your PATH:

```bash
export PATH=$PATH:$(go env GOPATH)/bin
```

### Building from source

```bash
git clone https://github.com/william-nobara/igrep.git
cd igrep
go build -o igrep .
```

Move the binary to a directory in your PATH:

```bash
sudo mv igrep /usr/local/bin/
```

## Usage

Start igrep in your project directory:

```bash
igrep
```

### Keybindings

- **Tab**: Switch between pattern input and path input
- **Up/Down** or **Ctrl+P/Ctrl+N**: Navigate through results
- **PgUp/PgDn**: Jump 10 results at a time
- **Esc** or **Ctrl+C**: Quit

### Basic Workflow

1. Type your search pattern in the pattern input field
2. Results appear automatically after 200ms of inactivity
3. Use arrow keys to navigate through matches
4. View file context in the right preview pane
5. Optionally, specify a path scope in the path input field
6. Press Esc or Ctrl+C to exit

### Search Options

igrep uses ripgrep's `--smart-case` flag by default:
- **Lowercase patterns**: Case-insensitive search
- **Uppercase patterns**: Case-sensitive search

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

## Dependencies

- **github.com/charmbracelet/bubbletea** (v1.2.4): Framework for building terminal user interfaces
- **github.com/charmbracelet/bubbles** (v0.20.0): Component library for Bubble Tea (textinput, viewport)
- **github.com/charmbracelet/lipgloss** (v1.0.0): Styling library for terminal UI

## License

MIT
