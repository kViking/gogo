package scripts

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// AddEditCommand adds the 'edit' command to the root command
func AddEditCommand(root *cobra.Command) {
	var newNameFlag string
	var newDescFlag string
	var newCmdFlag string
	var editCmd = &cobra.Command{
		Use:   "edit [script name]",
		Short: "Edit an existing script shortcut",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			scripts, err := loadScripts()
			if err != nil {
				color.New(color.FgRed).Fprintln(os.Stderr, "❌ Could not load scripts.")
				return
			}
			name := args[0]
			script, ok := scripts[name]
			if !ok {
				color.New(color.FgRed).Fprintf(os.Stderr, "❌ Script '%s' not found.\n", name)
				return
			}

			// If flags are set, edit directly and exit
			if cmd.Flags().Changed("name") {
				newName := newNameFlag
				if newName != "" && newName != name {
					scripts[newName] = script
					delete(scripts, name)
					name = newName
					_ = saveScripts(scripts)
					color.New(color.FgGreen).Println("✅ Script name updated.")
					return
				} else if newName == "" {
					reader := bufio.NewReader(os.Stdin)
					fmt.Printf("Current: %s\nEnter new name: ", name)
					input, _ := reader.ReadString('\n')
					input = strings.TrimSpace(input)
					if input != "" && input != name {
						scripts[input] = script
						delete(scripts, name)
						name = input
						_ = saveScripts(scripts)
						color.New(color.FgGreen).Println("✅ Script name updated.")
					}
					return
				}
			}
			if cmd.Flags().Changed("description") {
				newDesc := newDescFlag
				if newDesc != "" {
					script.Description = newDesc
					scripts[name] = script
					_ = saveScripts(scripts)
					color.New(color.FgGreen).Println("✅ Script description updated.")
					return
				} else {
					reader := bufio.NewReader(os.Stdin)
					fmt.Printf("Current: %s\nEnter new description: ", script.Description)
					input, _ := reader.ReadString('\n')
					input = strings.TrimSpace(input)
					if input != "" {
						script.Description = input
						scripts[name] = script
						_ = saveScripts(scripts)
						color.New(color.FgGreen).Println("✅ Script description updated.")
					}
					return
				}
			}
			if cmd.Flags().Changed("command") {
				newCmd := newCmdFlag
				if newCmd != "" {
					script.Command = newCmd
					scripts[name] = script
					_ = saveScripts(scripts)
					color.New(color.FgGreen).Println("✅ Script command updated.")
					return
				} else {
					reader := bufio.NewReader(os.Stdin)
					fmt.Printf("Current: %s\nEnter new command: ", script.Command)
					input, _ := reader.ReadString('\n')
					input = strings.TrimSpace(input)
					if input != "" {
						script.Command = input
						scripts[name] = script
						_ = saveScripts(scripts)
						color.New(color.FgGreen).Println("✅ Script command updated.")
					}
					return
				}
			}

			reader := bufio.NewReader(os.Stdin)
			for {
				fmt.Printf("\nEditing script: %s\n", name)
				fmt.Printf("1. Name: %s\n", name)
				fmt.Printf("2. Description: %s\n", script.Description)
				fmt.Printf("3. Command: %s\n", script.Command)
				fmt.Println("4. Variables:")
				idx := 5
				varKeys := []string{}
				for k, v := range script.Variables {
					fmt.Printf("   %d. %s: %s\n", idx, k, v)
					varKeys = append(varKeys, k)
					idx++
				}
				fmt.Println("0. Save and exit")
				fmt.Print("Choose what to edit (number): ")
				choiceRaw, _ := reader.ReadString('\n')
				choice := strings.TrimSpace(choiceRaw)
				switch choice {
				case "1":
					fmt.Printf("Current: %s\nEnter new name: ", name)
					newName, _ := reader.ReadString('\n')
					newName = strings.TrimSpace(newName)
					if newName != "" && newName != name {
						scripts[newName] = script
						delete(scripts, name)
						name = newName
					}
				case "2":
					fmt.Printf("Current: %s\nEnter new description: ", script.Description)
					desc, _ := reader.ReadString('\n')
					script.Description = strings.TrimSpace(desc)
					scripts[name] = script
				case "3":
					fmt.Printf("Current: %s\nEnter new command: ", script.Command)
					cmdStr, _ := reader.ReadString('\n')
					cmdStr = strings.TrimSpace(cmdStr)
					if cmdStr != "" {
						script.Command = cmdStr
						scripts[name] = script
					}
				case "0":
					_ = saveScripts(scripts)
					color.New(color.FgGreen).Println("✅ Script updated.")
					return
				default:
					// Check if editing a variable description
					idxNum := 0
					fmt.Sscanf(choice, "%d", &idxNum)
					if idxNum >= 5 && idxNum < 5+len(varKeys) {
						varKey := varKeys[idxNum-5]
						fmt.Printf("Current: %s\nEnter new description for variable '%s': ", script.Variables[varKey], varKey)
						newDesc, _ := reader.ReadString('\n')
						script.Variables[varKey] = strings.TrimSpace(newDesc)
						scripts[name] = script
					} else {
						color.New(color.FgRed).Println("Invalid choice.")
					}
				}
			}
		},
	}
	editCmd.Flags().StringVar(&newNameFlag, "name", "", "Edit the script's name directly")
	editCmd.Flags().StringVar(&newDescFlag, "description", "", "Edit the script's description directly")
	editCmd.Flags().StringVar(&newCmdFlag, "command", "", "Edit the script's command directly")
	root.AddCommand(editCmd)
}
