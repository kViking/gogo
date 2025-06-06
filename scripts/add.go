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
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a new script shortcut",
		Run: func(cmd *cobra.Command, args []string) {
			scripts, _ := loadScripts()
			reader := bufio.NewReader(os.Stdin)
			fmt.Print("📝 Enter PowerShell command (use {{VARNAME}} for variables): ")
			command, _ := reader.ReadString('\n')
			command = strings.TrimSpace(command)
			fmt.Print("🔖 Enter script name: ")
			name, _ := reader.ReadString('\n')
			name = strings.TrimSpace(name)
			fmt.Print("💡 Enter script description: ")
			desc, _ := reader.ReadString('\n')
			desc = strings.TrimSpace(desc)
			variables := map[string]string{}
			for _, v := range extractVariables(command) {
				fmt.Printf("✏️  Describe variable '%s': ", v)
				vd, _ := reader.ReadString('\n')
				variables[v] = strings.TrimSpace(vd)
			}
			if name == "" || command == "" {
				color.New(color.FgRed).Fprintln(os.Stderr, "❌ Name and command are required.")
				return
			}
			scripts[name] = ScriptConfig{
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
