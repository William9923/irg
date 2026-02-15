package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/William9923/irg/internal/search"
	"github.com/William9923/irg/internal/ui"
	"github.com/William9923/irg/internal/updater"
	tea "github.com/charmbracelet/bubbletea"
)

// Version information embedded at build time
var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
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
	// Version flag
	var showVersion = flag.Bool("version", false, "Print version information and exit")
	flag.Var(&typeFlags, "type", "Include only files of type (can be used multiple times)")
	flag.Var(&typeNotFlags, "type-not", "Exclude files of type (can be used multiple times)")
	flag.Parse()

	// Handle --version early
	if *showVersion {
		fmt.Printf("irg version %s\nCommit: %s\nBuilt: %s\n", version, commit, date)
		os.Exit(0)
	}

	// Handle upgrade command: upgrade [<version>]
	// Existing flag package will treat subcommands as first non-flag arg
	if len(flag.Args()) > 0 && flag.Args()[0] == "upgrade" {
		var targetVersion string
		if len(flag.Args()) > 1 {
			targetVersion = flag.Args()[1]
		}

		up := &updater.Updater{Repo: "William9923/irg", Binary: "irg"}
		latest, needsUpdate, err := up.Check()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Update check failed: %v\n", err)
			os.Exit(1)
		}
		if targetVersion == "" {
			if !needsUpdate {
				fmt.Println("irg is already up to date.")
				os.Exit(0)
			}
			targetVersion = latest
		}
		fmt.Printf("Updating irg to version %s...\n", targetVersion)
		if err := up.Update(targetVersion); err != nil {
			fmt.Fprintf(os.Stderr, "Update failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Update complete. Please restart irg to run the new version.")
		os.Exit(0)
	}

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
