package scripts

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/mattn/go-colorable"
	"github.com/spf13/cobra"
)

const (
	DefaultDesc = "Value for %s"
)

// Text styling helpers
var (
	errorText   = func(msg string) { fmt.Fprintln(colorable.NewColorableStdout(), "\x1b[31m"+msg+"\x1b[0m") }
	successText = func(msg string) { fmt.Fprintln(colorable.NewColorableStdout(), "\x1b[32m"+msg+"\x1b[0m") }
	infoText    = func(msg string) { fmt.Fprintln(colorable.NewColorableStdout(), "\x1b[36m"+msg+"\x1b[0m") }
	warnText    = func(msg string) { fmt.Fprintln(colorable.NewColorableStdout(), "\x1b[33m"+msg+"\x1b[0m") }
)

// runPowerShellScript executes a PowerShell script with the given content
func runPowerShellScript(scriptName, content string) error {
	tmpFile, err := os.CreateTemp("", scriptName+"_*.ps1")
	if err != nil {
		return fmt.Errorf("error creating temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(content); err != nil {
		return fmt.Errorf("error writing script content: %w", err)
	}
	tmpFile.Close()

	// Use pwsh if available, otherwise fall back to powershell
	shellCmd := "pwsh"
	if _, err := exec.LookPath(shellCmd); err != nil {
		shellCmd = "powershell"
	}

	cmd := exec.Command(shellCmd, "-File", tmpFile.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// AddScriptCommands dynamically adds all script shortcuts as subcommands
func AddScriptCommands(root *cobra.Command) {
	// Implementation depends on Gadget and GadgetStore from gadget.go
}

// createScriptRunFunc returns a function to run the script with variables
func createScriptRunFunc(name string, config interface{}) func(*cobra.Command, []string) {
	return func(cmd *cobra.Command, args []string) {
		// Implementation depends on Gadget and GadgetStore from gadget.go
	}
}

// createVariablesListFunc returns a function to list variables for a gadget
func createVariablesListFunc(name string, config interface{}) func(*cobra.Command, []string) {
	return func(cmd *cobra.Command, args []string) {
		// Implementation depends on Gadget and GadgetStore from gadget.go
	}
}
