package scripts

import (
	"fmt"
	"strings"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/formatters"
	"github.com/alecthomas/chroma/lexers"
	"github.com/mattn/go-colorable"
	"github.com/spf13/cobra"
)

// --- Add Command ---
func NewAddCommand() *cobra.Command {
	var scriptName, command, desc string
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a new GoGoGadget gadget (all input must be provided via flags)",
		Long: `Add a new GoGoGadget gadget (user-defined command).

All required information must be provided using flags.

Example:
  GoGoGadget add --command "Get-ChildItem {{folder}}" --scriptname filecount --desc "Counts files" --folder "Folder to count"`,
		Run: func(cmd *cobra.Command, args []string) {
			store, err := NewGadgetStore()
			if err != nil {
				colorText.Style("error", "\u274c Error loading gadgets: %s", err.Error())
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
				colorText.Style("error", "\u274c Missing required flags: %s", strings.Join(missing, ", "))
				colorText.PrintStyledLine(
					StyledChunk{"\nExample:", "info"},
				)
				colorText.PrintStyledLine(
					StyledChunk{"  GoGoGadget add --command \"Get-ChildItem {{folder}}\" --scriptname filecount --desc \"Counts files\" --folder \"Folder to count\"", "title"},
				)
				cmd.Help()
				return
			}
			err = CreateGadget(store, scriptName, command, desc, variables)
			if err != nil {
				colorText.Style("error", "\u274c %s", err.Error())
				return
			}
			colorText.Style("success", "\u2705 Gadget added!")
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
		Short: "Edit an existing gadget. To rename a gadget, use --name. To rename a variable, use --<var> --name <newname>. To change a variable's description, use --<var> --description <desc>.",
		Long: `Edit an existing gadget and its variables.

To rename a gadget, use --name.
To rename a variable, use --<var> --name <newname>.
To change a variable's description, use --<var> --description <desc>.

Tip: Use 'GoGoGadget peek [gadget name]' to see the current variables and descriptions before editing.`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			store, err := NewGadgetStore()
			if err != nil {
				colorText.Style("error", "\u274c Could not load gadgets.")
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

			// Allow editing variable names and descriptions:
			if store != nil {
				gadget, ok := store.Get(name)
				if ok {
					for oldVar := range gadget.Variables {
						if cmd.Flags().Changed(oldVar) {
							// Rename variable: --oldVar --name newName
							if cmd.Flags().Changed("name") {
								newVar, _ := cmd.Flags().GetString("name")
								if newVar != "" && newVar != oldVar {
									if updates["rename_vars"] == nil {
										updates["rename_vars"] = map[string]string{}
									}
									updates["rename_vars"].(map[string]string)[oldVar] = newVar
								}
							}
							// Change variable description: --oldVar --description newDesc
							if cmd.Flags().Changed("description") {
								desc, _ := cmd.Flags().GetString("description")
								if desc != "" {
									if updates["var_descs"] == nil {
										updates["var_descs"] = map[string]string{}
									}
									updates["var_descs"].(map[string]string)[oldVar] = desc
								}
							}
						}
					}
				}
			}

			if len(updates) == 0 {
				cmd.Help()
				return
			}
			err = EditGadget(store, name, updates)
			if err != nil {
				colorText.Style("error", "\u274c %s", err.Error())
				return
			}
			colorText.Style("success", "\u2705 Gadget updated!")
			finalName := name
			if cmd.Flags().Changed("name") && newNameFlag != "" {
				finalName = newNameFlag
			}
			gadget, _ := store.Get(finalName)
			printGadget(finalName, gadget)
		},
	}

	// Add flags for variable renaming and description (dynamic, after command is loaded)
	cmd.PreRun = func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			store, err := NewGadgetStore()
			if err == nil {
				gadget, ok := store.Get(args[0])
				if ok {
					for v := range gadget.Variables {
						cmd.Flags().Bool(v, false, "Target variable '"+v+"' for renaming or description change")
					}
				}
			}
		}
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
				colorText.Style("error", "\u274c Error loading gadgets: %s", err.Error())
				return
			}
			name := args[0]
			if name == "" {
				colorText.Style("error", "\u274c Gadget name is required.")
				return
			}
			err = store.Delete(name)
			if err != nil {
				colorText.Style("error", "\u274c Gadget '%s' not found or could not be deleted.", name)
				return
			}
			if err := store.Save(); err != nil {
				colorText.Style("error", "\u274c Error deleting gadget: %v", err)
				return
			}
			colorText.Style("success", "\u2705 Gadget deleted!")
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
			colorText.PrintStyledLine(
				StyledChunk{"List of GoGoGadget gadgets (user-defined commands):", "info"},
			)
			colorText.PrintStyledLine(
				StyledChunk{"Gadget Name", "title"},
				StyledChunk{"  ", ""},
				StyledChunk{"Description", "title"},
			)
			for name, gadget := range gadgets {
				colorText.PrintStyledLine(
					StyledChunk{name, "variable"},
					StyledChunk{"  ", ""},
					StyledChunk{gadget.Description, ""},
				)
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
				colorText.Style("error", "\u274c Error loading user_scripts.json: %s", err.Error())
				return
			}
			gadget, ok := store.Get(args[0])
			if !ok {
				colorText.Style("error", "\u274c Gadget '%s' not found.", args[0])
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

	colorText.PrintStyledLine(
		StyledChunk{"", ""},
	)

	colorText.PrintStyledLine(
		StyledChunk{"\nGadget Name:", "title"},
	)
	colorText.PrintStyledLine(
		StyledChunk{"  ", ""},
		StyledChunk{name, "variable"},
	)

	colorText.PrintStyledLine(
		StyledChunk{"", ""},
	)

	colorText.PrintStyledLine(StyledChunk{"Description:", "title"})
	fmt.Fprintln(out, "  ", gadget.Description)

	colorText.PrintStyledLine(
		StyledChunk{"", ""},
	)

	colorText.PrintStyledLine(StyledChunk{"Command:", "title"})
	// Use analyze.go's syntax highlighting for PowerShell commands
	fmt.Fprintln(out, "  "+HighlightPowerShellChroma(gadget.Command))

	colorText.PrintStyledLine(
		StyledChunk{"", ""},
	)

	if len(gadget.Variables) > 0 {
		colorText.PrintStyledLine(StyledChunk{"Variables:", "title"})
		for v, desc := range gadget.Variables {
			if desc == "" {
				desc = fmt.Sprintf("Value for %s", v)
			}
			colorText.PrintStyledLine(
				StyledChunk{"  ", ""},
				StyledChunk{v, "variable"},
				StyledChunk{" ", ""},
				StyledChunk{desc, ""},
			)
		}
	} else {
		colorText.PrintStyledLine(StyledChunk{"Variables:", "title"})
		fmt.Fprintln(out, "  (none)")
	}
	fmt.Fprintln(out)
}

// HighlightPowerShellChroma uses Chroma to syntax highlight PowerShell commands for the CLI.
func HighlightPowerShellChroma(cmd string) string {
	lexer := lexers.Get("powershell")
	if lexer == nil {
		return cmd
	}
	iterator, err := lexer.Tokenise(nil, cmd)
	if err != nil {
		return cmd
	}
	var b strings.Builder
	formatter := formatters.TTY16m
	_ = formatter.Format(&b, chroma.MustNewStyle("swapoff", chroma.StyleEntries{}), iterator)
	return b.String()
}
