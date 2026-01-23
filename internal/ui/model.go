package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/william-nobara/igrep/internal/editor"
	"github.com/william-nobara/igrep/internal/search"
)

const (
	debounceDelay  = 200 * time.Millisecond
	maxResults     = 10000
	previewContext = 5
)

type focusedInput int

const (
	focusPattern focusedInput = iota
	focusPath
)

type Model struct {
	patternInput textinput.Model
	pathInput    textinput.Model
	resultsView  viewport.Model
	previewView  viewport.Model
	focused      focusedInput

	searcher        *search.Searcher
	results         []search.Match
	selectedIndex   int
	searchCtx       context.Context
	searchCancel    context.CancelFunc
	caseSensitivity search.CaseSensitivity

	debounceToken int
	lastPattern   string
	lastPath      string

	width  int
	height int

	searching         bool
	matchCount        int
	searchTime        time.Duration
	searchStart       time.Time
	errorMessage      string
	previewPath       string
	previewLines      []string
	previewStart      int
	previewMatch      int
	previewSubmatches []search.Submatch

	ctrlCPressed  bool
	lastCtrlCTime time.Time
}

type searchResultMsg struct {
	matches []search.Match
	done    bool
}

type debounceMsg struct {
	token   int
	pattern string
	path    string
}

type searchErrorMsg struct {
	err error
}

type previewLoadedMsg struct {
	path       string
	lines      []string
	startLine  int
	matchLine  int
	submatches []search.Submatch
}

type editorFinishedMsg struct {
	err error
}

