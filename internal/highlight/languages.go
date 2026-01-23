package highlight

import (
	"path/filepath"
	"strings"
)

// Language mappings from file extensions to Chroma language identifiers
var extensionToLanguage = map[string]string{
	// Go
	".go":  "go",
	".mod": "go",
	".sum": "text",

	// Web technologies
	".js":     "javascript",
	".jsx":    "javascript",
	".ts":     "typescript",
	".tsx":    "typescript",
	".html":   "html",
	".htm":    "html",
	".css":    "css",
	".scss":   "scss",
	".sass":   "sass",
	".vue":    "vue",
	".svelte": "svelte",

	// Python
	".py":  "python",
	".pyx": "python",
	".pyi": "python",
	".pyw": "python",

	// Rust
	".rs": "rust",

	// C/C++
	".c":   "c",
	".h":   "c",
	".cc":  "cpp",
	".cpp": "cpp",
	".cxx": "cpp",
	".hpp": "cpp",
	".hxx": "cpp",

	// Java/Kotlin
	".java": "java",
	".kt":   "kotlin",
	".kts":  "kotlin",

	// C#
	".cs": "csharp",

	// Shell
	".sh":   "bash",
	".bash": "bash",
	".zsh":  "bash",
	".fish": "fish",

	// Config files
	".json": "json",
	".yaml": "yaml",
	".yml":  "yaml",
	".toml": "toml",
	".xml":  "xml",
	".ini":  "ini",

	// Markup
	".md":  "markdown",
	".tex": "latex",

	// Other languages
	".rb":    "ruby",
	".php":   "php",
	".lua":   "lua",
	".r":     "r",
	".sql":   "sql",
	".vim":   "vim",
	".zig":   "zig",
	".dart":  "dart",
	".swift": "swift",
	".scala": "scala",
	".clj":   "clojure",
	".hs":    "haskell",
	".ml":    "ocaml",
	".ex":    "elixir",
	".exs":   "elixir",
	".erl":   "erlang",
	".pl":    "perl",
}

// Common filename mappings (without extension)
var filenameToLanguage = map[string]string{
	"Dockerfile":       "docker",
	"dockerfile":       "docker",
	"Makefile":         "makefile",
	"makefile":         "makefile",
	"Vagrantfile":      "ruby",
	"Gemfile":          "ruby",
	"Rakefile":         "ruby",
	"CMakeLists.txt":   "cmake",
	".gitignore":       "text",
	".gitconfig":       "ini",
	"requirements.txt": "text",
	"package.json":     "json",
	"tsconfig.json":    "json",
	"cargo.toml":       "toml",
	"pyproject.toml":   "toml",
}

// DetectLanguage determines the programming language from a file path
func DetectLanguage(filePath string) string {
	if filePath == "" {
		return ""
	}

	// Extract filename and extension
	filename := filepath.Base(filePath)
	extension := strings.ToLower(filepath.Ext(filePath))

	// Try filename-based detection first
	if language, exists := filenameToLanguage[filename]; exists {
		return language
	}

	// Try extension-based detection
	if language, exists := extensionToLanguage[extension]; exists {
		return language
	}

	// Return empty string if no language detected
	return ""
}

// GetSupportedExtensions returns a list of supported file extensions
func GetSupportedExtensions() []string {
	extensions := make([]string, 0, len(extensionToLanguage))
	for ext := range extensionToLanguage {
		extensions = append(extensions, ext)
	}
	return extensions
}

// GetSupportedFilenames returns a list of supported special filenames
func GetSupportedFilenames() []string {
	filenames := make([]string, 0, len(filenameToLanguage))
	for filename := range filenameToLanguage {
		filenames = append(filenames, filename)
	}
	return filenames
}

// IsTextFile returns true if the file is likely a text file that can be highlighted
func IsTextFile(filePath string) bool {
	language := DetectLanguage(filePath)
	return language != "" && language != "text"
}
