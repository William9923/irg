package highlight

import (
	"testing"
)

func TestErrorHandling(t *testing.T) {
	h := New(true, "monokai")

	// Test with malformed Go code that might cause lexer issues
	malformedCode := `package main
import "fmt
func main() {
	fmt.Println("unclosed quote
	// unclosed brace`

	result := h.Highlight(malformedCode, "test.go")

	// Should not panic and should return some result (even if it's the original)
	if len(result) == 0 {
		t.Error("Expected non-empty result even for malformed code")
	}
}

func TestInvalidStyle(t *testing.T) {
	// Test with non-existent style - should fallback gracefully
	h := New(true, "non-existent-style")

	testCode := `package main
func main() {
	println("test")
}`

	result := h.Highlight(testCode, "test.go")

	// Should still work with fallback style
	if len(result) == 0 {
		t.Error("Expected highlighting to work even with invalid style")
	}
}

func TestNilFormatter(t *testing.T) {
	// Test edge case where formatter might be nil
	h := &Highlighter{
		enabled: true,
		style:   "monokai",
		// formatter and styleObj are nil
	}

	testCode := "package main"
	result := h.Highlight(testCode, "test.go")

	// Should gracefully handle nil formatter and return original content
	if result != testCode {
		t.Error("Expected original content when formatter is nil")
	}
}

func TestVeryLargeContent(t *testing.T) {
	h := New(true, "monokai")

	// Create a large content string
	largeCode := "package main\n"
	for i := 0; i < 1000; i++ {
		largeCode += "// This is a comment line\n"
	}
	largeCode += "func main() { println(\"test\") }"

	result := h.Highlight(largeCode, "test.go")

	// Should handle large content without issues
	if len(result) == 0 {
		t.Error("Expected highlighting to work for large content")
	}
}

func TestSpecialCharacters(t *testing.T) {
	h := New(true, "monokai")

	// Test with special Unicode characters
	codeWithUnicode := `package main
import "fmt"
func main() {
	fmt.Println("Hello 世界! 🌍")
	// Test with emoji: 🚀💻🎉
}`

	result := h.Highlight(codeWithUnicode, "test.go")

	if len(result) == 0 {
		t.Error("Expected highlighting to work with Unicode characters")
	}
}
