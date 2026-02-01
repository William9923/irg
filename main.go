package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/William9923/irg/internal/search"
	"github.com/William9923/irg/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
)

type arrayFlags []string

func (i *arrayFlags) String() string {
	return strings.Join(*i, ",")
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func main() {
	var caseFlag = flag.String("case", "smart", "Case sensitivity mode: smart, sensitive, insensitive")
	var typeFlags arrayFlags
	var typeNotFlags arrayFlags
	flag.Var(&typeFlags, "type", "Include only files of type (can be used multiple times)")
	flag.Var(&typeNotFlags, "type-not", "Exclude files of type (can be used multiple times)")
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
	model.SetFileTypes(typeFlags, typeNotFlags)

	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running irg: %v\n", err)
		os.Exit(1)
	}
}
