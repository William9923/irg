package highlight

import (
	"testing"
)

func TestDetectLanguage(t *testing.T) {
	tests := []struct {
		filePath string
		expected string
	}{
		{"main.go", "go"},
		{"script.py", "python"},
		{"app.js", "javascript"},
		{"style.css", "css"},
		{"config.json", "json"},
		{"README.md", "markdown"},
		{"Dockerfile", "docker"},
		{"Makefile", "makefile"},
		{"unknown.xyz", ""},
		{"", ""},
	}

	for _, test := range tests {
		result := DetectLanguage(test.filePath)
		if result != test.expected {
			t.Errorf("DetectLanguage(%q) = %q; expected %q", test.filePath, result, test.expected)
		}
	}
}

func TestHighlighter(t *testing.T) {
	h := New(true, "monokai")

	if !h.IsEnabled() {
		t.Error("Expected highlighter to be enabled")
	}

	testCode := `package main
import "fmt"
func main() {
	fmt.Println("Hello, World!")
}`

	result := h.Highlight(testCode, "test.go")

	if result == testCode {
		t.Log("Warning: Highlighting may not be working (result unchanged)")
	}

	if len(result) == 0 {
		t.Error("Expected non-empty result from highlighting")
	}
}

func TestHighlighterDisabled(t *testing.T) {
	h := New(false, "monokai")

	if h.IsEnabled() {
		t.Error("Expected highlighter to be disabled")
	}

	testCode := "package main"
	result := h.Highlight(testCode, "test.go")

	if result != testCode {
		t.Error("Expected disabled highlighter to return unchanged content")
	}
}

func TestHighlighterToggle(t *testing.T) {
	h := New(true, "monokai")

	h.SetEnabled(false)
	if h.IsEnabled() {
		t.Error("Expected highlighter to be disabled after SetEnabled(false)")
	}

	h.SetEnabled(true)
	if !h.IsEnabled() {
		t.Error("Expected highlighter to be enabled after SetEnabled(true)")
	}
}

func TestIsSupported(t *testing.T) {
	h := New(true, "monokai")

	supportedFiles := []string{"test.go", "script.py", "app.js"}
	for _, file := range supportedFiles {
		if !h.IsSupported(file) {
			t.Errorf("Expected %q to be supported", file)
		}
	}

	unsupportedFiles := []string{"unknown.xyz", ""}
	for _, file := range unsupportedFiles {
		if h.IsSupported(file) {
			t.Errorf("Expected %q to be unsupported", file)
		}
	}
}

func TestHighlightLines(t *testing.T) {
	h := New(true, "monokai")

	lines := []string{
		"package main",
		"import \"fmt\"",
		"func main() {",
		"  fmt.Println(\"test\")",
		"}",
	}

	result := h.HighlightLines(lines, "test.go")

	if len(result) != len(lines) {
		t.Errorf("Expected %d lines, got %d", len(lines), len(result))
	}

	for i, line := range result {
		if len(line) == 0 {
			t.Errorf("Line %d is empty after highlighting", i)
		}
	}
}

func TestUnsupportedLanguage(t *testing.T) {
	h := New(true, "monokai")

	testCode := "some random content"
	result := h.Highlight(testCode, "unknown.xyz")

	if result != testCode {
		t.Error("Expected unsupported language to return unchanged content")
	}
}

func TestEmptyContent(t *testing.T) {
	h := New(true, "monokai")

	result := h.Highlight("", "test.go")
	if result != "" {
		t.Error("Expected empty content to return empty string")
	}

	lines := h.HighlightLines([]string{}, "test.go")
	if len(lines) != 0 {
		t.Error("Expected empty lines to return empty slice")
	}
}
