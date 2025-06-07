package scripts

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

type PSToken struct {
	Content interface{} `json:"Content"`
	Type    interface{} `json:"Type"`
}

var knownCommands map[string]struct{}

// getKnownCommands dynamically loads all available PowerShell command names
func getKnownCommands() map[string]struct{} {
	if knownCommands != nil {
		return knownCommands
	}
	cmd := exec.Command("pwsh", "-NoProfile", "-Command", "Get-Command | Select-Object -ExpandProperty Name | ConvertTo-Json")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		knownCommands = map[string]struct{}{}
		return knownCommands
	}
	var names []string
	if err := json.Unmarshal(out.Bytes(), &names); err != nil {
		knownCommands = map[string]struct{}{}
		return knownCommands
	}
	m := make(map[string]struct{}, len(names))
	for _, n := range names {
		m[strings.ToLower(n)] = struct{}{}
	}
	knownCommands = m
	return knownCommands
}

// Helper to check if a string is a known PowerShell command
func isKnownCommand(s string) bool {
	s = strings.ToLower(s)
	if strings.Contains(s, "-") {
		parts := strings.SplitN(s, "-", 2)
		if len(parts) == 2 && len(parts[0]) > 0 && len(parts[1]) > 0 {
			return true
		}
	}
	_, ok := getKnownCommands()[s]
	return ok
}

// Analyze prompts for a PowerShell command, tokenizes it, and highlights likely user-input sections with descriptions.
// If command is empty, it prompts the user.
func Analyze(command ...string) error {
	var cmdStr string
	if len(command) > 0 && strings.TrimSpace(strings.Join(command, " ")) != "" {
		cmdStr = strings.Join(command, " ")
	} else {
		fmt.Print("Enter your PowerShell command: ")
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		cmdStr = strings.TrimSpace(input)
	}

	// Define highlightColor and descColor at the top of the function so they are available everywhere
	highlightColor := color.New(color.FgHiYellow, color.Bold).SprintFunc()
	descColor := color.New(color.FgCyan).SprintFunc()

	// Use PowerShell AST to get parameter-value pairs
	psScript := fmt.Sprintf(`
$ast = [System.Management.Automation.Language.Parser]::ParseInput('%s', [ref]$null, [ref]$null)
$cmdAsts = $ast.FindAll({$args[0] -is [System.Management.Automation.Language.CommandAst]}, $true)
@($cmdAsts | ForEach-Object {
    [PSCustomObject]@{
        CommandName = $_.CommandElements[0].Value
        Elements = $_.CommandElements | Select-Object @{n='Type';e={ $_.GetType().Name }}, @{n='Text';e={ $_.ToString() }}
    }
}) | ConvertTo-Json -Compress -Depth 5
`, strings.ReplaceAll(cmdStr, "'", "''"))
	cmd := exec.Command("pwsh", "-NoProfile", "-Command", psScript)

	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to run PowerShell: %w", err)
	}

	jsonOut := strings.TrimSpace(out.String())
	if jsonOut == "" {
		return fmt.Errorf("PowerShell AST output was empty")
	}

	// If the output is a single object (starts with '{'), wrap it in a list for robust parsing
	if strings.HasPrefix(jsonOut, "{") {
		jsonOut = "[" + jsonOut + "]"
	}

	type CommandAstResult struct {
		CommandName string `json:"CommandName"`
		Elements    []struct {
			Type string `json:"Type"`
			Text string `json:"Text"`
		} `json:"Elements"`
	}

	var cmdResults []CommandAstResult
	if err := json.Unmarshal([]byte(jsonOut), &cmdResults); err != nil {
		// Print error and return, but do not print usage/help
		fmt.Fprintf(os.Stderr, "Error: failed to parse PowerShell AST output: %v\nRaw output: %s\n", err, jsonOut)
		return fmt.Errorf("failed to parse PowerShell AST output")
	}
	if len(cmdResults) == 0 {
		fmt.Fprintln(os.Stderr, "No command AST found.")
		return fmt.Errorf("no command AST found")
	}

	// Print the user's original command, using the custom highlighter
	fmt.Println(color.New(color.FgGreen, color.Bold).Sprint("\nOriginal command:"))
	var origHighlighted []string
	ast := cmdResults[0].Elements
	for i := 0; i < len(ast); i++ {
		t := ast[i]
		if t.Type == "StringConstantExpressionAst" && isKnownCommand(t.Text) {
			origHighlighted = append(origHighlighted, highlightColor(t.Text))
		} else if t.Type == "CommandParameterAst" {
			origHighlighted = append(origHighlighted, color.New(color.FgCyan, color.Bold).Sprint(t.Text))
		} else if t.Type == "StringConstantExpressionAst" {
			origHighlighted = append(origHighlighted, color.New(color.FgHiWhite).Sprint(t.Text))
		} else {
			origHighlighted = append(origHighlighted, t.Text)
		}
	}
	fmt.Println(strings.Join(origHighlighted, " "))

	var highlighted []string
	var descriptions []string
	varCounters := map[string]int{"string": 0, "path": 0}

	// For each command found in the AST, process and reconstruct the string for highlighting
	for _, cmdAst := range cmdResults {
		astTokens := cmdAst.Elements
		commandIdx := -1
		for i, t := range astTokens {
			if t.Type == "StringConstantExpressionAst" && isKnownCommand(t.Text) {
				commandIdx = i
				break
			}
		}

		for i := 0; i < len(astTokens); i++ {
			t := astTokens[i]
			if i == commandIdx {
				// This is the command name, print as-is
				highlighted = append(highlighted, t.Text)
				continue
			}
			// Only use the string Type (from GetType().Name) for all logic
			if t.Type == "StringConstantExpressionAst" {
				prevIsParam := false
				for j := i - 1; j >= 0; j-- {
					if astTokens[j].Type == "CommandAst" {
						continue
					}
					if astTokens[j].Type == "CommandParameterAst" {
						prevIsParam = true
					}
					break
				}
				if !prevIsParam && !isKnownCommand(t.Text) {
					varCounters["string"]++
					var varName string
					if varCounters["string"] == 1 {
						varName = "string"
					} else {
						varName = fmt.Sprintf("string%d", varCounters["string"])
					}
					highlighted = append(highlighted, highlightColor("{{"+varName+"}}"))
					descriptions = append(descriptions, fmt.Sprintf("%s %s", highlightColor(varName), descColor(fmt.Sprintf("\u2190 was '%s', positional argument (type: %s)", t.Text, t.Type))))
					continue
				}
			}
			if t.Type != "CommandParameterAst" {
				if t.Type != "PipelineAst" && t.Type != "ScriptBlockAst" && t.Type != "StatementBlockAst" && t.Type != "CommandExpressionAst" {
					highlighted = append(highlighted, t.Text)
				}
				continue
			}
			param := t.Text
			isValue := false
			if i+1 < len(astTokens) {
				val := astTokens[i+1]
				isValue = val.Type != "CommandParameterAst" && val.Type != "CommandAst" && val.Type != "PipelineAst" && val.Type != "ScriptBlockAst" && val.Type != "StatementBlockAst" && val.Type != "CommandExpressionAst" && val.Type != "Keyword" && val.Type != "StatementSeparatorAst" && val.Text != "|" && !strings.HasSuffix(val.Text, "-Object") && !isKnownCommand(val.Text)
				if isValue {
					paramLower := strings.ToLower(strings.TrimLeft(param, "-"))
					var varName string
					if strings.Contains(paramLower, "path") || strings.Contains(paramLower, "file") || strings.Contains(paramLower, "dir") || isLikelyPath(val.Text) {
						varCounters["path"]++
						if varCounters["path"] == 1 {
							varName = "path"
						} else {
							varName = fmt.Sprintf("path%d", varCounters["path"])
						}
					} else {
						varCounters["string"]++
						if varCounters["string"] == 1 {
							varName = "string"
						} else {
							varName = fmt.Sprintf("string%d", varCounters["string"])
						}
					}
					highlighted = append(highlighted, param+" "+highlightColor("{{"+varName+"}}"))
					descriptions = append(descriptions, fmt.Sprintf("%s %s", highlightColor(varName), descColor(fmt.Sprintf("\u2190 was '%s', value for %s (type: %s)", val.Text, param, val.Type))))
					i++ // skip value
					continue
				}
			}
			highlighted = append(highlighted, highlightColor(param))
			descriptions = append(descriptions, fmt.Sprintf("%s %s", highlightColor(param), descColor("\u2190 flag parameter (no value) or command name follows")))
		}
		// Add a separator between commands if there are multiple
		if len(cmdResults) > 1 {
			highlighted = append(highlighted, color.New(color.FgMagenta, color.Bold).Sprint("|"))
		}
	}

	fmt.Println(color.New(color.FgGreen, color.Bold).Sprint("\nSuggested parameterization:"))
	fmt.Println(strings.Join(highlighted, " "))

	if len(descriptions) > 0 {
		fmt.Println(color.New(color.FgGreen, color.Bold).Sprint("\nDescriptions of highlighted sections:"))
		for _, desc := range descriptions {
			fmt.Println("  " + desc)
		}
	}
	return nil
}

