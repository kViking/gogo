package scripts

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func NewAddCommand() *cobra.Command {
	var scriptName, command, desc string

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a new script shortcut",
		PreRun: func(cmd *cobra.Command, args []string) {
			// If --command is provided, dynamically add flags for variables
			c, _ := cmd.Flags().GetString("command")
			for _, v := range extractVariables(c) {
				if cmd.Flags().Lookup(v) == nil {
					cmd.Flags().String(v, "", fmt.Sprintf("Description for variable '%s'", v))
				}
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			scripts, _ := loadScripts()
			reader := bufio.NewReader(os.Stdin)

			// Get command
			command, _ = cmd.Flags().GetString("command")
			if command == "" {
				fmt.Print("📝 Enter PowerShell command (use {{VARNAME}} for variables): ")
				c, _ := reader.ReadString('\n')
				command = strings.TrimSpace(c)
			}

			// Get script name
			scriptName, _ = cmd.Flags().GetString("scriptname")
			if scriptName == "" {
				fmt.Print("🔖 Enter script name: ")
				n, _ := reader.ReadString('\n')
				scriptName = strings.TrimSpace(n)
			}

			// Get description
			desc, _ = cmd.Flags().GetString("desc")
			if desc == "" {
				fmt.Print("💡 Enter script description: ")
				d, _ := reader.ReadString('\n')
				desc = strings.TrimSpace(d)
			}

			variables := map[string]string{}
			for _, v := range extractVariables(command) {
				val, _ := cmd.Flags().GetString(v)
				if val == "" {
					fmt.Printf("✏️  Describe variable '%s': ", v)
					vd, _ := reader.ReadString('\n')
					val = strings.TrimSpace(vd)
				}
				variables[v] = val
			}

			if scriptName == "" || command == "" {
				color.New(color.FgRed).Fprintln(os.Stderr, "❌ Name and command are required.")
				return
			}
			scripts[scriptName] = ScriptConfig{
				Description: desc,
				Command:     command,
				Variables:   variables,
			}
			if err := saveScripts(scripts); err != nil {
				color.New(color.FgRed).Fprintln(os.Stderr, "❌ Error saving script:", err)
				return
			}
			color.New(color.FgGreen).Println("✅ Script added!")
		},
	}

	cmd.Flags().StringVar(&scriptName, "scriptname", "", "Name of the script")
	cmd.Flags().StringVar(&command, "command", "", "PowerShell command (use {{VARNAME}} for variables)")
	cmd.Flags().StringVar(&desc, "desc", "", "Script description")

	return cmd
}

func extractVariables(command string) []string {
	var vars []string
	seen := map[string]bool{}
	re := regexp.MustCompile(`\{\{([A-Za-z0-9_]+)\}\}`)
	matches := re.FindAllStringSubmatch(command, -1)
	for _, m := range matches {
		if !seen[m[1]] {
			vars = append(vars, m[1])
			seen[m[1]] = true
		}
	}
	return vars
}
