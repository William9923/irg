package highlight

import (
	"testing"
)

func TestStyleChanges(t *testing.T) {
	h := New(true, "monokai")

	if h.GetStyle() != "monokai" {
		t.Errorf("Expected initial style 'monokai', got %q", h.GetStyle())
	}

	h.SetStyle("github")
	if h.GetStyle() != "github" {
		t.Errorf("Expected style 'github', got %q", h.GetStyle())
	}

	testCode := `function test() { return "hello"; }`
	result1 := h.Highlight(testCode, "test.js")

	h.SetStyle("dracula")
	result2 := h.Highlight(testCode, "test.js")

	if len(result1) == 0 || len(result2) == 0 {
		t.Error("Expected highlighting results to be non-empty")
	}
}

func TestLanguageVariants(t *testing.T) {
	tests := map[string][]string{
		"javascript": {"test.js", "app.jsx"},
		"typescript": {"main.ts", "component.tsx"},
		"python":     {"script.py", "module.pyx", "types.pyi"},
		"cpp":        {"main.cpp", "header.hpp", "source.cxx"},
	}

	for expectedLang, filePaths := range tests {
		for _, filePath := range filePaths {
			detected := DetectLanguage(filePath)
			if detected != expectedLang {
				t.Errorf("For %q, expected language %q, got %q", filePath, expectedLang, detected)
			}
		}
	}
}
