package scripts

import (
	"fmt"

	"github.com/spf13/cobra"
)

// ShowScriptVariables prints the variables and their descriptions for a given script name
func ShowScriptVariables(scriptName string) {
	scripts, err := loadScripts()
	if err != nil {
		colorText.Red("❌ Error loading user_scripts.json: " + err.Error())
		return
	}
	config, ok := scripts[scriptName]
	if !ok {
		colorText.Red(fmt.Sprintf("❌ Script '%s' not found.", scriptName))
		return
	}
	varNames := extractVariables(config.Command)
	if len(varNames) == 0 {
		colorText.Yellow("This shortcut has no variables.")
		return
	}
	colorText.Cyan(fmt.Sprintf("Variables for '%s':", scriptName))
	for _, varName := range varNames {
		desc := getVariableDescription(varName, config)
		colorText.Green(fmt.Sprintf("  %s: ", varName))
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
