// filepath: /home/ellieo/Documents/kscripts/scripts/firstrun.go
package scripts

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gogo/src/style"

	"github.com/spf13/cobra"
)

// Settings represents the user settings stored in settings.json
type Settings struct {
	FirstRun bool `json:"firstRun"`
}

// getSettingsPath returns the user-writable path for settings.json
func getSettingsPath() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		// fallback to home dir
		dir, _ = os.UserHomeDir()
	}
	dir = filepath.Join(dir, "GoGoGadget")
	_ = os.MkdirAll(dir, 0755)
	return filepath.Join(dir, "settings.json")
}

// updateSettingsFile updates the settings.json file to mark firstRun as false
func updateSettingsFile() {
	settingsPath := getSettingsPath()
	settings := Settings{FirstRun: false}
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		fmt.Println("Error marshaling settings:", err)
		return
	}
	if err := os.WriteFile(settingsPath, data, 0644); err != nil {
		fmt.Println("Error writing settings file:", err)
	}
}

// CheckFirstRun checks if this is the user's first run and shows the message if so
func CheckFirstRun() {
	settingsPath := getSettingsPath()
	settings := Settings{FirstRun: true}
	if _, err := os.Stat(settingsPath); err == nil {
		data, err := os.ReadFile(settingsPath)
		if err == nil && len(data) > 0 {
			_ = json.Unmarshal(data, &settings)
		}
	}
	if settings.FirstRun {
		ShowFirstRunMessage()
	}
}

// getUserConfirmation asks the user for confirmation and exits
// Exits with code 0 if confirmed, code 1 if not confirmed
// Updates the settings file to mark firstRun as false if confirmed
func getUserConfirmation() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(style.PromptStyle.Render("Do you understand and wish to continue? (y/yes): "))
	response, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println(style.ErrorStyle.Render("Error reading response:"), err)
		os.Exit(1)
	}

	// Trim any whitespace, newlines, and convert to lowercase
	response = strings.TrimSpace(strings.ToLower(response))

	// Accept y or yes (case-insensitive)
	if response == "y" || response == "yes" {
		// Update settings file
		updateSettingsFile()

		// Use styles from style.go for success message
		fmt.Println(style.SuccessStyle.Render("✅ You're all set up! Run GoGoGadget --help to see available commands, or GoGoGadget [gadget] --help to see help for a gadget"))
		os.Exit(0) // Exit with success code
	}

	// If the user doesn't confirm, exit with error code
	fmt.Println(style.InfoStyle.Render("⚠️ Operation cancelled"))
	os.Exit(1)
}

// ShowFirstRunMessage displays the first-run warning message with the critical part in red
// and asks for user confirmation before proceeding
func ShowFirstRunMessage() {
	// Define the warning message as a string literal
	warningMsg := style.HeaderStyle.Render("*** ------------------------------- ***\nYou are running GoGoGadget for the first time! This is exciting! You need to know a couple of things:\n\n")

	// Print the first part of the message
	fmt.Print(warningMsg)

	// Show the first sentence of point 1 in bright red
	fmt.Print(style.ErrorStyle.Render("GoGoGadget does NOT have any checks for your PowerShell commands."))

	// Define the rest of the message as a string literal
	restOfMsg := style.InfoStyle.Render(` It will run them as-is, with variables replaced exactly as you specify. Make sure you test your commands before saving them with GoGoGadget!
2. Your gadgets are stored in a json file in your $LOCALAPPDATA directory (check yours with $env:LOCALAPPDATA). You can edit this file directly if you want without fear of breaking anything, but there are robust built in tools to edit the shortcuts as well. GUI is planned for a future release.

Print this message again with 'GoGoGadget first-run' if you need to see it again.
You can always run 'GoGoGadget help' for instructions on how to use the tool.
*** ------------------------------- ***
`)
	fmt.Print(restOfMsg)

	// Get user confirmation, will exit if not confirmed
	getUserConfirmation()
}

// NewFirstRunCommand creates a new cobra command to show the first-run message again
func NewFirstRunCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "first-run",
		Short: "Show the first-run warning message again",
		Long: `Display the first-run warning message that appears when GoGoGadget is run for the first time.
This command is useful if you want to review the warning message again.`,
		Run: func(cmd *cobra.Command, args []string) {
			// For the first-run command, we want to show the message but not exit automatically
			// so we need a special version of the confirmation handling

			// First show the warning message
			// Define the warning message as a string literal
			warningMsg := style.HeaderStyle.Render("*** ------------------------------- ***\nYou are running GoGoGadget for the first time! This is exciting! You need to know a couple of things:\n\n")
			fmt.Print(warningMsg)

			// Show the first sentence of point 1 in bright red
			fmt.Print(style.ErrorStyle.Render("GoGoGadget does NOT have any checks for your PowerShell commands."))

			// Define the rest of the message as a string literal
			restOfMsg := style.InfoStyle.Render(` It will run them as-is, with variables replaced exactly as you specify. Make sure you test your commands before saving them with GoGoGadget!
2. Your gadgets are stored in a json file in the app directory (wherever you installed GoGoGadget). You can edit this file directly if you want without fear of breaking anything, but there are robust built in tools to edit the shortcuts as well. GUI is planned for a future release.

Print this message again with 'GoGoGadget first-run' if you need to see it again.
You can always run 'GoGoGadget help' for instructions on how to use the tool.
*** ------------------------------- ***
`)
			fmt.Print(restOfMsg)

			// For this command, we get confirmation but don't exit
			reader := bufio.NewReader(os.Stdin)
			fmt.Print(style.PromptStyle.Render("Do you understand? (y/yes): "))
			response, _ := reader.ReadString('\n')
			response = strings.TrimSpace(strings.ToLower(response))

			if response == "y" || response == "yes" {
				fmt.Println(style.SuccessStyle.Render("✅ Continuing with GoGoGadget"))
			} else {
				fmt.Println(style.InfoStyle.Render("⚠️ Please review the information above"))
			}
		},
	}

	return cmd
}

type firstrunModel struct {
	errorMsg string
	success  bool
}

func (m firstrunModel) View() string {
	header := style.HeaderStyle.Render("Welcome to GoGoGadget!")
	errorStyle := style.ErrorStyle
	successStyle := style.SuccessStyle
	help := style.MenuHelpStyle.Render("Enter: Confirm  Esc: Cancel  ←/→: Move")

	var s string
	s += header + "\n\n"
	if m.errorMsg != "" {
		s += errorStyle.Render(m.errorMsg) + "\n\n"
	}
	if m.success {
		s += successStyle.Render("🎉 Setup complete!") + "\n\n"
		s += help + "\n"
		return s
	}
	s += style.InfoStyle.Render("Let's get started by adding your first gadget!")
	s += "\n\n" + help
	return s
}