func NewModel() Model {
	patternTi := textinput.New()
	patternTi.Placeholder = "Search pattern..."
	patternTi.Focus()
	patternTi.CharLimit = 256
	patternTi.Width = 40

	pathTi := textinput.New()
	pathTi.Placeholder = "Path (default: .)"
	pathTi.CharLimit = 256
	pathTi.Width = 30

	resultsVp := viewport.New(40, 20)
	previewVp := viewport.New(40, 20)

	return Model{
		patternInput:    patternTi,
		pathInput:       pathTi,
		resultsView:     resultsVp,
		previewView:     previewVp,
		focused:         focusPattern,
		searcher:        search.NewSearcher(),
		results:         make([]search.Match, 0),
		lastPath:        ".",
		caseSensitivity: search.CaseSmart,
		width:           80, // Default width for help positioning
		height:          24, // Default height for help positioning
	}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			now := time.Now()
			if m.ctrlCPressed && now.Sub(m.lastCtrlCTime) < 2*time.Second {
				return m, tea.Quit
			}
			m.ctrlCPressed = true
			m.lastCtrlCTime = now
			return m, nil

		case "tab":
			if m.focused == focusPattern {
				m.focused = focusPath
				m.patternInput.Blur()
				m.pathInput.Focus()
			} else {
				m.focused = focusPattern
				m.pathInput.Blur()
				m.patternInput.Focus()
			}
			return m, nil

		case "ctrl+t":
			switch m.caseSensitivity {
			case search.CaseSmart:
				m.caseSensitivity = search.CaseSensitive
			case search.CaseSensitive:
				m.caseSensitivity = search.CaseInsensitive
			case search.CaseInsensitive:
				m.caseSensitivity = search.CaseSmart
			}
			pattern := m.patternInput.Value()
			path := m.pathInput.Value()
			if pattern != "" {
				cmds = append(cmds, m.executeSearch(pattern, path))
			}
			return m, tea.Batch(cmds...)

		case "up", "ctrl+p":
			if m.selectedIndex > 0 {
				m.selectedIndex--
				m.updateResultsView()
				cmds = append(cmds, m.loadPreview())
			}
			return m, tea.Batch(cmds...)

		case "down", "ctrl+n":
			if m.selectedIndex < len(m.results)-1 {
				m.selectedIndex++
				m.updateResultsView()
				cmds = append(cmds, m.loadPreview())
			}
			return m, tea.Batch(cmds...)

		case "enter":
			if m.selectedIndex < len(m.results) && len(m.results) > 0 {
				return m, m.openInEditor()
			}
			return m, nil

		case "pgup":
			m.selectedIndex -= 10
			if m.selectedIndex < 0 {
				m.selectedIndex = 0
			}
			m.updateResultsView()
			cmds = append(cmds, m.loadPreview())
			return m, tea.Batch(cmds...)

		case "pgdown":
			m.selectedIndex += 10
			if m.selectedIndex >= len(m.results) {
				m.selectedIndex = len(m.results) - 1
			}
			if m.selectedIndex < 0 {
				m.selectedIndex = 0
			}
			m.updateResultsView()
			cmds = append(cmds, m.loadPreview())
			return m, tea.Batch(cmds...)
		}

		// Reset Ctrl+C state on any other key press
		if msg.String() != "ctrl+c" {
			m.ctrlCPressed = false
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		patternWidth := (msg.Width - 10) * 2 / 3
		pathWidth := (msg.Width - 10) - patternWidth
		m.patternInput.Width = patternWidth
		m.pathInput.Width = pathWidth

		listWidth := msg.Width / 3
		previewWidth := msg.Width - listWidth - 5

		m.resultsView.Width = listWidth
		m.resultsView.Height = msg.Height - 7 // Account for input row + help text
		m.previewView.Width = previewWidth
		m.previewView.Height = msg.Height - 7 // Account for input row + help text

		m.updateResultsView()
		m.updatePreviewView()
		return m, nil

	case debounceMsg:
		if msg.token == m.debounceToken {
			return m, m.executeSearch(msg.pattern, msg.path)
		}
		return m, nil

	case searchResultMsg:
		m.results = append(m.results, msg.matches...)
		m.matchCount = len(m.results)

		if msg.done {
			m.searching = false
			m.searchTime = time.Since(m.searchStart)
		}

		if len(m.results) > maxResults {
			m.results = m.results[:maxResults]
		}

		m.updateResultsView()

		if len(m.results) > 0 && m.previewPath == "" {
			cmds = append(cmds, m.loadPreview())
		}

		return m, tea.Batch(cmds...)

	case searchErrorMsg:
		m.errorMessage = msg.err.Error()
		m.searching = false
		return m, nil

	case editorFinishedMsg:
		if msg.err != nil {
			m.errorMessage = fmt.Sprintf("Editor error: %v", msg.err)
		} else {
			m.errorMessage = ""
		}
		return m, nil

	case previewLoadedMsg:
		if m.selectedIndex < len(m.results) && m.results[m.selectedIndex].Path == msg.path {
			m.previewPath = msg.path
			m.previewLines = msg.lines
			m.previewStart = msg.startLine
			m.previewMatch = msg.matchLine
			m.previewSubmatches = msg.submatches
			m.updatePreviewView()
		}
		return m, nil
	}

	var patternCmd, pathCmd tea.Cmd
	m.patternInput, patternCmd = m.patternInput.Update(msg)
	m.pathInput, pathCmd = m.pathInput.Update(msg)
	cmds = append(cmds, patternCmd, pathCmd)

	currentPattern := m.patternInput.Value()
	currentPath := m.pathInput.Value()
	if currentPath == "" {
		currentPath = "."
	}

	if currentPattern != m.lastPattern || currentPath != m.lastPath {
		m.lastPattern = currentPattern
		m.lastPath = currentPath
		m.debounceToken++
		token := m.debounceToken

		if m.searchCancel != nil {
			m.searchCancel()
		}

		m.previewPath = ""
		m.previewLines = nil
		m.previewSubmatches = nil

		cmds = append(cmds, tea.Tick(debounceDelay, func(t time.Time) tea.Msg {
			return debounceMsg{token: token, pattern: currentPattern, path: currentPath}
		}))
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) openInEditor() tea.Cmd {
	if m.selectedIndex >= len(m.results) {
		return nil
	}

	match := m.results[m.selectedIndex]

	ed, err := editor.GetEditor()
	if err != nil {
		return func() tea.Msg {
			return editorFinishedMsg{err: err}
		}
	}

	cmd := ed.BuildCommand(match.Path, match.LineNumber)

	return tea.ExecProcess(cmd, func(err error) tea.Msg {
		return editorFinishedMsg{err: err}
	})
}

func (m *Model) loadPreview() tea.Cmd {
	if m.selectedIndex >= len(m.results) {
		return nil
	}

	match := m.results[m.selectedIndex]

	return func() tea.Msg {
		ctx, err := search.GetFileContextWithMatches(match.Path, match.LineNumber, previewContext, match.Submatches)
		if err != nil {
			return previewLoadedMsg{path: match.Path, lines: []string{"Error loading preview: " + err.Error()}, startLine: 1, matchLine: 1}
		}

		return previewLoadedMsg{
			path:       match.Path,
			lines:      ctx.Lines,
			startLine:  ctx.StartLine,
			matchLine:  ctx.MatchLine,
			submatches: ctx.Submatches,
		}
	}
}

