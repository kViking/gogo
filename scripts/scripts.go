package scripts

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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

type ScriptConfig struct {
	Description string            `json:"description"`
	Command     string            `json:"command"`
	Variables   map[string]string `json:"variables"`
}

type Scripts map[string]ScriptConfig

// getUserScriptsPath returns the user-writable path for user_scripts.json
func getUserScriptsPath() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		// fallback to home dir
		dir, _ = os.UserHomeDir()
	}
	dir = filepath.Join(dir, "GoGoGadget")
	_ = os.MkdirAll(dir, 0755)
	return filepath.Join(dir, "user_scripts.json")
}

// loadScripts loads all scripts from the user_scripts.json file in user config dir
func loadScripts() (Scripts, error) {
	scriptsPath := getUserScriptsPath()
	if _, err := os.Stat(scriptsPath); os.IsNotExist(err) {
		emptyScripts := make(Scripts)
		data, _ := json.MarshalIndent(emptyScripts, "", "  ")
		os.WriteFile(scriptsPath, data, 0644)
		return emptyScripts, nil
	}
	data, err := os.ReadFile(scriptsPath)
	if err != nil {
		return nil, err
	}
	var scripts Scripts
	err = json.Unmarshal(data, &scripts)
	if err != nil {
		return nil, err
	}
	return scripts, nil
}

// saveScripts saves all scripts to the user_scripts.json file in user config dir
func saveScripts(scripts Scripts) error {
	data, err := json.MarshalIndent(scripts, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(getUserScriptsPath(), data, 0644)
}

// getVariableDescription returns the description for a variable or a default
func getVariableDescription(varName string, config ScriptConfig) string {
	desc := config.Variables[varName]
	if desc == "" {
		desc = fmt.Sprintf(DefaultDesc, varName)
	}
	return desc
}

// promptForVariable asks the user to input a value for a variable
func promptForVariable(varName string, config ScriptConfig) string {
	desc := getVariableDescription(varName, config)
	infoText(fmt.Sprintf("Enter %s: ", desc))

	var value string
	fmt.Scanln(&value)
	return value
}

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
	scripts, err := loadScripts()
	if err != nil {
		errorText(fmt.Sprintf("❌ Error loading user_scripts.json: %v", err))
		return // No scripts yet
	}

	for name, config := range scripts {
		varNames := extractVariables(config.Command)

		scriptCmd := &cobra.Command{
			Use:   name,
			Short: config.Description,
			Long: config.Description + `

Example usage:
  GoGoGadget ` + name + ` value1 value2
  GoGoGadget ` + name + ` -VAR1 value1 -VAR2 value2
`,
			Args: cobra.ArbitraryArgs,
			Run:  createScriptRunFunc(name, config),
		}

		// Add flags for each variable
		for _, varName := range varNames {
			desc := getVariableDescription(varName, config)
			scriptCmd.Flags().String(varName, "", desc)
		}

		root.AddCommand(scriptCmd)
	}
}

// createScriptRunFunc returns a function to run the script with variables
func createScriptRunFunc(name string, config ScriptConfig) func(*cobra.Command, []string) {
	return func(cmd *cobra.Command, args []string) {
		vars := make(map[string]string)

		// Always get the latest variable list from the script definition
		scripts, err := loadScripts()
		if err != nil {
			errorText(fmt.Sprintf("❌ Error loading user_scripts.json: %v", err))
			return
		}
		config, ok := scripts[name]
		if !ok {
			errorText(fmt.Sprintf("❌ Gadget '%s' not found.\n", name))
			return
		}
		varNames := extractVariables(config.Command)

		// First, try to match provided args to variables by order
		for i, varName := range varNames {
			val, _ := cmd.Flags().GetString(varName)
			if val == "" && i < len(args) && args[i] != "" {
				val = args[i]
			}
			vars[varName] = val
		}

		// Now prompt for any missing variables
		for _, varName := range varNames {
			if vars[varName] == "" {
				vars[varName] = promptForVariable(varName, config)
			}
		}

		// Replace variables in the command
		psCommand := config.Command
		for varName, value := range vars {
			psCommand = strings.ReplaceAll(psCommand, "{{"+varName+"}}", value)
		}

		// Create and run the script
		scriptContent := fmt.Sprintf("# %s\n%s\n", config.Description, psCommand)
		if err := runPowerShellScript(name, scriptContent); err != nil {
			errorText("❌ Error running your gadget. Please check your command and variable values.")
			errorText(fmt.Sprintf("Details: %v", err))
			_ = cmd.Help()
		} else {
			successText("✅ Gadget finished! If you expected output, check above.")
		}
	}
}

// createVariablesListFunc returns a function to list variables for a gadget
func createVariablesListFunc(name string, config ScriptConfig) func(*cobra.Command, []string) {
	return func(cmd *cobra.Command, args []string) {
		varNames := extractVariables(config.Command)
		if len(varNames) == 0 {
			warnText("This gadget has no variables.")
			return
		}
		infoText(fmt.Sprintf("Variables for '%s':", name))
		for _, varName := range varNames {
			desc := getVariableDescription(varName, config)
			successText(fmt.Sprintf("  %s: %s", varName, desc))
		}
	}
}
