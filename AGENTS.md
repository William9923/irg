# AGENTS.md - AI Coding Agent Guidelines for irg

## Project Overview

**irg** is an interactive ripgrep TUI tool built with Go and the Charm Bubble Tea framework. It provides real-time search with split-pane results, live preview, and syntax highlighting.

## Quick Reference - Build, Lint, Test

| Command | Description |
|---------|-------------|
| `go build -o irg .` | Build the binary |
| `go run .` | Run without building |
| `go run . --case=smart` | Run with case sensitivity mode |
| `go test ./...` | Run all tests |
| `go test -v ./internal/highlight` | Run tests in specific package |
| `go test -v -run TestHighlighter ./...` | **Run a single test by name** |
| `go test -race ./...` | Run tests with race detector |
| `go test -bench=. ./...` | Run benchmarks |
| `go vet ./...` | Static analysis |
| `goimports -w .` | Format + organize imports (preferred) |
| `golangci-lint run` | Comprehensive linting (if installed) |

**Note**: Tests require ripgrep (`rg`) installed in PATH.

## Project Structure

```
irg/
├── main.go                      # Entry point, CLI flags, ripgrep check
├── internal/
│   ├── search/ripgrep.go        # Ripgrep JSON parsing, streaming results
│   ├── ui/model.go              # Bubble Tea Model/View/Update
│   ├── highlight/               # Syntax highlighting (chroma)
│   └── editor/                  # External editor integration
├── go.mod                       # Module: github.com/William9923/irg
└── .github/workflows/           # CI/CD (Go 1.23, 1.21)
```

## Dependencies

- **Runtime**: ripgrep (`rg`) in PATH (required)
- **Go**: 1.23.4+ (tested on 1.21-1.23)
- **Key Libraries**:
  - `github.com/charmbracelet/bubbletea@v1.2.4` - TUI framework
  - `github.com/charmbracelet/bubbles@v0.20.0` - UI components
  - `github.com/charmbracelet/lipgloss@v1.0.0` - Terminal styling
  - `github.com/alecthomas/chroma/v2@v2.23.1` - Syntax highlighting

## Code Style Guidelines

### Import Organization

**CRITICAL**: Group imports in 3 sections, separated by blank lines. Use `goimports -w .` to auto-format.

```go
import (
    // 1. Standard library (alphabetical)
    "context"
    "fmt"

    // 2. Third-party packages (alphabetical)
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"

    // 3. Internal packages (alphabetical)
    "github.com/William9923/irg/internal/search"
)
```

### Naming Conventions

| Element | Convention | Example |
|---------|------------|---------|
| Packages | lowercase, short | `search`, `ui`, `highlight` |
| Exported types/funcs | PascalCase | `Match`, `Highlighter`, `NewModel()` |
| Unexported types | camelCase | `focusedInput`, `debounceMsg` |
| Constants | camelCase | `debounceDelay = 200 * time.Millisecond` |
| Bubble Tea messages | camelCase + `Msg` | `searchResultMsg`, `previewLoadedMsg` |
| Test functions | `Test<Func>_<Scenario>` | `TestHighlighter_EmptyInput` |

### Error Handling

- **Always** return errors, never panic
- Use defer for cleanup immediately after acquisition
- Wrap errors with context: `fmt.Errorf("parse %s: %w", file, err)`

```go
file, err := os.Open(path)
if err != nil {
    return fmt.Errorf("open %s: %w", path, err)
}
defer file.Close()
```

### Bubble Tea Patterns

**Model struct** - Group fields with comments:

```go
type Model struct {
    // UI components
    patternInput textinput.Model
    resultsView  viewport.Model

    // State
    results []search.Match

    // Search coordination
    searchCtx    context.Context
    searchCancel context.CancelFunc
}
```

**Update method** - Type switch pattern:

```go
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "ctrl+c":
            return m, tea.Quit
        }
    case searchResultMsg:
        m.results = msg.matches
    }
    return m, nil
}
```

**Commands** - Return `tea.Cmd` for async work:

```go
func (m *Model) loadPreviewCmd() tea.Cmd {
    return func() tea.Msg {
        lines, _ := loadFile(m.currentFile)
        return previewLoadedMsg{lines: lines}
    }
}
```

### Context & Cancellation

Always use context for cancellable operations:

```go
ctx, cancel := context.WithCancel(context.Background())
m.searchCancel = cancel  // Store for cancellation

// In goroutine
select {
case <-ctx.Done():
    return
case results <- data:
}
```

### Performance

- Use `strings.Builder` for concatenation in loops
- Use buffered channels: `make(chan Match, 100)`
- Define constants at package level:

```go
const (
    debounceDelay      = 200 * time.Millisecond
    maxResults         = 10000
    maxHighlightLength = 100 * 1024  // 100KB
)
```

## Testing Guidelines

### Structure & Running

```bash
# Test files: *_test.go in same package
go test ./...                           # All tests
go test -v ./internal/highlight         # Specific package
go test -v -run TestHighlighter ./...   # Single test pattern
go test -race ./...                     # Race detector
```

### Table-Driven Test Pattern

```go
func TestDetectLanguage(t *testing.T) {
    tests := []struct {
        name string
        path string
        want string
    }{
        {"go file", "main.go", "go"},
        {"python", "script.py", "python"},
        {"unknown", "file.xyz", ""},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := DetectLanguage(tt.path)
            if got != tt.want {
                t.Errorf("got %q, want %q", got, tt.want)
            }
        })
    }
}
```

## Common Patterns

### 1. Ripgrep JSON Parsing (Two-Phase)

```go
var msg RipgrepMessage
json.Unmarshal(data, &msg)

if msg.Type == "match" {
    var matchData MatchData
    json.Unmarshal(msg.Data, &matchData)  // Deferred parsing
}
```

### 2. Debouncing User Input

```go
m.debounceToken++
token := m.debounceToken

return tea.Tick(200*time.Millisecond, func(t time.Time) tea.Msg {
    return debounceMsg{token: token, pattern: pattern}
})

// In Update - only execute latest
case debounceMsg:
    if msg.token == m.debounceToken {
        return m, m.performSearch()
    }
```

### 3. Streaming Results with Batching

```go
ticker := time.NewTicker(50 * time.Millisecond)
batch := make([]Match, 0, 100)

for {
    select {
    case match := <-results:
        batch = append(batch, match)
        if len(batch) >= 100 {
            sendBatch(batch)
            batch = batch[:0]
        }
    case <-ticker.C:
        if len(batch) > 0 {
            sendBatch(batch)
            batch = batch[:0]
        }
    }
}
```

## Do's and Don'ts

### ✅ Do

- Use `goimports -w .` to auto-organize imports
- Cancel previous search before starting new one
- Handle `tea.WindowSizeMsg` for terminal resize
- Use `strings.Builder` for string concatenation in loops
- Batch UI updates (50ms or 100 items) for smooth rendering
- Close channels in defer after creation
- Test edge cases: empty input, large files, Unicode

### ❌ Don't

- Don't block `Update()` - return commands for async work
- Don't ignore `context.Done()` in goroutines
- Don't use `fmt.Sprintf` in tight loops (performance)
- Don't hardcode terminal dimensions
- Don't panic unless truly unrecoverable
- Don't skip error wrapping (use `%w`)

## CI/CD

GitHub Actions runs on push/PR to main:
- Tests on Go 1.23, 1.21 with ripgrep installed
- Commands: `go test -v ./...`, `go vet ./...`, `go build -v .`
- Release builds for: linux/darwin/windows (amd64/arm64)