func (m *Model) executeSearch(pattern, path string) tea.Cmd {
	m.results = m.results[:0]
	m.selectedIndex = 0
	m.matchCount = 0
	m.searching = true
	m.errorMessage = ""
	m.searchStart = time.Now()
	m.previewPath = ""
	m.previewLines = nil
	m.previewSubmatches = nil

	m.searchCtx, m.searchCancel = context.WithCancel(context.Background())

	return func() tea.Msg {
		results := make(chan search.Match, 100)

		err := m.searcher.Search(m.searchCtx, pattern, path, m.caseSensitivity, results)
		if err != nil {
			return searchErrorMsg{err: err}
		}

		// Batch results every 50ms to reduce UI redraws while maintaining responsiveness
		var batch []search.Match
		batchTicker := time.NewTicker(50 * time.Millisecond)
		defer batchTicker.Stop()

		for {
			select {
			case match, ok := <-results:
				if !ok {
					return searchResultMsg{matches: batch, done: true}
				}
				batch = append(batch, match)

				if len(batch) >= 100 {
					return searchResultMsg{matches: batch, done: false}
				}

			case <-batchTicker.C:
				if len(batch) > 0 {
					return searchResultMsg{matches: batch, done: false}
				}

			case <-m.searchCtx.Done():
				return searchResultMsg{matches: batch, done: true}
			}
		}
	}
}

// highlightMatches applies highlighting to matched text using submatch positions
func highlightMatches(text string, submatches []search.Submatch, highlightStyle lipgloss.Style) string {
	if len(submatches) == 0 {
		return text
	}

	// Sort submatches by start position to handle overlaps correctly
	sortedMatches := make([]search.Submatch, len(submatches))
	copy(sortedMatches, submatches)

	// Simple bubble sort since we typically have few submatches
	for i := 0; i < len(sortedMatches); i++ {
		for j := i + 1; j < len(sortedMatches); j++ {
			if sortedMatches[i].Start > sortedMatches[j].Start {
				sortedMatches[i], sortedMatches[j] = sortedMatches[j], sortedMatches[i]
			}
		}
	}

	var sb strings.Builder
	lastEnd := 0

	for _, match := range sortedMatches {
		// Handle bounds checking
		start := match.Start
		end := match.End
		if start < 0 || end < 0 || start >= len(text) || end > len(text) || start >= end {
			continue
		}

		// Add text before this match
		if start > lastEnd {
			sb.WriteString(text[lastEnd:start])
		}

		// Add highlighted match text
		matchText := text[start:end]
		sb.WriteString(highlightStyle.Render(matchText))

		lastEnd = end
	}

	// Add remaining text after last match
	if lastEnd < len(text) {
		sb.WriteString(text[lastEnd:])
	}

	return sb.String()
}

func (m *Model) updateResultsView() {
	var sb strings.Builder

	pathStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
	lineNumStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	selectedStyle := lipgloss.NewStyle().Background(lipgloss.Color("237")).Bold(true)
	matchHighlightStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Bold(true)
	selectedMatchHighlightStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Bold(true)

	for i, match := range m.results {
		lineText := strings.TrimRight(match.LineText, "\n\r")
		maxTextLen := m.resultsView.Width - 20
		if maxTextLen > 0 && len(lineText) > maxTextLen {
			lineText = lineText[:maxTextLen-3] + "..."
		}

		var highlightedText string
		if i == m.selectedIndex {
			highlightedText = highlightMatches(lineText, match.Submatches, selectedMatchHighlightStyle)
		} else {
			highlightedText = highlightMatches(lineText, match.Submatches, matchHighlightStyle)
		}

		line := fmt.Sprintf("%s:%s: %s",
			pathStyle.Render(match.Path),
			lineNumStyle.Render(fmt.Sprintf("%d", match.LineNumber)),
			highlightedText)

		if i == m.selectedIndex {
			line = selectedStyle.Render("> " + line)
		} else {
			line = "  " + line
		}

		sb.WriteString(line)
		sb.WriteString("\n")
	}

	m.resultsView.SetContent(sb.String())

	if m.selectedIndex >= 0 && len(m.results) > 0 {
		targetLine := m.selectedIndex
		m.resultsView.SetYOffset(targetLine - m.resultsView.Height/2)
	}
}

