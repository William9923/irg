# igrep - Interactive Grep

A terminal UI for interactive grep search with real-time results and live file preview. Inspired by ijq (interactive jq), igrep provides a responsive interface for searching through codebases using ripgrep's powerful search engine.

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

1. Type your search pattern in the pattern input field
2. Results appear automatically after 200ms of inactivity
3. Use arrow keys to navigate through matches
4. View file context in the right preview pane
5. Optionally, specify a path scope in the path input field
6. Press Esc or Ctrl+C to exit

### Search Options

igrep supports three case sensitivity modes:

- **Smart** (default): Case-insensitive search for lowercase patterns, case-sensitive for patterns with uppercase letters
- **Sensitive**: Always case-sensitive search  
- **Insensitive**: Always case-insensitive search

You can set the initial mode with `--case=smart|sensitive|insensitive` or toggle between modes interactively with **Ctrl+T**.

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

## Future Plans

1. Open in $EDITOR - Press Enter to open file at matched line in user's editor => DONE
2. Regex/literal toggle - Switch between regex and fixed-string mode
3. Case sensitivity toggle - Toggle smart-case vs case-sensitive vs case-insensitive => DONE
4. File type filters - Filter by extension using ripgrep's --type flag
5. Match highlighting - Highlight the actual matched text within lines (submatch data is already parsed)
6. Syntax highlighting - Highlight code in preview pane using chroma
7. History persistence - Save search patterns to XDG data directory
8. Add tests - No tests exist yet; could add tests for search package

## Dependencies

- **github.com/charmbracelet/bubbletea** (v1.2.4): Framework for building terminal user interfaces
- **github.com/charmbracelet/bubbles** (v0.20.0): Component library for Bubble Tea (textinput, viewport)
- **github.com/charmbracelet/lipgloss** (v1.0.0): Styling library for terminal UI

## License

MIT
