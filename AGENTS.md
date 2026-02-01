# AGENTS.md - AI Coding Agent Guidelines for irg

## Project Overview

**irg** is an interactive ripgrep TUI tool built with Go and the Charm ecosystem (Bubble Tea, Bubbles, Lipgloss). It wraps ripgrep with a real-time search interface.

## Quick Reference

| Command | Description |
|---------|-------------|
| `go build` | Build the binary |
| `go build -o irg .` | Build with explicit output name |
| `go run .` | Run without building |
| `go test ./...` | Run all tests |
| `go test -v ./internal/search` | Run tests in specific package |
| `go test -v -run TestFunctionName ./...` | Run a single test by name |
| `go test -race ./...` | Run tests with race detector |
| `go vet ./...` | Static analysis |
| `gofmt -w .` | Format all Go files |
| `goimports -w .` | Format + organize imports |
| `golangci-lint run` | Comprehensive linting (if installed) |

## Project Structure

```
irg/
├── main.go                      # Entry point, ripgrep check, tea.Program setup
├── internal/
│   ├── search/
│   │   └── ripgrep.go           # Ripgrep JSON parsing, streaming, context loading
│   └── ui/
│       └── model.go             # Bubble Tea Model/View/Update, TUI logic
├── go.mod                       # Module: github.com/William9923/irg
└── go.sum
```

## Dependencies

- **Runtime**: ripgrep (`rg`) must be in PATH
- **Go**: 1.23.4+
- **Libraries**:
  - `github.com/charmbracelet/bubbletea` - TUI framework
  - `github.com/charmbracelet/bubbles` - UI components (textinput, viewport)
  - `github.com/charmbracelet/lipgloss` - Styling

## Code Style Guidelines

### Import Organization

Group imports in this order, separated by blank lines:

```go
import (
    // 1. Standard library
    "context"
    "fmt"
    "strings"

    // 2. Third-party packages
    "github.com/charmbracelet/bubbles/textinput"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"

    // 3. Internal packages
    "github.com/William9923/irg/internal/search"
)
```

### Naming Conventions

| Element | Convention | Example |
|---------|------------|---------|
| Packages | lowercase, short | `search`, `ui` |
| Exported types | PascalCase | `Match`, `Searcher`, `Model` |
| Unexported types | camelCase | `focusedInput`, `debounceMsg` |
| Constants | camelCase (unexported) | `debounceDelay`, `maxResults` |
| Iota enums | camelCase | `focusPattern`, `focusPath` |
| Methods | PascalCase if exported | `Init()`, `Update()`, `View()` |
| Message types | camelCase + `Msg` suffix | `searchResultMsg`, `debounceMsg` |

### Type Definitions

Define types at package level, grouped by purpose:

```go
// Public types first
type Match struct {
    Path       string
    LineNumber int
    LineText   string
    Submatches []Submatch
}

// Internal message types (for Bubble Tea)
type searchResultMsg struct {
    matches []search.Match
    done    bool
}
```

### Error Handling

- Return errors, don't panic
- Close resources in defer immediately after acquisition
- Check errors before using results

```go
file, err := os.Open(path)
if err != nil {
    return nil, err
}
defer file.Close()
```

### Bubble Tea Patterns

**Model struct**: Group fields by concern:

```go
type Model struct {
    // UI components
    patternInput textinput.Model
    previewView  viewport.Model

    // State
    results       []search.Match
    selectedIndex int

    // Search coordination
    searchCtx    context.Context
    searchCancel context.CancelFunc

    // Dimensions
    width  int
    height int
}
```

**Update method**: Use type switch with explicit key handling:

```go
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "ctrl+c", "esc":
            return m, tea.Quit
        case "up", "ctrl+p":
            // handle
        }
    case tea.WindowSizeMsg:
        // handle resize
    case customMsg:
        // handle custom message
    }
    return m, nil
}
```

**Commands**: Return `tea.Cmd` from funcs, use closures for async:

```go
func (m *Model) loadPreview() tea.Cmd {
    return func() tea.Msg {
        // async work here
        return previewLoadedMsg{...}
    }
}
```

### Styling with Lipgloss

Define styles inline where used, or at method scope for reuse:

```go
func (m *Model) updateResultsView() {
    pathStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
    selectedStyle := lipgloss.NewStyle().Background(lipgloss.Color("237")).Bold(true)
    // use styles...
}
```

### Context and Cancellation

Always use context for cancellable operations:

```go
ctx, cancel := context.WithCancel(context.Background())
// Store cancel func to call later
m.searchCancel = cancel

// In goroutine, check context
select {
case <-ctx.Done():
    return
default:
    // continue work
}
```

### Channel Patterns

Use buffered channels for producer-consumer:

```go
results := make(chan search.Match, 100)

// Producer goroutine
go func() {
    defer close(results)
    for /* items */ {
        select {
        case results <- item:
        case <-ctx.Done():
            return
        }
    }
}()
```

### Constants

Define at package level with descriptive names:

```go
const (
    debounceDelay  = 200 * time.Millisecond
    maxResults     = 10000
    previewContext = 5
)
```

## Testing Guidelines

### File Naming

- Test files: `*_test.go` in same package
- Example: `ripgrep_test.go` for `ripgrep.go`

### Test Function Naming

```go
func TestSearch_EmptyPattern(t *testing.T) { }
func TestSearch_WithPath(t *testing.T) { }
func TestGetFileContext_InvalidPath(t *testing.T) { }
```

### Table-Driven Tests

```go
func TestSomething(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {"empty input", "", "", false},
        {"valid input", "foo", "bar", false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := Something(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("got = %v, want %v", got, tt.want)
            }
        })
    }
}
```

## Common Patterns in This Codebase

### JSON Parsing (ripgrep output)

```go
type RipgrepMessage struct {
    Type string          `json:"type"`
    Data json.RawMessage `json:"data"`  // Defer parsing
}

// Two-phase unmarshal
var msg RipgrepMessage
json.Unmarshal(data, &msg)
if msg.Type == "match" {
    var matchData MatchData
    json.Unmarshal(msg.Data, &matchData)
}
```

### Debouncing Input

```go
m.debounceToken++
token := m.debounceToken

cmd := tea.Tick(debounceDelay, func(t time.Time) tea.Msg {
    return debounceMsg{token: token, pattern: pattern}
})

// In Update, check token matches current
case debounceMsg:
    if msg.token == m.debounceToken {
        // Execute search
    }
```

## Do's and Don'ts

### Do

- Use `strings.Builder` for string concatenation in loops
- Batch UI updates (50ms intervals) for smooth rendering
- Cancel previous operations before starting new ones
- Handle `tea.WindowSizeMsg` to adapt to terminal size

### Don't

- Don't use `fmt.Sprintf` in tight loops (use `strings.Builder`)
- Don't block the Update function - return commands instead
- Don't ignore context cancellation in goroutines
- Don't hardcode terminal dimensions