func isLikelyPath(s string) bool {
	// Windows or Unix path
	return strings.Contains(s, `\`) || strings.Contains(s, `/`)
}

// NewAnalyzeCommand creates a new Cobra command for analyze.
func NewAnalyzeCommand() *cobra.Command {
	var command string

	cmd := &cobra.Command{
		Use:   "analyze [command]",
		Short: "Analyze a PowerShell command and highlight likely user input sections",
		Long: `Analyze a PowerShell one-liner and highlight sections that are likely to be user input, such as file paths, strings, or numbers.

If you don't know what parts of your command to make into variables, this is where to start. This tool uses PowerShell's tokenizer to make educated guesses about which parts of your command might work. Drop in your working command, and it will suggest how to parameterize it.

Highlighted sections and their descriptions are guesses only—please review, test, and adjust as needed for your use case.

Examples:
  gogogadget analyze --command "Get-Content 'C:\\Users\\me\\file.txt' -Encoding UTF8"
  gogogadget analyze Get-Content 'C:\\Users\\me\\file.txt' -Encoding UTF8
  gogogadget analyze
  # (then enter your command at the prompt)

Example output:

  Suggested parameterization:
    Get-Content ["C:\Users\me\file.txt"] -Encoding [UTF8]

  Descriptions of highlighted sections:
    "C:\Users\me\file.txt" ← likely a string value (file or folder path)
    UTF8 ← likely a string value (encoding type)
`,
		Example: `  gogogadget analyze --command "Get-Content 'C:\\Users\\me\\file.txt' -Encoding UTF8"
  gogogadget analyze Get-Content 'C:\\Users\\me\\file.txt' -Encoding UTF8
  gogogadget analyze
  # (then enter your command at the prompt)
`,
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Priority: --command flag > positional args > prompt
			if command != "" {
				return Analyze(command)
			}
			if len(args) > 0 {
				return Analyze(strings.Join(args, " "))
			}
			return Analyze()
		},
	}

	cmd.Flags().StringVar(&command, "command", "", "PowerShell command to analyze (optional, can also be provided as arguments)")

	return cmd
}
