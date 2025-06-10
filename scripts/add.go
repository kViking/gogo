package scripts

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/mattn/go-colorable"
	"github.com/spf13/cobra"
)

func NewAddCommand() *cobra.Command {
	var scriptName, command, desc string

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a new GoGoGadget gadget",
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

			out := colorable.NewColorableStdout()
			fmt.Fprintln(out) // Blank line before add process
			colorText.Cyan("Add a new GoGoGadget gadget (user-defined command):")
			fmt.Fprintln(out)

			// Get command
			command, _ = cmd.Flags().GetString("command")
			if command == "" {
				fmt.Fprint(out, "\x1b[36m📝 Enter the PowerShell command this gadget will run (you can use \x1b[1;35m{{variable}}\x1b[0m\x1b[36m for variables you want to fill in each time): \x1b[0m")
				c, _ := reader.ReadString('\n')
				command = strings.TrimSpace(c)
			}

			// Get gadget name
			scriptName, _ = cmd.Flags().GetString("scriptname")
			nameRe := regexp.MustCompile(`^[A-Za-z0-9_\-]+$`)
			for {
				if scriptName == "" {
					fmt.Fprint(out, "\x1b[36m🔖 Enter gadget name: \x1b[0m")
					n, _ := reader.ReadString('\n')
					scriptName = strings.TrimSpace(n)
				}
				// Validate: no spaces, no punctuation
				if !nameRe.MatchString(scriptName) {
					colorText.Yellow("⚠️  Gadget names cannot contain spaces or punctuation. Use only letters, numbers, dashes, or underscores. Please enter a new name.")
					scriptName = ""
					continue
				}
				break
			}

			// Get description
			desc, _ = cmd.Flags().GetString("desc")
			if desc == "" {
				fmt.Fprint(out, "\x1b[36m💡 Enter gadget description: \x1b[0m")
				d, _ := reader.ReadString('\n')
				desc = strings.TrimSpace(d)
			}

			variables := map[string]string{}
			for _, v := range extractVariables(command) {
				val, _ := cmd.Flags().GetString(v)
				if val == "" {
					fmt.Fprintf(out, "\x1b[33m✏️  Describe variable '%s': \x1b[0m", v)
					vd, _ := reader.ReadString('\n')
					val = strings.TrimSpace(vd)
				}
				variables[v] = val
			}

			if scriptName == "" || command == "" {
				fmt.Fprintln(colorable.NewColorableStderr(), "\x1b[31m❌ Gadget name and command are required.\x1b[0m")
				return
			}
			scripts[scriptName] = ScriptConfig{
				Description: desc,
				Command:     command,
				Variables:   variables,
			}
			if err := saveScripts(scripts); err != nil {
				fmt.Fprintln(colorable.NewColorableStderr(), "\x1b[31m❌ Error saving gadget:\x1b[0m", err)
				return
			}
			fmt.Fprintln(out)
			colorText.Green("✅ Gadget added!")
			fmt.Fprintln(out)
		},
	}

	cmd.Flags().StringVar(&scriptName, "scriptname", "", "Name of the gadget")
	cmd.Flags().StringVar(&command, "command", "", "PowerShell command (use {{VARNAME}} for variables)")
	cmd.Flags().StringVar(&desc, "desc", "", "Gadget description")

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
