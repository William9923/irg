package highlight

import (
	"strings"
	"testing"
	"time"
)

func TestPerformanceOptimizations(t *testing.T) {
	h := New(true, "monokai")

	// Test lexer caching - second call should be faster
	testCode := `package main
import "fmt"
func main() { fmt.Println("test") }`

	start := time.Now()
	result1 := h.Highlight(testCode, "test.go")
	firstDuration := time.Since(start)

	start = time.Now()
	result2 := h.Highlight(testCode, "test.go")
	secondDuration := time.Since(start)

	if result1 != result2 {
		t.Error("Cached result should be identical")
	}

	if secondDuration > firstDuration {
		t.Log("Warning: Second call was slower than first (cache may not be effective)")
	}
}

func TestContentSizeLimits(t *testing.T) {
	h := New(true, "monokai")

	// Test maximum content length limit
	largeContent := strings.Repeat("package main\n", 10000) // > 100KB
	result := h.Highlight(largeContent, "test.go")

	if result != largeContent {
		t.Error("Expected large content to be returned unchanged")
	}

	// Test large line limit
	longLine := strings.Repeat("a", 15*1024) // > 10KB line
	lines := []string{"package main", longLine, "func main() {}"}
	result_lines := h.HighlightLines(lines, "test.go")

	if len(result_lines) != len(lines) {
		t.Errorf("Expected %d lines, got %d", len(lines), len(result_lines))
	}

	// Should return original lines due to long line
	for i, line := range result_lines {
		if line != lines[i] {
			t.Error("Expected lines with long content to be returned unchanged")
			break
		}
	}
}

func TestCacheManagement(t *testing.T) {
	h := New(true, "monokai")

	// Populate cache
	h.Highlight("package main", "test.go")
	h.Highlight("console.log('test')", "test.js")
	h.Highlight("print('test')", "test.py")

	// Verify cache has entries
	h.cacheMutex.RLock()
	cacheSize := len(h.lexerCache)
	h.cacheMutex.RUnlock()

	if cacheSize == 0 {
		t.Error("Expected cache to have entries")
	}

	// Clear cache
	h.ClearCache()

	h.cacheMutex.RLock()
	cacheSize = len(h.lexerCache)
	h.cacheMutex.RUnlock()

	if cacheSize != 0 {
		t.Error("Expected cache to be empty after clearing")
	}

	// Test cache clearing on disable
	h.Highlight("package main", "test.go")
	h.SetEnabled(false)

	h.cacheMutex.RLock()
	cacheSize = len(h.lexerCache)
	h.cacheMutex.RUnlock()

	if cacheSize != 0 {
		t.Error("Expected cache to be cleared when highlighting is disabled")
	}
}

func TestHighlightLinesEfficiency(t *testing.T) {
	h := New(true, "monokai")

	lines := []string{
		"package main",
		"import \"fmt\"",
		"func main() {",
		"    fmt.Println(\"Hello, World!\")",
		"}",
	}

	// Test that HighlightLines works correctly
	start := time.Now()
	result := h.HighlightLines(lines, "test.go")
	duration := time.Since(start)

	if len(result) != len(lines) {
		t.Errorf("Expected %d lines, got %d", len(lines), len(result))
	}

	if duration > 100*time.Millisecond {
		t.Logf("Warning: HighlightLines took %v, may be slower than expected", duration)
	}

	// Verify that each line has some content
	for i, line := range result {
		if len(line) == 0 {
			t.Errorf("Line %d is empty after highlighting", i)
		}
	}
}

func BenchmarkHighlight(b *testing.B) {
	h := New(true, "monokai")
	code := `package main
import (
    "fmt"
    "net/http"
)

func handler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hello, %s!", r.URL.Path[1:])
}

func main() {
    http.HandleFunc("/", handler)
    http.ListenAndServe(":8080", nil)
}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Highlight(code, "test.go")
	}
}

func BenchmarkHighlightCached(b *testing.B) {
	h := New(true, "monokai")
	code := `package main
func main() { println("test") }`

	// Prime the cache
	h.Highlight(code, "test.go")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Highlight(code, "test.go")
	}
}
