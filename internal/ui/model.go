package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/William9923/irg/internal/editor"
	"github.com/William9923/irg/internal/highlight"
	"github.com/William9923/irg/internal/search"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	focusTypes
)

type Model struct {
	patternInput textinput.Model
	pathInput    textinput.Model
	typesInput   textinput.Model
	resultsView  viewport.Model
	previewView  viewport.Model
	focused      focusedInput

	searcher        *search.Searcher
	results         []search.Match
	selectedIndex   int
	searchCtx       context.Context
	searchCancel    context.CancelFunc
	caseSensitivity search.CaseSensitivity

	fileTypes     []string
	fileTypesNot  []string
	lastFileTypes []string

	allTypes          []string // All ripgrep types loaded at startup
	filteredTypes     []string // Currently filtered types for dropdown
	dropdownVisible   bool     // Is dropdown open?
	dropdownIndex     int      // Currently highlighted dropdown item
	dropdownMaxHeight int      // Max items to show (8)

	highlighter *highlight.Highlighter

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

	typesTi := textinput.New()
	typesTi.Placeholder = "Types (e.g., go,rust)"
	typesTi.CharLimit = 256
	typesTi.Width = 30

	resultsVp := viewport.New(40, 20)
	previewVp := viewport.New(40, 20)

	m := Model{
		patternInput:      patternTi,
		pathInput:         pathTi,
		typesInput:        typesTi,
		resultsView:       resultsVp,
		previewView:       previewVp,
		focused:           focusPattern,
		searcher:          search.NewSearcher(),
		results:           make([]search.Match, 0),
		lastPath:          ".",
		caseSensitivity:   search.CaseSmart,
		highlighter:       highlight.New(true, "monokai"),
		width:             80, // Default width for help positioning
		height:            24, // Default height for help positioning
		dropdownMaxHeight: 8,
	}

	m.allTypes, _ = search.LoadRipgrepTypes()
	return m
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

// calculateViewportHeight returns the correct viewport height based on dropdown visibility
// Base height calculation: windowHeight - 7 (for input row + help text + borders)
// When dropdown is visible: subtract additional space for dropdown (11 lines for 8 items + borders)
func (m *Model) calculateViewportHeight() int {
	baseHeight := m.height - 7
	if m.dropdownVisible {
		// Dropdown takes ~11 lines: 8 items + borders + padding + counter
		dropdownHeight := 11
		if len(m.filteredTypes) < m.dropdownMaxHeight {
			// If fewer items, dropdown is smaller: items + 3 for borders/padding
			dropdownHeight = len(m.filteredTypes) + 3
		}
		baseHeight -= dropdownHeight
	}
	if baseHeight < 5 {
		baseHeight = 5 // Minimum viable height
	}
	return baseHeight
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
			wasDropdownVisible := m.dropdownVisible
			if m.focused == focusPattern {
				m.focused = focusPath
				m.patternInput.Blur()
				m.pathInput.Focus()
			} else if m.focused == focusPath {
				m.focused = focusTypes
				m.pathInput.Blur()
				m.typesInput.Focus()
			} else {
				m.focused = focusPattern
				m.typesInput.Blur()
				m.patternInput.Focus()
			}
			m.dropdownVisible = false
			// Update viewport heights when dropdown visibility changes
			if wasDropdownVisible {
				viewportHeight := m.calculateViewportHeight()
				m.resultsView.Height = viewportHeight
				m.previewView.Height = viewportHeight
				m.updateResultsView()
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

		case "ctrl+h":
			m.highlighter.SetEnabled(!m.highlighter.IsEnabled())
			m.updatePreviewView()
			return m, nil

		case "up", "ctrl+p":
			if m.dropdownVisible {
				if m.dropdownIndex > 0 {
					m.dropdownIndex--
				} else {
					m.dropdownIndex = len(m.filteredTypes) - 1
				}
				return m, nil
			}
			if m.selectedIndex > 0 {
				m.selectedIndex--
				m.updateResultsView()
				cmds = append(cmds, m.loadPreview())
			}
			return m, tea.Batch(cmds...)

		case "down", "ctrl+n":
			if m.dropdownVisible {
				if m.dropdownIndex < len(m.filteredTypes)-1 {
					m.dropdownIndex++
				} else {
					m.dropdownIndex = 0
				}
				return m, nil
			}
			if m.selectedIndex < len(m.results)-1 {
				m.selectedIndex++
				m.updateResultsView()
				cmds = append(cmds, m.loadPreview())
			}
			return m, tea.Batch(cmds...)

		case "enter":
			if m.dropdownVisible && len(m.filteredTypes) > 0 {
				selectedType := m.filteredTypes[m.dropdownIndex]
				currentVal := m.typesInput.Value()
				parts := strings.Split(currentVal, ",")
				if len(parts) > 0 {
					parts[len(parts)-1] = selectedType
					newVal := strings.Join(parts, ",")
					m.typesInput.SetValue(newVal)
					m.typesInput.SetCursor(len(newVal))
					m.dropdownVisible = false
					// Update viewport heights when dropdown is closed
					viewportHeight := m.calculateViewportHeight()
					m.resultsView.Height = viewportHeight
					m.previewView.Height = viewportHeight
					m.updateResultsView()
					m.fileTypes = parseTypes(newVal)
					return m, m.executeSearch(m.patternInput.Value(), m.pathInput.Value())
				}
			}
			if m.selectedIndex < len(m.results) && len(m.results) > 0 {
				return m, m.openInEditor()
			}
			return m, nil

		case "esc":
			if m.dropdownVisible {
				m.dropdownVisible = false
				// Update viewport heights when dropdown is closed
				viewportHeight := m.calculateViewportHeight()
				m.resultsView.Height = viewportHeight
				m.previewView.Height = viewportHeight
				m.updateResultsView()
				return m, nil
			}
			if m.focused == focusTypes {
				m.typesInput.SetValue("")
				m.fileTypes = nil
				return m, m.executeSearch(m.patternInput.Value(), m.pathInput.Value())
			}

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

	case tea.MouseMsg:
		// Handle mouse wheel events by updating selectedIndex instead of letting
		// the viewport handle scrolling directly. This ensures scroll position
		// stays synchronized with the selected item through updateResultsView().
		switch msg.Button {
		case tea.MouseButtonWheelUp:
			// Scroll up by 3 lines (default mouse wheel delta)
			m.selectedIndex -= 3
			if m.selectedIndex < 0 {
				m.selectedIndex = 0
			}
			m.updateResultsView()
			cmds = append(cmds, m.loadPreview())
			return m, tea.Batch(cmds...)

		case tea.MouseButtonWheelDown:
			// Scroll down by 3 lines (default mouse wheel delta)
			m.selectedIndex += 3
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
		// For other mouse events (clicks, etc.), let viewport handle them
		var cmd tea.Cmd
		m.resultsView, cmd = m.resultsView.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		patternWidth := (msg.Width - 15) / 2
		pathWidth := (msg.Width - 15) / 4
		typesWidth := (msg.Width - 15) - patternWidth - pathWidth
		m.patternInput.Width = patternWidth
		m.pathInput.Width = pathWidth
		m.typesInput.Width = typesWidth

		listWidth := msg.Width / 3
		previewWidth := msg.Width - listWidth - 5

		viewportHeight := m.calculateViewportHeight()
		m.resultsView.Width = listWidth
		m.resultsView.Height = viewportHeight
		m.previewView.Width = previewWidth
		m.previewView.Height = viewportHeight

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

	var patternCmd, pathCmd, typesCmd tea.Cmd
	m.patternInput, patternCmd = m.patternInput.Update(msg)
	m.pathInput, pathCmd = m.pathInput.Update(msg)
	m.typesInput, typesCmd = m.typesInput.Update(msg)
	cmds = append(cmds, patternCmd, pathCmd, typesCmd)

	currentPattern := m.patternInput.Value()
	currentPath := m.pathInput.Value()
	if currentPath == "" {
		currentPath = "."
	}
	currentTypes := m.typesInput.Value()

	if m.focused == focusTypes && msg != nil {
		if _, ok := msg.(tea.KeyMsg); ok {
			wasDropdownVisible := m.dropdownVisible
			parts := strings.Split(currentTypes, ",")
			lastPart := strings.TrimSpace(parts[len(parts)-1])
			if lastPart != "" {
				m.filteredTypes = nil
				for _, t := range m.allTypes {
					if strings.HasPrefix(t, lastPart) {
						m.filteredTypes = append(m.filteredTypes, t)
					}
				}
				m.dropdownVisible = len(m.filteredTypes) > 0
				if m.dropdownIndex >= len(m.filteredTypes) {
					m.dropdownIndex = 0
				}
			} else {
				m.dropdownVisible = false
			}
			// Update viewport heights when dropdown visibility changes
			if wasDropdownVisible != m.dropdownVisible {
				viewportHeight := m.calculateViewportHeight()
				m.resultsView.Height = viewportHeight
				m.previewView.Height = viewportHeight
				m.updateResultsView()
			}
		}
	}

	newFileTypes := parseTypes(currentTypes)
	typesChanged := false
	if len(newFileTypes) != len(m.lastFileTypes) {
		typesChanged = true
	} else {
		for i := range newFileTypes {
			if newFileTypes[i] != m.lastFileTypes[i] {
				typesChanged = true
				break
			}
		}
	}

	if currentPattern != m.lastPattern || currentPath != m.lastPath || typesChanged {
		m.lastPattern = currentPattern
		m.lastPath = currentPath
		m.lastFileTypes = newFileTypes
		m.fileTypes = newFileTypes
		m.debounceToken++
		token := m.debounceToken

		if m.searchCancel != nil {
			m.searchCancel()
		}

		m.previewPath = ""
		m.previewLines = nil
		m.previewSubmatches = nil
		m.updatePreviewView()

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

		err := m.searcher.Search(m.searchCtx, pattern, path, m.caseSensitivity, m.fileTypes, m.fileTypesNot, results)
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

func (m *Model) SetFileTypes(types, typesNot []string) {
	m.fileTypes = types
	m.fileTypesNot = typesNot
	if len(types) > 0 {
		m.typesInput.SetValue(strings.Join(types, ","))
	}
}

func parseTypes(input string) []string {
	if input == "" {
		return nil
	}
	parts := strings.Split(input, ",")
	var types []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			types = append(types, p)
		}
	}
	return types
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
		centerOffset := targetLine - m.resultsView.Height/2

		// Calculate the maximum valid offset to prevent scrolling past content
		// Content has len(m.results) lines, viewport shows Height lines
		// Maximum offset is when the last line is at the bottom of the viewport
		maxOffset := len(m.results) - m.resultsView.Height

		// Clamp the offset to valid range [0, maxOffset]
		offset := centerOffset
		if offset < 0 {
			offset = 0
		}
		if maxOffset > 0 && offset > maxOffset {
			offset = maxOffset
		}

		m.resultsView.SetYOffset(offset)
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

	matchLineNumStyle := lipgloss.NewStyle().Background(lipgloss.Color("226")).Foreground(lipgloss.Color("0")).Bold(true).Width(4)
	matchTextHighlightStyle := lipgloss.NewStyle().Background(lipgloss.Color("226")).Foreground(lipgloss.Color("196")).Bold(true)

	sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("14")).Bold(true).Render(m.previewPath))
	sb.WriteString("\n")
	sb.WriteString(separatorStyle.Render(strings.Repeat("─", m.previewView.Width-2)))
	sb.WriteString("\n")

	for i, line := range m.previewLines {
		lineNum := m.previewStart + i

		var processedLine string
		if m.highlighter.IsEnabled() && m.highlighter.IsSupported(m.previewPath) {
			processedLine = m.highlighter.Highlight(line, m.previewPath)
		} else {
			processedLine = line
		}

		if lineNum == m.previewMatch {
			styledLineNum := matchLineNumStyle.Render(fmt.Sprintf("%4d", lineNum))

			var highlightedLine string
			if m.highlighter.IsEnabled() && m.highlighter.IsSupported(m.previewPath) {
				// For syntax-highlighted lines, just use a subtle background for the entire line
				// instead of trying to highlight specific matches within colored text
				highlightedLine = lipgloss.NewStyle().Background(lipgloss.Color("236")).Render(processedLine)
			} else {
				// For plain text, use the existing match highlighting
				highlightedLine = highlightMatches(processedLine, m.previewSubmatches, matchTextHighlightStyle)
			}

			sb.WriteString(styledLineNum + " " + highlightedLine)
		} else {
			normalLineNum := normalLineNumStyle.Render(fmt.Sprintf("%4d", lineNum))
			sb.WriteString(normalLineNum + " " + processedLine)
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

func (m *Model) getSyntaxHighlightingStatus() string {
	if m.highlighter.IsEnabled() {
		return "On"
	}
	return "Off"
}

func (m Model) View() string {
	viewportHeight := m.calculateViewportHeight()
	resultsStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Width(m.width / 3).
		Height(viewportHeight)

	previewStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Width(m.width - m.width/3 - 5).
		Height(viewportHeight)

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

	var patternBox, pathBox, typesBox string
	if m.focused == focusPattern {
		patternBox = activeInputStyle.Render(m.patternInput.View())
		pathBox = inactiveInputStyle.Render(m.pathInput.View())
		typesBox = inactiveInputStyle.Render(m.typesInput.View())
	} else if m.focused == focusPath {
		patternBox = inactiveInputStyle.Render(m.patternInput.View())
		pathBox = activeInputStyle.Render(m.pathInput.View())
		typesBox = inactiveInputStyle.Render(m.typesInput.View())
	} else {
		patternBox = inactiveInputStyle.Render(m.patternInput.View())
		pathBox = inactiveInputStyle.Render(m.pathInput.View())
		typesBox = activeInputStyle.Render(m.typesInput.View())
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
		typeInfo := ""
		if len(m.fileTypes) > 0 {
			typeInfo = fmt.Sprintf(" [📁 %s]", strings.Join(m.fileTypes, ","))
		}
		status = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Render(
			fmt.Sprintf("%d matches in %s%s (%s)",
				m.matchCount, pathInfo, typeInfo, m.searchTime.Round(time.Millisecond)))
	} else if m.lastPattern != "" {
		status = "No matches"
	}

	inputRow := lipgloss.JoinHorizontal(lipgloss.Top, patternBox, " ", pathBox, " ", typesBox, "  ", statusStyle.Render(status))

	var dropdown string
	if m.dropdownVisible {
		dropdownStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Background(lipgloss.Color("235")).
			Padding(0, 1).
			Width(m.typesInput.Width + 2)

		var ds strings.Builder
		start := 0
		if m.dropdownIndex >= m.dropdownMaxHeight {
			start = m.dropdownIndex - m.dropdownMaxHeight + 1
		}
		end := start + m.dropdownMaxHeight
		if end > len(m.filteredTypes) {
			end = len(m.filteredTypes)
		}

		for i := start; i < end; i++ {
			t := m.filteredTypes[i]
			selected := i == m.dropdownIndex
			isSelected := false
			for _, ft := range m.fileTypes {
				if ft == t {
					isSelected = true
					break
				}
			}

			prefix := "  "
			if selected {
				prefix = "> "
			}

			suffix := ""
			if isSelected {
				suffix = " [✓]"
			}

			line := prefix + t + suffix
			if selected {
				ds.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12")).Render(line))
			} else {
				ds.WriteString(line)
			}
			ds.WriteString("\n")
		}

		if len(m.filteredTypes) > m.dropdownMaxHeight {
			ds.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(fmt.Sprintf("  [ %d/%d ]", m.dropdownIndex+1, len(m.filteredTypes))))
		}

		dropdown = dropdownStyle.Render(ds.String())
	}

	var helpText string
	if len(m.results) > 0 {
		helpText = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(
			"Keys: ↑/↓ or Ctrl+P/N (navigate) | Enter (open in editor) | Tab (switch input) | Ctrl+T (case: " + m.getCaseSensitivityName() + ") | Ctrl+H (syntax: " + m.getSyntaxHighlightingStatus() + ") | Ctrl+C twice (quit)")
	} else {
		helpText = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(
			"Keys: Tab (switch input) | Ctrl+T (case: " + m.getCaseSensitivityName() + ") | Ctrl+H (syntax: " + m.getSyntaxHighlightingStatus() + ") | Ctrl+C twice (quit)")
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

	if m.dropdownVisible {
		// Append dropdown to help text area if it fits, or just render it below
		// In a real TUI, we'd use relative positioning.
		return lipgloss.JoinVertical(lipgloss.Left, view, dropdown)
	}

	return view
}
