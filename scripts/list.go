package scripts

import (
	"fmt"

	"github.com/mattn/go-colorable"
	"github.com/spf13/cobra"
)

func NewListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all gadgets",
		Run: func(cmd *cobra.Command, args []string) {
			scripts, err := loadScripts()
			if err != nil {
				fmt.Fprintln(colorable.NewColorableStderr(), "\x1b[31m❌ Error loading gadgets:\x1b[0m", err)
				return
			}
			if len(scripts) == 0 {
				fmt.Fprintln(colorable.NewColorableStdout(), "\x1b[36mNo gadgets found. Add one with 'GoGoGadget add'.\x1b[0m")
				return
			}
			fmt.Fprintln(colorable.NewColorableStdout(), "\x1b[36mList of GoGoGadget gadgets (user-defined commands):\x1b[0m")
			fmt.Fprintf(colorable.NewColorableStdout(), "\x1b[36m%-20s  %-40s  \x1b[0m\n", "Gadget Name", "Description")
			for name, script := range scripts {
				fmt.Fprintf(colorable.NewColorableStdout(), "\x1b[1;35m%-20s\x1b[0m  %-40s\n", name, script.Description)
			}
		},
	}
	return cmd
}
