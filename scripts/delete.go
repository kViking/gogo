package scripts

import (
	"os"

	"github.com/fatih/color"
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
				color.New(color.FgRed).Fprintf(os.Stderr, "❌ Script '%s' not found.\n", name)
				return
			}
			delete(scripts, name)
			if err := saveScripts(scripts); err != nil {
				color.New(color.FgRed).Fprintln(os.Stderr, "❌ Error deleting script:", err)
				return
			}
			color.New(color.FgGreen).Printf("✅ Script '%s' deleted!\n", name)
		},
	}
	return cmd
}
