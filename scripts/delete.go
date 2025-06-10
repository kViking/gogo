package scripts

import (
	"fmt"

	"github.com/mattn/go-colorable"
	"github.com/spf13/cobra"
)

func NewDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete [gadget name]",
		Short: "Delete a GoGoGadget gadget (user-defined command)",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			out := colorable.NewColorableStdout()
			fmt.Fprint(out, "\x1b[36m🗑️  Enter the name of the gadget to delete: \x1b[0m")
			scripts, _ := loadScripts()
			name := args[0]
			if name == "" {
				fmt.Fprintln(out, "\x1b[31m❌ Gadget name is required.\x1b[0m")
				return
			}
			if _, ok := scripts[name]; !ok {
				colorText.Red(fmt.Sprintf("❌ Gadget '%s' not found.\n", name))
				return
			}
			delete(scripts, name)
			if err := saveScripts(scripts); err != nil {
				colorText.Red(fmt.Sprintf("❌ Error deleting gadget: %v", err))
				return
			}
			colorText.Green("✅ Gadget deleted!")
		},
	}
	return cmd
}
