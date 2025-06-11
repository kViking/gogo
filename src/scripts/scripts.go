package scripts

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
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

// LoadScripts loads all scripts from the user_scripts.json file in user config dir
func LoadScripts() (Scripts, error) {
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

// SaveScripts saves all scripts to the user_scripts.json file in user config dir
func SaveScripts(scripts Scripts) error {
	data, err := json.MarshalIndent(scripts, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(getUserScriptsPath(), data, 0644)
}

// GetVariableDescription returns the description for a variable or a default
func GetVariableDescription(varName string, config ScriptConfig) string {
	desc := config.Variables[varName]
	if desc == "" {
		desc = fmt.Sprintf(DefaultDesc, varName)
	}
	return desc
}

// promptForVariable asks the user to input a value for a variable
func promptForVariable(varName string, config ScriptConfig) string {
	desc := GetVariableDescription(varName, config)
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

// GetGadgetNames returns a sorted list of all gadget names
func GetGadgetNames() []string {
	scriptsMap, _ := LoadScripts()
	var names []string
	for name := range scriptsMap {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// AddScriptCommands dynamically adds all script shortcuts as subcommands
func AddScriptCommands(root *cobra.Command) {
	scripts, err := LoadScripts()
	if err != nil {
		errorText(fmt.Sprintf("❌ Error loading user_scripts.json: %v", err))
		return // No scripts yet
	}

	for name, config := range scripts {
		varNames := ExtractVariables(config.Command)

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
			desc := GetVariableDescription(varName, config)
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
		scripts, err := LoadScripts()
		if err != nil {
			errorText(fmt.Sprintf("❌ Error loading user_scripts.json: %v", err))
			return
		}
		config, ok := scripts[name]
		if !ok {
			errorText(fmt.Sprintf("❌ Gadget '%s' not found.\n", name))
			return
		}
		varNames := ExtractVariables(config.Command)

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

		// Replace variables in the command using the centralized helper
		psCommand := ReplaceVariables(config.Command, vars)

		// Use strings.Builder for scriptContent
		var sb strings.Builder
		sb.WriteString("# " + config.Description + "\n")
		sb.WriteString(psCommand + "\n")
		scriptContent := sb.String()

		if err := runPowerShellScript(name, scriptContent); err != nil {
			errorText("❌ Error running your gadget. Please check your command and variable values.")
			errorText(fmt.Sprintf("Details: %v", err))
			_ = cmd.Help()
		} else {
			successText("✅ Gadget finished! If you expected output, check above.")
		}
	}
}

// NewRootCommand returns the root cobra.Command for CLI mode (verb-first)
func NewRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "gogogadget",
		Short: "GoGoGadget: Run your favorite PowerShell commands with easy shortcuts!",
		Long:  "GoGoGadget lets you save, run, and manage PowerShell command shortcuts with variables.",
	}

	// List gadgets
	rootCmd.AddCommand(InitialListGadgetsCommand())

	// Add a new gadget
	rootCmd.AddCommand(InitialAddGadgetCommand())

	// Delete a gadget
	rootCmd.AddCommand(InitialDeleteGadgetCommand())

	// Edit a gadget (launches TUI for editing that gadget)
	rootCmd.AddCommand(InitialEditGadgetCommand())

	// Alias: 'list' for 'gadgets'
	rootCmd.AddCommand(&cobra.Command{
		Use:        "list",
		Short:      "Alias for 'gadgets'",
		Deprecated: "use 'gadgets' instead",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Parent().SetArgs([]string{"gadgets"})
			cmd.Parent().Execute()
		},
	})

	return rootCmd
}
