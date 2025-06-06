package main

import (
	"encoding/json"
	"gogo/scripts"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// Settings represents the user settings stored in settings.json
type Settings struct {
	FirstRun bool `json:"firstRun"`
}

// checkFirstRun checks if this is the first time the program is run
// and displays a warning message if it is
func checkFirstRun() {
	// First try the current directory
	cwdSettings := "settings.json"
	execDirSettings := ""
	settings := Settings{
		FirstRun: true, // Default to true unless we find a settings file
	}

	// Check if settings exists in current directory
	if _, err := os.Stat(cwdSettings); err == nil {
		// Settings file exists in current directory, read it
		data, err := os.ReadFile(cwdSettings)
		if err == nil && len(data) > 0 {
			// Successfully read the file and it's not empty
			if err := json.Unmarshal(data, &settings); err != nil {
				// If we can't unmarshal, use default settings (FirstRun = true)
			}
		}
	} else {
		// Get the executable's directory path for checking there
		execPath, err := os.Executable()
		if err != nil {
			// Can't find executable path, use default settings
			// Already set FirstRun = true above
		} else {
			execDir := filepath.Dir(execPath)
			execDirSettings = filepath.Join(execDir, "settings.json")

			// Check if settings file exists in executable directory
			if _, err := os.Stat(execDirSettings); err == nil {
				// Settings file exists in executable directory, read it
				data, err := os.ReadFile(execDirSettings)
				if err == nil && len(data) > 0 {
					// Successfully read the file and it's not empty
					if err := json.Unmarshal(data, &settings); err != nil {
						// If we can't unmarshal, use default settings (FirstRun = true)
					}
				}
			}
		}
	}

	// Check if this is the first run (after either creating or reading settings)
	if settings.FirstRun {
		scripts.ShowFirstRunMessage()
	}
}

func main() {
	// Check if this is the first run and show warning message if needed
	// This must be the first thing we do to ensure the warning is shown before anything else
	checkFirstRun()

	rootCmd := &cobra.Command{
		Use:   "GoGoGadget",
		Short: "GoGoGadget: PowerShell script shortcuts made easy!",
		Long: `
GoGoGadget is a CLI tool for creating, managing, and running PowerShell script shortcuts with variable support.

Use 'GoGoGadget add' to create a new shortcut, 'GoGoGadget list' to see all, or run your scripts directly as subcommands!`,
	}

	rootCmd.AddCommand(scripts.NewAddCommand())
	rootCmd.AddCommand(scripts.NewListCommand())
	rootCmd.AddCommand(scripts.NewDeleteCommand())
	rootCmd.AddCommand(scripts.NewVariablesCommand())
	scripts.AddScriptCommands(rootCmd)
	scripts.AddEditCommand(rootCmd)

	if err := rootCmd.Execute(); err != nil {
		color.New(color.FgRed).Fprintf(os.Stderr, "❌ Error: %v\n", err)
		os.Exit(1)
	}
}
