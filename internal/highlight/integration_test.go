package highlight

import (
	"testing"
)

func TestUIIntegrationErrorHandling(t *testing.T) {
	// Test the same pattern used in the UI layer
	h := New(true, "monokai")

	testCases := []struct {
		name     string
		content  string
		filePath string
	}{
		{"Empty file path", "package main", ""},
		{"Unsupported extension", "some content", "file.unknown"},
		{"Binary content", "\x00\x01\x02\x03", "binary.exe"},
		{"Very large line", string(make([]byte, 100000)), "large.go"},
		{"Invalid UTF-8", "package main\x80\x81", "invalid.go"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var processedContent string
			if h.IsEnabled() && h.IsSupported(tc.filePath) {
				processedContent = h.Highlight(tc.content, tc.filePath)
			} else {
				processedContent = tc.content
			}

			// Should never panic and should return some content
			if len(processedContent) == 0 && len(tc.content) > 0 {
				t.Errorf("Expected non-empty result for non-empty input")
			}
		})
	}
}

func TestMultipleHighlighterInstances(t *testing.T) {
	// Test that multiple highlighters don't interfere with each other
	h1 := New(true, "monokai")
	h2 := New(true, "github")
	h3 := New(false, "dracula")

	testCode := `package main
func main() {
	println("test")
}`

	result1 := h1.Highlight(testCode, "test.go")
	result2 := h2.Highlight(testCode, "test.go")
	result3 := h3.Highlight(testCode, "test.go")

	// h3 is disabled, should return original
	if result3 != testCode {
		t.Error("Disabled highlighter should return original content")
	}

	// h1 and h2 should both work but may be different due to different styles
	if len(result1) == 0 || len(result2) == 0 {
		t.Error("Enabled highlighters should return non-empty results")
	}
}
