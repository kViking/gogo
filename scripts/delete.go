package scripts

import (
	"fmt"

	"github.com/mattn/go-colorable"
	"github.com/spf13/cobra"
)

func NewDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete [script name]",
		Short: "Delete a script shortcut",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			scripts, _ := loadScripts()
			name := args[0]
			if _, ok := scripts[name]; !ok {
				colorText.Red(fmt.Sprintf("❌ Script '%s' not found.\n", name))
				return
			}
			delete(scripts, name)
			if err := saveScripts(scripts); err != nil {
				colorText.Red(fmt.Sprintf("❌ Error deleting script: %v", err))
				return
			}
			fmt.Fprintln(colorable.NewColorableStdout(), "\x1b[32m✅ Script '"+name+"' deleted!\x1b[0m")
		},
	}
	return cmd
}
