package scripts

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/mattn/go-colorable"
	"github.com/spf13/cobra"
)

// --- Add Command ---
func NewAddCommand() *cobra.Command {
	var scriptName, command, desc string
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a new GoGoGadget gadget",
		PreRun: func(cmd *cobra.Command, args []string) {
			c, _ := cmd.Flags().GetString("command")
			for _, v := range ExtractVariables(c) {
				if cmd.Flags().Lookup(v) == nil {
					cmd.Flags().String(v, "", fmt.Sprintf("Description for variable '%s'", v))
				}
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			store, err := NewGadgetStore()
			if err != nil {
				colorText.Red("\u274c Error loading gadgets: " + err.Error())
				return
			}
			reader := bufio.NewReader(os.Stdin)
			out := colorable.NewColorableStdout()
			fmt.Fprintln(out)
			colorText.Cyan("Add a new GoGoGadget gadget (user-defined command):")
			fmt.Fprintln(out)

			command, _ = cmd.Flags().GetString("command")
			if command == "" {
				fmt.Fprint(out, "\x1b[36m📝 Enter the PowerShell command this gadget will run (you can use \x1b[1;35m{{variable}}\x1b[0m\x1b[36m for variables you want to fill in each time): \x1b[0m")
				c, _ := reader.ReadString('\n')
				command = strings.TrimSpace(c)
			}

			scriptName, _ = cmd.Flags().GetString("scriptname")
			for {
				if scriptName == "" {
					fmt.Fprint(out, "\x1b[36m🔖 Enter gadget name: \x1b[0m")
					n, _ := reader.ReadString('\n')
					scriptName = strings.TrimSpace(n)
				}
				if err := ValidateGadgetName(scriptName); err != nil {
					colorText.Yellow("\u26a0\ufe0f  " + err.Error() + " Please enter a new name.")
					scriptName = ""
					continue
				}
				break
			}

			desc, _ = cmd.Flags().GetString("desc")
			if desc == "" {
				fmt.Fprint(out, "\x1b[36m💡 Enter gadget description: \x1b[0m")
				d, _ := reader.ReadString('\n')
				desc = strings.TrimSpace(d)
			}

			variables := map[string]string{}
			for _, v := range ExtractVariables(command) {
				val, _ := cmd.Flags().GetString(v)
				if val == "" {
					fmt.Fprintf(out, "\x1b[33m✏️  Describe variable '%s': \x1b[0m", v)
					vd, _ := reader.ReadString('\n')
					val = strings.TrimSpace(vd)
				}
				variables[v] = val
			}

			err = CreateGadget(store, scriptName, command, desc, variables)
			if err != nil {
				colorText.Red("\u274c " + err.Error())
				return
			}
			fmt.Fprintln(out)
			colorText.Green("\u2705 Gadget added!")
			fmt.Fprintln(out)

			gadget, _ := store.Get(scriptName)
			printGadget(scriptName, gadget)
		},
	}
	cmd.Flags().StringVar(&scriptName, "scriptname", "", "Name of the gadget")
	cmd.Flags().StringVar(&command, "command", "", "PowerShell command (use {{VARNAME}} for variables)")
	cmd.Flags().StringVar(&desc, "desc", "", "Gadget description")
	return cmd
}

// --- Edit Command ---
func NewEditCommand() *cobra.Command {
	var newNameFlag, newDescFlag, newCmdFlag string
	cmd := &cobra.Command{
		Use:   "edit [gadget name]",
		Short: "Edit an existing gadget",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			store, err := NewGadgetStore()
			if err != nil {
				colorText.Red("\u274c Could not load gadgets.")
				return
			}
			name := args[0]
			updates := map[string]interface{}{}
			if cmd.Flags().Changed("name") {
				updates["name"] = newNameFlag
			}
			if cmd.Flags().Changed("description") {
				updates["description"] = newDescFlag
			}
			if cmd.Flags().Changed("command") {
				updates["command"] = newCmdFlag
			}
			if len(updates) == 0 {
				cmd.Help()
				return
			}
			err = EditGadget(store, name, updates)
			if err != nil {
				colorText.Red("\u274c " + err.Error())
				return
			}
			colorText.Green("\u2705 Gadget updated!")
			finalName := name
			if cmd.Flags().Changed("name") && newNameFlag != "" {
				finalName = newNameFlag
			}
			gadget, _ := store.Get(finalName)
			printGadget(finalName, gadget)
		},
	}
	cmd.Flags().StringVar(&newNameFlag, "name", "", "New gadget name")
	cmd.Flags().StringVar(&newDescFlag, "description", "", "New gadget description")
	cmd.Flags().StringVar(&newCmdFlag, "command", "", "New PowerShell command")
	return cmd
}

