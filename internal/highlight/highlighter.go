package highlight

import (
	"bytes"
	"strings"
	"sync"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

const (
	maxHighlightLength = 100 * 1024 // 100KB max content to highlight
	maxLineLength      = 10 * 1024  // 10KB max line length
)

// Highlighter provides syntax highlighting functionality
type Highlighter struct {
	enabled   bool
	style     string
	formatter chroma.Formatter
	styleObj  *chroma.Style

	lexerCache map[string]chroma.Lexer
	cacheMutex sync.RWMutex
}

// New creates a new syntax highlighter instance
func New(enabled bool, style string) *Highlighter {
	h := &Highlighter{
		enabled:    enabled,
		style:      style,
		lexerCache: make(map[string]chroma.Lexer),
	}

	if enabled {
		h.initialize()
	}

	return h
}

// initialize sets up the formatter and style for highlighting
func (h *Highlighter) initialize() {
	// Get terminal formatter
	h.formatter = formatters.Get("terminal")
	if h.formatter == nil {
		h.formatter = formatters.Fallback
	}

	// Get style
	h.styleObj = styles.Get(h.style)
	if h.styleObj == nil {
		h.styleObj = styles.Fallback
	}
}

// getLexer retrieves a cached lexer or creates a new one
func (h *Highlighter) getLexer(language, filePath string) chroma.Lexer {
	h.cacheMutex.RLock()
	if lexer, exists := h.lexerCache[language]; exists {
		h.cacheMutex.RUnlock()
		return lexer
	}
	h.cacheMutex.RUnlock()

	lexer := lexers.Get(language)
	if lexer == nil {
		lexer = lexers.Match(filePath)
	}
	if lexer == nil {
		return nil
	}

	lexer = chroma.Coalesce(lexer)

	h.cacheMutex.Lock()
	h.lexerCache[language] = lexer
	h.cacheMutex.Unlock()

	return lexer
}

// Highlight applies syntax highlighting to the given content
// Returns the highlighted content or original content if highlighting fails/disabled
func (h *Highlighter) Highlight(content, filePath string) string {
	if !h.enabled || content == "" || h.formatter == nil || h.styleObj == nil {
		return content
	}

	// Skip highlighting for very large content
	if len(content) > maxHighlightLength {
		return content
	}

	language := DetectLanguage(filePath)
	if language == "" {
		return content
	}

	lexer := h.getLexer(language, filePath)
	if lexer == nil {
		return content
	}

	iterator, err := lexer.Tokenise(nil, content)
	if err != nil {
		return content
	}

	var buf bytes.Buffer
	err = h.formatter.Format(&buf, h.styleObj, iterator)
	if err != nil {
		return content
	}

	return buf.String()
}

// HighlightLines applies syntax highlighting to multiple lines efficiently
func (h *Highlighter) HighlightLines(lines []string, filePath string) []string {
	if !h.enabled || len(lines) == 0 {
		return lines
	}

	// Check if any line is too long
	for _, line := range lines {
		if len(line) > maxLineLength {
			return lines // Return original if any line is too long
		}
	}

	// Join lines to maintain syntax context across line breaks
	content := strings.Join(lines, "\n")

	// Check total length
	if len(content) > maxHighlightLength {
		return lines
	}

	highlighted := h.Highlight(content, filePath)
	if highlighted == content {
		return lines // Highlighting failed, return original
	}

	// Split back into lines
	return strings.Split(highlighted, "\n")
}

// IsEnabled returns whether syntax highlighting is enabled
func (h *Highlighter) IsEnabled() bool {
	return h.enabled
}

// SetEnabled enables or disables syntax highlighting
func (h *Highlighter) SetEnabled(enabled bool) {
	h.enabled = enabled
	if enabled && h.formatter == nil {
		h.initialize()
	}
	if !enabled {
		// Clear cache to save memory when disabled
		h.cacheMutex.Lock()
		h.lexerCache = make(map[string]chroma.Lexer)
		h.cacheMutex.Unlock()
	}
}

// SetStyle changes the highlighting style
func (h *Highlighter) SetStyle(style string) {
	h.style = style
	if h.enabled {
		h.styleObj = styles.Get(style)
		if h.styleObj == nil {
			h.styleObj = styles.Fallback
		}
	}
}

// GetStyle returns the current style name
func (h *Highlighter) GetStyle() string {
	return h.style
}

// IsSupported checks if syntax highlighting is supported for the given file
func (h *Highlighter) IsSupported(filePath string) bool {
	language := DetectLanguage(filePath)
	if language == "" {
		return false
	}

	lexer := lexers.Get(language)
	if lexer == nil {
		lexer = lexers.Match(filePath)
	}

	return lexer != nil
}

// ClearCache clears the lexer cache to free memory
func (h *Highlighter) ClearCache() {
	h.cacheMutex.Lock()
	defer h.cacheMutex.Unlock()
	h.lexerCache = make(map[string]chroma.Lexer)
}
