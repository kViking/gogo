// filepath: /home/ellieo/Documents/kscripts/scripts/firstrun.go
package scripts

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

// getUserConfirmation asks the user for confirmation and exits
// Exits with code 0 if confirmed, code 1 if not confirmed
// Updates the settings file to mark firstRun as false if confirmed
func getUserConfirmation() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Do you understand and wish to continue? (y/yes): ")
	response, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading response:", err)
		os.Exit(1)
	}

	// Trim any whitespace, newlines, and convert to lowercase
	response = strings.TrimSpace(strings.ToLower(response))

	// Accept y or yes (case-insensitive)
	if response == "y" || response == "yes" {
		// Update settings file
		updateSettingsFile()

		colorText.Green("✅ You're all set up! Run GoGoGadget --help to see available commands, or GoGoGadget [command] --help to see help for a command")
		os.Exit(0) // Exit with success code
	}

	// If the user doesn't confirm, exit with error code
	colorText.Yellow("⚠️ Operation cancelled")
	os.Exit(1)
}

// ShowFirstRunMessage displays the first-run warning message with the critical part in red
// and asks for user confirmation before proceeding
func ShowFirstRunMessage() {
	// Define the warning message as a string literal
	warningMsg := `*** ------------------------------- ***
You are running GoGoGadget for the first time! This is exciting! You need to know a couple of things:

`

	// Print the first part of the message
	fmt.Print(warningMsg)

	// Show the first sentence of point 1 in bright red
	fmt.Print("1. ")
	fmt.Print("\x1b[1;91mGoGoGadget does NOT have any checks for your PowerShell scripts.\x1b[0m")

	// Define the rest of the message as a string literal
	restOfMsg := ` It will run them as-is, with variables replaced exactly as you specify. Make sure you test your scripts before saving them with GoGoGadget!
2. Your scripts are stored in a json file in the app directory (wherever you installed GoGoGadget). You can edit this file directly if you want without fear of breaking anything, but there are robust built in tools to edit the shortcuts as well. GUI is planned for a future release.

Print this message again with 'GoGoGadget first-run' if you need to see it again.
You can always run 'GoGoGadget help' for instructions on how to use the tool.
*** ------------------------------- ***
`
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
			warningMsg := `*** ------------------------------- ***
You are running GoGoGadget for the first time! This is exciting! You need to know a couple of things:

`
			fmt.Print(warningMsg)

			// Show the first sentence of point 1 in bright red
			fmt.Print("1. ")
			fmt.Print("\x1b[1;91mGoGoGadget does NOT have any checks for your PowerShell scripts.\x1b[0m")

			// Define the rest of the message as a string literal
			restOfMsg := ` It will run them as-is, with variables replaced exactly as you specify. Make sure you test your scripts before saving them with GoGoGadget!
2. Your scripts are stored in a json file in the app directory (wherever you installed GoGoGadget). You can edit this file directly if you want without fear of breaking anything, but there are robust built in tools to edit the shortcuts as well. GUI is planned for a future release.

Print this message again with 'GoGoGadget first-run' if you need to see it again.
You can always run 'GoGoGadget help' for instructions on how to use the tool.
*** ------------------------------- ***
`
			fmt.Print(restOfMsg)

			// For this command, we get confirmation but don't exit
			reader := bufio.NewReader(os.Stdin)
			fmt.Print("Do you understand? (y/yes): ")
			response, _ := reader.ReadString('\n')
			response = strings.TrimSpace(strings.ToLower(response))

			if response == "y" || response == "yes" {
				colorText.Green("✅ Continuing with GoGoGadget")
			} else {
				colorText.Yellow("⚠️ Please review the information above")
			}
		},
	}

	return cmd
}
