package main

import (
	"fmt"
	"os"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/william-nobara/igrep/internal/ui"
)

func main() {
	if _, err := exec.LookPath("rg"); err != nil {
		fmt.Fprintln(os.Stderr, "Error: ripgrep (rg) is not installed or not in PATH")
		fmt.Fprintln(os.Stderr, "Please install ripgrep: https://github.com/BurntSushi/ripgrep#installation")
		os.Exit(1)
	}

	p := tea.NewProgram(
		ui.NewModel(),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running igrep: %v\n", err)
		os.Exit(1)
	}
}
