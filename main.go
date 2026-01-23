package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/William9923/igrep/internal/search"
	"github.com/William9923/igrep/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	var caseFlag = flag.String("case", "smart", "Case sensitivity mode: smart, sensitive, insensitive")
	flag.Parse()

	if _, err := exec.LookPath("rg"); err != nil {
		fmt.Fprintln(os.Stderr, "Error: ripgrep (rg) is not installed or not in PATH")
		fmt.Fprintln(os.Stderr, "Please install ripgrep: https://github.com/BurntSushi/ripgrep#installation")
		os.Exit(1)
	}

	var caseSensitivity search.CaseSensitivity
	switch strings.ToLower(*caseFlag) {
	case "smart":
		caseSensitivity = search.CaseSmart
	case "sensitive":
		caseSensitivity = search.CaseSensitive
	case "insensitive":
		caseSensitivity = search.CaseInsensitive
	default:
		fmt.Fprintln(os.Stderr, "Error: --case must be one of: smart, sensitive, insensitive")
		os.Exit(1)
	}

	model := ui.NewModel()
	model.SetCaseSensitivity(caseSensitivity)

	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running igrep: %v\n", err)
		os.Exit(1)
	}
}
