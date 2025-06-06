package scripts

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// ShowScriptVariables prints the variables and their descriptions for a given script name
func ShowScriptVariables(scriptName string) {
	scripts, err := loadScripts()
	if err != nil {
		color.New(color.FgRed).Fprintln(os.Stderr, "❌ Error loading user_scripts.json:", err)
		return
	}
	config, ok := scripts[scriptName]
	if !ok {
		color.New(color.FgRed).Fprintf(os.Stderr, "❌ Script '%s' not found.\n", scriptName)
		return
	}
	varNames := extractVariables(config.Command)
	if len(varNames) == 0 {
		color.New(color.FgYellow).Println("This shortcut has no variables.")
		return
	}
	color.New(color.FgCyan).Printf("Variables for '%s':\n", scriptName)
	for _, varName := range varNames {
		desc := getVariableDescription(varName, config)
		color.New(color.FgGreen).Printf("  %s: ", varName)
		fmt.Printf("%s\n", desc)
	}
}

// NewVariablesCommand returns a cobra.Command for 'variables [script]'
func NewVariablesCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "variables [script]",
		Short: "Show variables and their descriptions for a script shortcut",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ShowScriptVariables(args[0])
		},
	}
	return cmd
}
