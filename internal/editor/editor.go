package editor

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// Editor represents an external editor configuration
type Editor struct {
	Name      string
	Path      string
	Args      []string
	UsesShell bool
}

// GetEditor returns the user's preferred editor based on environment variables
// and platform defaults, with proper fallback chain
func GetEditor() (*Editor, error) {
	// Try $EDITOR first
	if editorEnv := os.Getenv("EDITOR"); editorEnv != "" {
		editor, err := parseEditorString(editorEnv)
		if err == nil {
			return editor, nil
		}
	}

	// Try $VISUAL as fallback
	if visualEnv := os.Getenv("VISUAL"); visualEnv != "" {
		editor, err := parseEditorString(visualEnv)
		if err == nil {
			return editor, nil
		}
	}

	// Platform-specific defaults
	return getPlatformDefault()
}

// parseEditorString parses an editor string that may contain arguments
// Examples: "vim", "code --wait", "nvim -a -b"
func parseEditorString(editorStr string) (*Editor, error) {
	parts := strings.Fields(editorStr)
	if len(parts) == 0 {
		return nil, fmt.Errorf("empty editor string")
	}

	editorPath := parts[0]
	editorArgs := parts[1:]

	// Check if the editor binary exists
	if _, err := exec.LookPath(editorPath); err != nil {
		return nil, fmt.Errorf("editor not found: %s", editorPath)
	}

	// Determine the editor name for command building
	editorName := getEditorName(editorPath)

	return &Editor{
		Name:      editorName,
		Path:      editorPath,
		Args:      editorArgs,
		UsesShell: false,
	}, nil
}

// getEditorName extracts the editor name from the path for command building
func getEditorName(path string) string {
	// Extract basename and remove common suffixes
	name := path
	if lastSlash := strings.LastIndex(name, "/"); lastSlash != -1 {
		name = name[lastSlash+1:]
	}
	if lastBackslash := strings.LastIndex(name, "\\"); lastBackslash != -1 {
		name = name[lastBackslash+1:]
	}

	// Remove common extensions on Windows
	if runtime.GOOS == "windows" {
		if strings.HasSuffix(name, ".exe") {
			name = name[:len(name)-4]
		}
		if strings.HasSuffix(name, ".cmd") {
			name = name[:len(name)-4]
		}
		if strings.HasSuffix(name, ".bat") {
			name = name[:len(name)-4]
		}
	}

	return name
}

// getPlatformDefault returns the default editor for the current platform
func getPlatformDefault() (*Editor, error) {
	var defaultEditor string

	switch runtime.GOOS {
	case "windows":
		defaultEditor = "notepad"
	case "darwin":
		defaultEditor = "nano"
	default: // linux and other unix-like
		defaultEditor = "nano"
	}

	if _, err := exec.LookPath(defaultEditor); err != nil {
		return nil, fmt.Errorf("no editor available: please set $EDITOR environment variable")
	}

	return &Editor{
		Name:      defaultEditor,
		Path:      defaultEditor,
		Args:      []string{},
		UsesShell: false,
	}, nil
}

// BuildCommand creates an exec.Cmd to open the specified file at the given line
func (e *Editor) BuildCommand(filename string, lineNumber int) *exec.Cmd {
	return e.BuildCommandWithSpecialHandling(filename, lineNumber)
}

// getLineNumberArgs returns the appropriate arguments to specify a line number
func (e *Editor) getLineNumberArgs(lineNumber int) []string {
	switch e.Name {
	case "vim", "vi", "nvim":
		return []string{fmt.Sprintf("+%d", lineNumber)}
	case "emacs":
		return []string{fmt.Sprintf("+%d", lineNumber)}
	case "nano":
		return []string{fmt.Sprintf("+%d", lineNumber)}
	case "kak", "kakoune":
		return []string{fmt.Sprintf("+%d", lineNumber)}
	case "gedit":
		return []string{fmt.Sprintf("+%d", lineNumber)}
	case "micro":
		return []string{fmt.Sprintf("+%d", lineNumber)}
	case "notepad++":
		return []string{fmt.Sprintf("-n%d", lineNumber)}
	default:
		// Try the common +N format for unknown editors
		return []string{fmt.Sprintf("+%d", lineNumber)}
	}
}

// isGUIApp determines if the editor is a GUI application on macOS
func (e *Editor) isGUIApp() bool {
	guiApps := []string{"code", "vscode", "subl", "sublime_text", "atom", "textmate"}
	for _, app := range guiApps {
		if e.Name == app {
			return true
		}
	}
	return false
}

// BuildCommandWithSpecialHandling handles special cases for certain editors
func (e *Editor) BuildCommandWithSpecialHandling(filename string, lineNumber int) *exec.Cmd {
	args := make([]string, len(e.Args))
	copy(args, e.Args)

	switch e.Name {
	case "hx", "helix":
		// Helix uses filename:line format
		args = append(args, fmt.Sprintf("%s:%d", filename, lineNumber))
	case "code", "vscode":
		// VSCode uses --goto filename:line
		args = append(args, "--goto", fmt.Sprintf("%s:%d", filename, lineNumber))
	case "subl", "sublime_text":
		// Sublime Text uses filename:line format
		args = append(args, fmt.Sprintf("%s:%d", filename, lineNumber))
	case "atom":
		// Atom uses filename:line format
		args = append(args, fmt.Sprintf("%s:%d", filename, lineNumber))
	default:
		// Use standard line number args + filename
		args = append(args, e.getLineNumberArgs(lineNumber)...)
		args = append(args, filename)
	}

	if e.UsesShell {
		if runtime.GOOS == "windows" {
			return exec.Command("cmd.exe", "/c", e.Path+" "+strings.Join(args, " "))
		} else {
			return exec.Command("sh", "-c", e.Path+" "+strings.Join(args, " "))
		}
	} else if runtime.GOOS == "darwin" && e.isGUIApp() {
		// Use 'open -a' for GUI applications on macOS
		openArgs := []string{"-a", e.Path}
		openArgs = append(openArgs, args...)
		return exec.Command("open", openArgs...)
	} else {
		return exec.Command(e.Path, args...)
	}
}
