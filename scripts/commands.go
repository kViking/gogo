package scripts

import (
	"fmt"
	"strings"

	"github.com/mattn/go-colorable"
	"github.com/spf13/cobra"
)

// --- Add Command ---
func NewAddCommand() *cobra.Command {
	var scriptName, command, desc string
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a new GoGoGadget gadget (all input must be provided via flags)",
		Long:  `Add a new GoGoGadget gadget (user-defined command).\n\nAll required information must be provided using flags.\nExample:\n  GoGoGadget add --command "Get-ChildItem {{folder}}" --scriptname filecount --desc "Counts files" --folder "Folder to count"`,
		Run: func(cmd *cobra.Command, args []string) {
			store, err := NewGadgetStore()
			if err != nil {
				colorText.Red("\u274c Error loading gadgets: " + err.Error())
				return
			}
			command, _ = cmd.Flags().GetString("command")
			scriptName, _ = cmd.Flags().GetString("scriptname")
			desc, _ = cmd.Flags().GetString("desc")

			missing := []string{}
			if command == "" {
				missing = append(missing, "--command")
			}
			if scriptName == "" {
				missing = append(missing, "--scriptname")
			}
			if desc == "" {
				missing = append(missing, "--desc")
			}
			vars := ExtractVariables(command)
			variables := map[string]string{}
			for _, v := range vars {
				val, _ := cmd.Flags().GetString(v)
				if val == "" {
					missing = append(missing, "--"+v)
				} else {
					variables[v] = val
				}
			}
			if len(missing) > 0 {
				colorText.Red("\u274c Missing required flags: " + strings.Join(missing, ", "))
				fmt.Fprintln(colorable.NewColorableStdout(), "\nExample:")
				fmt.Fprint(colorable.NewColorableStdout(), "  ")
				colorText.Bold("GoGoGadget add --command \"Get-ChildItem {{folder}}\" --scriptname filecount --desc \"Counts files\" --folder \"Folder to count\"")
				cmd.Help()
				return
			}
			err = CreateGadget(store, scriptName, command, desc, variables)
			if err != nil {
				colorText.Red("\u274c " + err.Error())
				return
			}
			colorText.Green("\u2705 Gadget added!")
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
				fmt.Fprintln(colorable.NewColorableStdout(), "No gadgets found. Add one with 'GoGoGadget add'.")
				return
			}
			out := colorable.NewColorableStdout()
			fmt.Fprintln(out, "List of GoGoGadget gadgets (user-defined commands):")
			fmt.Fprintf(out, "%-20s  %-40s\n", "Gadget Name", "Description")
			for name, gadget := range gadgets {
				colorText.Magenta("%-20s", name)
				fmt.Fprintf(out, "  %-40s\n", gadget.Description)
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
	fmt.Fprintln(out, "\nGadget Name:      "+name)
	fmt.Fprintln(out, "Description:      "+gadget.Description)
	fmt.Fprintln(out, "Command:")
	fmt.Fprintln(out, "  "+highlightPowerShell(gadget.Command))
	if len(gadget.Variables) > 0 {
		fmt.Fprintln(out, "Variables:")
		for v, desc := range gadget.Variables {
			if desc == "" {
				desc = fmt.Sprintf("Value for %s", v)
			}
			colorText.Magenta("  %-15s", v)
			fmt.Fprintf(out, " %s\n", desc)
		}
	} else {
		fmt.Fprintln(out, "Variables:  (none)")
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
