package scripts

import (
	"fmt"

	"github.com/mattn/go-colorable"
	"github.com/spf13/cobra"
)

func NewListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all script shortcuts",
		Run: func(cmd *cobra.Command, args []string) {
			scripts, err := loadScripts()
			if err != nil {
				fmt.Fprintln(colorable.NewColorableStderr(), "\x1b[31m❌ Error loading scripts:\x1b[0m", err)
				return
			}
			if len(scripts) == 0 {
				colorText.Yellow("No scripts found.")
				return
			}
			fmt.Fprintln(colorable.NewColorableStdout(), "\x1b[36mAvailable scripts:\x1b[0m")
			for name, script := range scripts {
				fmt.Fprintf(colorable.NewColorableStdout(), "\x1b[32m- %s: %s\x1b[0m\n", name, script.Description)
			}
		},
	}
	return cmd
}
