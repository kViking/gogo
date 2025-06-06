package scripts

import (
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func NewListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all script shortcuts",
		Run: func(cmd *cobra.Command, args []string) {
			scripts, err := loadScripts()
			if err != nil {
				color.New(color.FgRed).Fprintln(os.Stderr, "❌ Error loading scripts:", err)
				return
			}
			if len(scripts) == 0 {
				color.New(color.FgYellow).Println("No scripts found.")
				return
			}
			color.New(color.FgCyan).Println("Available scripts:")
			for name, script := range scripts {
				color.New(color.FgGreen).Printf("- %s: %s\n", name, script.Description)
			}
		},
	}
	return cmd
}