func (m *Model) updatePreviewView() {
	if len(m.previewLines) == 0 {
		m.previewView.SetContent("No preview available")
		return
	}

	var sb strings.Builder
	normalLineNumStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Width(4)
	separatorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))

	matchLineStyle := lipgloss.NewStyle().Background(lipgloss.Color("226")).Foreground(lipgloss.Color("0"))
	matchLineNumStyle := lipgloss.NewStyle().Background(lipgloss.Color("226")).Foreground(lipgloss.Color("0")).Bold(true).Width(4)

	sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("14")).Bold(true).Render(m.previewPath))
	sb.WriteString("\n")
	sb.WriteString(separatorStyle.Render(strings.Repeat("─", m.previewView.Width-2)))
	sb.WriteString("\n")

	for i, line := range m.previewLines {
		lineNum := m.previewStart + i

		if lineNum == m.previewMatch {
			styledLineNum := matchLineNumStyle.Render(fmt.Sprintf("%4d", lineNum))
			styledLine := matchLineStyle.Render(line)
			sb.WriteString(styledLineNum + " " + styledLine)
		} else {
			normalLineNum := normalLineNumStyle.Render(fmt.Sprintf("%4d", lineNum))
			sb.WriteString(normalLineNum + " " + line)
		}
		sb.WriteString("\n")
	}

	m.previewView.SetContent(sb.String())
}

func (m *Model) SetCaseSensitivity(caseSensitivity search.CaseSensitivity) {
	m.caseSensitivity = caseSensitivity
}

func (m *Model) getCaseSensitivityName() string {
	switch m.caseSensitivity {
	case search.CaseSmart:
		return "Smart"
	case search.CaseSensitive:
		return "Sensitive"
	case search.CaseInsensitive:
		return "Insensitive"
	default:
		return "Smart"
	}
}

func (m Model) View() string {
	resultsStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Width(m.width / 3).
		Height(m.height - 7) // Account for input row + help text

	previewStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Width(m.width - m.width/3 - 5).
		Height(m.height - 7) // Account for input row + help text

	activeInputStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(0, 1)

	inactiveInputStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1)

	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))

	mainContent := lipgloss.JoinHorizontal(
		lipgloss.Top,
		resultsStyle.Render(m.resultsView.View()),
		previewStyle.Render(m.previewView.View()),
	)

	var patternBox, pathBox string
	if m.focused == focusPattern {
		patternBox = activeInputStyle.Render(m.patternInput.View())
		pathBox = inactiveInputStyle.Render(m.pathInput.View())
	} else {
		patternBox = inactiveInputStyle.Render(m.patternInput.View())
		pathBox = activeInputStyle.Render(m.pathInput.View())
	}

	var status string
	if m.searching {
		status = "Searching..."
	} else if m.errorMessage != "" {
		status = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render(m.errorMessage)
	} else if m.matchCount > 0 {
		pathInfo := m.lastPath
		if pathInfo == "." {
			pathInfo = "current directory"
		}
		status = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Render(
			fmt.Sprintf("%d matches in %s (%s)",
				m.matchCount, pathInfo, m.searchTime.Round(time.Millisecond)))
	} else if m.lastPattern != "" {
		status = "No matches"
	}

	inputRow := lipgloss.JoinHorizontal(lipgloss.Top, patternBox, " ", pathBox, "  ", statusStyle.Render(status))

	var helpText string
	if len(m.results) > 0 {
		helpText = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(
			"Keys: ↑/↓ or Ctrl+P/N (navigate) | Enter (open in editor) | Tab (switch input) | Ctrl+T (case: " + m.getCaseSensitivityName() + ") | Ctrl+C twice (quit)")
	} else {
		helpText = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(
			"Keys: Tab (switch input) | Ctrl+T (case: " + m.getCaseSensitivityName() + ") | Ctrl+C twice (quit)")
	}
	if m.ctrlCPressed && time.Since(m.lastCtrlCTime) < 2*time.Second {
		helpText = lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Render(
			"Press Ctrl+C again to quit")
	}
	var viewComponents []string
	viewComponents = append(viewComponents, mainContent, inputRow)
	if helpText != "" {
		viewComponents = append(viewComponents, helpText)
	}

	view := lipgloss.JoinVertical(lipgloss.Left, viewComponents...)

	return view
}
