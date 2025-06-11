package main

import (
	"gogo/src/analyze"
	"gogo/src/menu"
	scripts "gogo/src/scripts"
	"os"
)

func main() {
	rootCmd := scripts.NewRootCommand()
	rootCmd.AddCommand(analyze.InitialAnalyzeCommand())
	if len(os.Args) > 1 && os.Args[1] != "tui" {
		// Run as CLI (cobra)
		if err := rootCmd.Execute(); err != nil {
			os.Exit(1)
		}
		return
	}
	// If no args or 'tui', run TUI
	scripts.CheckFirstRun()
	menu.RunMenuTUI()
}
