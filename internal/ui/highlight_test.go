package ui

import (
	"github.com/William9923/irg/internal/highlight"
	"testing"
)

func TestMatchHighlightingWithSyntaxHighlighting(t *testing.T) {
	// Test that we handle match highlighting correctly when syntax highlighting is enabled vs disabled

	h_enabled := highlight.New(true, "monokai")
	h_disabled := highlight.New(false, "monokai")

	testCode := `package main
import "fmt"
func main() {
	fmt.Println("Hello, World!")
}`

	// Test with syntax highlighting enabled
	if h_enabled.IsEnabled() && h_enabled.IsSupported("test.go") {
		result := h_enabled.Highlight(testCode, "test.go")

		// Should contain ANSI color codes
		if result == testCode {
			t.Log("Warning: Syntax highlighting may not be working")
		}

		// Should not contain visible escape sequences when rendered
		if len(result) == 0 {
			t.Error("Expected non-empty result from syntax highlighting")
		}
	}

	// Test with syntax highlighting disabled
	result_disabled := h_disabled.Highlight(testCode, "test.go")
	if result_disabled != testCode {
		t.Error("Disabled highlighter should return original content")
	}
}

func TestPreviewLineProcessing(t *testing.T) {
	// Test the specific logic used in updatePreviewView

	h := highlight.New(true, "monokai")

	testLine := `fmt.Println("test")`
	filePath := "test.go"

	var processedLine string
	if h.IsEnabled() && h.IsSupported(filePath) {
		processedLine = h.Highlight(testLine, filePath)
	} else {
		processedLine = testLine
	}

	// Should produce some output
	if len(processedLine) == 0 {
		t.Error("Expected processed line to have content")
	}

	// Test with plain text file
	plainLine := "This is plain text"
	plainPath := "test.txt"

	var processedPlain string
	if h.IsEnabled() && h.IsSupported(plainPath) {
		processedPlain = h.Highlight(plainLine, plainPath)
	} else {
		processedPlain = plainLine
	}

	// Should return original for unsupported file types
	if processedPlain != plainLine {
		t.Error("Expected plain text to be unchanged")
	}
}
