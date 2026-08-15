package main

import (
	"flag"
	"fmt"
	"os"

	"tboard/internal/config"
	"tboard/internal/domain"
	"tboard/internal/probe"
	"tboard/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

var version = "0.1.0"

func printUsage() {
	fmt.Printf("tboard v%s - Interactive 3D visual ping dashboard TUI\n\n", version)
	fmt.Println("Usage:")
	fmt.Println("  tboard [flags] [host[:port] ...]")
	fmt.Println("\nExamples:")
	fmt.Println("  tboard 1.1.1.1 8.8.8.8 google.com:443")
	fmt.Println("  tboard -c ~/.config/tboard/config.yaml")
	fmt.Println("  tboard -t dracula 1.1.1.1 github.com")
	fmt.Println("\nFlags:")
	fmt.Println("  -c, --config <file>   Path to config file (.json or .yaml)")
	fmt.Println("  -p, --port <port>     Default port when omitted (default: 80)")
	fmt.Println("  -t, --theme <name>    Initial theme: gruvbox-light, dracula, auto")
	fmt.Println("  -v, --version         Show version information")
	fmt.Println("  -h, --help            Show help message")
}

func main() {
	var (
		configFile  string
		defaultPort int
		themeName   string
		showVersion bool
		showHelp    bool
	)

	flag.StringVar(&configFile, "c", "", "Path to config file")
	flag.StringVar(&configFile, "config", "", "Path to config file")
	flag.IntVar(&defaultPort, "p", 80, "Default target port")
	flag.IntVar(&defaultPort, "port", 80, "Default target port")
	flag.StringVar(&themeName, "t", "gruvbox-light", "Initial theme")
	flag.StringVar(&themeName, "theme", "gruvbox-light", "Initial theme")
	flag.BoolVar(&showVersion, "v", false, "Show version")
	flag.BoolVar(&showVersion, "version", false, "Show version")
	flag.BoolVar(&showHelp, "h", false, "Show help")
	flag.BoolVar(&showHelp, "help", false, "Show help")

	flag.Usage = printUsage
	flag.Parse()

	if showHelp {
		printUsage()
		os.Exit(0)
	}

	if showVersion {
		fmt.Printf("tboard v%s\n", version)
		os.Exit(0)
	}

	var targets []domain.Target
	selectedThemeIdx := ui.ParseThemeName(themeName)

	if configFile == "" {
		configFile = config.DiscoverConfigFile()
	}

	if configFile != "" {
		cfg, err := config.LoadConfigFile(configFile)
		if err == nil {
			if cfg.DefaultPort > 0 {
				defaultPort = cfg.DefaultPort
			}
			if cfg.Theme != "" && themeName == "gruvbox-light" {
				selectedThemeIdx = ui.ParseThemeName(cfg.Theme)
			}
			for i, ct := range cfg.Targets {
				p := ct.Port
				if p <= 0 {
					p = defaultPort
				}
				targets = append(targets, domain.NewTarget(ct.Host, p, 50, 10, i))
			}
		} else if configFile != "" && flag.Lookup("c").Value.String() != "" {
			fmt.Fprintf(os.Stderr, "Error loading config file %s: %v\n", configFile, err)
			os.Exit(1)
		}
	}

	posArgs := flag.Args()
	if len(posArgs) > 0 {
		cliTargets := make([]domain.Target, 0, len(posArgs))
		for i, arg := range posArgs {
			host, port := probe.ParseHostPort(arg, defaultPort)
			cliTargets = append(cliTargets, domain.NewTarget(host, port, 50, 10, i))
		}
		targets = cliTargets
	}

	p := tea.NewProgram(ui.NewModel(targets, selectedThemeIdx), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running ping dashboard: %v\n", err)
		os.Exit(1)
	}
}
