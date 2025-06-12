package main

import (
	"encoding/json"
	"fmt"
	"gogo/scripts"
	"os"
	"path/filepath"

	"github.com/mattn/go-colorable"
	"github.com/spf13/cobra"
)

// Settings represents the user settings stored in settings.json
type Settings struct {
	FirstRun bool `json:"firstRun"`
}

// getSettingsPath returns the user-writable path for settings.json
func getSettingsPath() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		// fallback to home dir
		dir, _ = os.UserHomeDir()
	}
	dir = filepath.Join(dir, "GoGoGadget")
	_ = os.MkdirAll(dir, 0755)
	return filepath.Join(dir, "settings.json")
}

// checkFirstRun checks if this is the first time the program is run
// and displays a warning message if it is
func checkFirstRun() {
	settingsPath := getSettingsPath()
	settings := Settings{FirstRun: true}
	if _, err := os.Stat(settingsPath); err == nil {
		data, err := os.ReadFile(settingsPath)
		if err == nil && len(data) > 0 {
			_ = json.Unmarshal(data, &settings)
		}
	}
	if settings.FirstRun {
		scripts.ShowFirstRunMessage()
	}
}

func main() {
	// Check if this is the first run and show warning message if needed
	checkFirstRun()

	rootCmd := &cobra.Command{
		Use:   "GoGoGadget",
		Short: "GoGoGadget: PowerShell script shortcuts made easy!",
		Long: `
GoGoGadget is a CLI tool for creating, managing, and running PowerShell script shortcuts with variable support.

Use 'GoGoGadget add' to create a new shortcut, 'GoGoGadget list' to see all, or run your scripts directly as subcommands!`,
		Run: func(cmd *cobra.Command, args []string) {
			out := colorable.NewColorableStdout()
			fmt.Fprintln(out)
			fmt.Fprintln(out, "\x1b[1;36mGoGoGadget\x1b[0m: \x1b[1;37mRun your \x1b[1;35mgadgets\x1b[0m\x1b[1;37m (user-defined commands) easily!\x1b[0m")
			fmt.Fprintln(out)
			fmt.Fprintln(out, "\x1b[1;32m•\x1b[0m Use '\x1b[1;33mGoGoGadget add\x1b[0m' to create a new gadget, '\x1b[1;33mGoGoGadget list\x1b[0m' to see all gadgets, '\x1b[1;33mGoGoGadget edit\x1b[0m' to modify a gadget, and '\x1b[1;33mGoGoGadget delete\x1b[0m' to remove a gadget.")
			fmt.Fprintln(out, "\x1b[1;32m•\x1b[0m Each gadget runs a \x1b[1;36mPowerShell\x1b[0m command and can use variables (e.g., \x1b[1;35m{{variable}}\x1b[0m) for customization.")
			fmt.Fprintln(out)
		},
	}

	rootCmd.AddCommand(scripts.NewAddCommand())
	rootCmd.AddCommand(scripts.NewListCommand())
	rootCmd.AddCommand(scripts.NewDeleteCommand())
	rootCmd.AddCommand(scripts.NewAnalyzeCommand())
	rootCmd.AddCommand(scripts.NewPeekCommand())
	rootCmd.AddCommand(scripts.NewEditCommand())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(colorable.NewColorableStderr(), "\x1b[31m❌ Error: \x1b[0m", err)
		os.Exit(1)
	}
}
