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
			// Print a styled GoGoGadget intro: only the word GoGoGadget uses 'gogogadget', the rest uses 'title'
			scripts.ColorText{}.PrintStyledLine(
				scripts.StyledChunk{Text: "GoGoGadget", Style: "gogogadget"},
				scripts.StyledChunk{Text: ": Run your gadgets (user-defined commands) easily!", Style: "title"},
			)
			scripts.ColorText{}.PrintStyledLine(
				scripts.StyledChunk{Text: "• Use '", Style: ""},
				scripts.StyledChunk{Text: "GoGoGadget", Style: "gogogadget"},
				scripts.StyledChunk{Text: " add", Style: "highlight"},
				scripts.StyledChunk{Text: "' to create a new gadget, '", Style: ""},
				scripts.StyledChunk{Text: "GoGoGadget", Style: "gogogadget"},
				scripts.StyledChunk{Text: " list", Style: "highlight"},
				scripts.StyledChunk{Text: "' to see all gadgets, '", Style: ""},
				scripts.StyledChunk{Text: "GoGoGadget", Style: "gogogadget"},
				scripts.StyledChunk{Text: " edit", Style: "highlight"},
				scripts.StyledChunk{Text: "' to modify a gadget, and '", Style: ""},
				scripts.StyledChunk{Text: "GoGoGadget", Style: "gogogadget"},
				scripts.StyledChunk{Text: " delete", Style: "highlight"},
				scripts.StyledChunk{Text: "' to remove a gadget.", Style: ""},
			)
			scripts.ColorText{}.PrintStyledLine(
				scripts.StyledChunk{Text: "• Each gadget runs a PowerShell command and can use variables (e.g., ", Style: ""},
				scripts.StyledChunk{Text: "{{variable}}", Style: "highlight"},
				scripts.StyledChunk{Text: ") for customization.", Style: ""},
			)
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