// --- Delete Command ---
func NewDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete [gadget name]",
		Short: "Delete a GoGoGadget gadget (user-defined command)",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			store, err := NewGadgetStore()
			if err != nil {
				colorText.Red("\u274c Error loading gadgets: " + err.Error())
				return
			}
			name := args[0]
			if name == "" {
				fmt.Fprintln(colorable.NewColorableStdout(), "\x1b[31m\u274c Gadget name is required.\x1b[0m")
				return
			}
			err = store.Delete(name)
			if err != nil {
				colorText.Red(fmt.Sprintf("\u274c Gadget '%s' not found or could not be deleted.\n", name))
				return
			}
			if err := store.Save(); err != nil {
				colorText.Red(fmt.Sprintf("\u274c Error deleting gadget: %v", err))
				return
			}
			colorText.Green("\u2705 Gadget deleted!")
		},
	}
	return cmd
}

// --- List Command ---
func NewListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all gadgets",
		Run: func(cmd *cobra.Command, args []string) {
			store, err := NewGadgetStore()
			if err != nil {
				fmt.Fprintln(colorable.NewColorableStderr(), "\x1b[31m\u274c Error loading gadgets:\x1b[0m", err)
				return
			}
			gadgets := store.List()
			if len(gadgets) == 0 {
				fmt.Fprintln(colorable.NewColorableStdout(), "\x1b[36mNo gadgets found. Add one with 'GoGoGadget add'.\x1b[0m")
				return
			}
			fmt.Fprintln(colorable.NewColorableStdout(), "\x1b[36mList of GoGoGadget gadgets (user-defined commands):\x1b[0m")
			fmt.Fprintf(colorable.NewColorableStdout(), "\x1b[36m%-20s  %-40s  \x1b[0m\n", "Gadget Name", "Description")
			for name, gadget := range gadgets {
				fmt.Fprintf(colorable.NewColorableStdout(), "\x1b[1;35m%-20s\x1b[0m  %-40s\n", name, gadget.Description)
			}
		},
	}
	return cmd
}

// --- Peek Command ---
func NewPeekCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "peek [gadget]",
		Short: "Show all details for a gadget (name, description, command, variables)",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			store, err := NewGadgetStore()
			if err != nil {
				colorText.Red("\u274c Error loading user_scripts.json: " + err.Error())
				return
			}
			gadget, ok := store.Get(args[0])
			if !ok {
				colorText.Red(fmt.Sprintf("\u274c Gadget '%s' not found.", args[0]))
				return
			}
			printGadget(args[0], gadget)
		},
	}
	return cmd
}

// printGadget pretty-prints all info about a gadget, with syntax highlighting for the command.
func printGadget(name string, gadget Gadget) {
	out := colorable.NewColorableStdout()
	fmt.Fprintln(out, "\n\x1b[1;36mGadget Name:\x1b[0m      \x1b[1;35m"+name+"\x1b[0m")
	fmt.Fprintln(out, "\x1b[1;36mDescription:\x1b[0m      "+gadget.Description)
	fmt.Fprintln(out, "\x1b[1;36mCommand:\x1b[0m")
	fmt.Fprintln(out, "  "+highlightPowerShell(gadget.Command))
	if len(gadget.Variables) > 0 {
		fmt.Fprintln(out, "\x1b[1;36mVariables:\x1b[0m")
		for v, desc := range gadget.Variables {
			if desc == "" {
				desc = fmt.Sprintf("Value for %s", v)
			}
			fmt.Fprintf(out, "  \x1b[1;35m%-15s\x1b[0m %s\n", v, desc)
		}
	} else {
		fmt.Fprintln(out, "\x1b[1;36mVariables:\x1b[0m  (none)")
	}
	fmt.Fprintln(out)
}

// highlightPowerShell does basic syntax highlighting for PowerShell commands.
func highlightPowerShell(cmd string) string {
	keywords := []string{"Get-", "Set-", "Write-", "ForEach-", "If", "Else", "Function", "Param", "Return"}
	for _, kw := range keywords {
		cmd = strings.ReplaceAll(cmd, kw, "\x1b[1;34m"+kw+"\x1b[0m")
	}
	return cmd
}
