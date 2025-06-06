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

	// Use PowerShell AST to get parameter-value pairs
	psScript := fmt.Sprintf(`$ast = [System.Management.Automation.Language.Parser]::ParseInput('%s', [ref]$null, [ref]$null); $ast.FindAll({$args[0] -is [System.Management.Automation.Language.CommandAst]}, $true) | ForEach-Object { $_.CommandElements | ForEach-Object { [PSCustomObject]@{ Type = $_.GetType().Name; Text = $_.ToString() } } } | ConvertTo-Json`, strings.ReplaceAll(cmdStr, "'", "''"))
	fmt.Println("[DEBUG] PowerShell AST command to analyze:", psScript)
	cmd := exec.Command("pwsh", "-NoProfile", "-Command", psScript)

	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		fmt.Println("[DEBUG] PowerShell execution error:", err)
		return fmt.Errorf("failed to run PowerShell: %w", err)
	}

	fmt.Println("[DEBUG] Raw PowerShell AST output:", out.String())

	var astTokens []struct {
		Type string `json:"Type"`
		Text string `json:"Text"`
	}
	if err := json.Unmarshal(out.Bytes(), &astTokens); err != nil {
		fmt.Println("[DEBUG] JSON unmarshal error:", err)
		return fmt.Errorf("failed to parse PowerShell AST output: %w", err)
	}

	fmt.Printf("[DEBUG] Parsed AST tokens: %+v\n", astTokens)

	// Highlight parameter-value pairs
	var highlighted []string
	var descriptions []string
	highlightColor := color.New(color.FgHiYellow, color.Bold).SprintFunc()
	descColor := color.New(color.FgCyan).SprintFunc()
	varCounters := map[string]int{"string": 0, "number": 0, "path": 0}
	for i := 0; i < len(astTokens); i++ {
		t := astTokens[i]
		if t.Type == "CommandParameterAst" {
			param := t.Text
			// Check if next token is a value and not another parameter/command/pipe
			if i+1 < len(astTokens) {
				val := astTokens[i+1]
				// Only treat as a value if it's not another parameter, command, or pipe
				if val.Type != "CommandParameterAst" && val.Type != "CommandAst" && val.Type != "PipelineAst" && val.Type != "ScriptBlockAst" && val.Type != "StatementBlockAst" && val.Type != "CommandExpressionAst" && val.Type != "Keyword" && val.Type != "StatementSeparatorAst" && val.Text != "|" {
					// Heuristics for common PowerShell parameter names
					paramLower := strings.ToLower(strings.TrimLeft(param, "-"))
					var varName string
					if strings.Contains(paramLower, "path") || strings.Contains(paramLower, "file") || strings.Contains(paramLower, "dir") || isLikelyPath(val.Text) {
						varCounters["path"]++
						if varCounters["path"] == 1 {
							varName = "path"
						} else {
							varName = fmt.Sprintf("path%d", varCounters["path"])
						}
					} else if strings.Contains(paramLower, "count") || strings.Contains(paramLower, "size") || strings.Contains(paramLower, "length") || strings.Contains(paramLower, "number") || isNumeric(val.Text) {
						varCounters["number"]++
						if varCounters["number"] == 1 {
							varName = "integer"
						} else {
							varName = fmt.Sprintf("integer%d", varCounters["number"])
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
					descriptions = append(descriptions, fmt.Sprintf("%s %s", highlightColor(varName), descColor(fmt.Sprintf("\u2190 was '%s', value for %s", val.Text, param))))
					i++ // skip value
				} else {
					// Parameter is a flag (no value)
					highlighted = append(highlighted, highlightColor(param))
					descriptions = append(descriptions, fmt.Sprintf("%s %s", highlightColor(param), descColor("\u2190 flag parameter (no value)")))
				}
			} else {
				// Parameter is a flag (no value)
				highlighted = append(highlighted, highlightColor(param))
				descriptions = append(descriptions, fmt.Sprintf("%s %s", highlightColor(param), descColor("\u2190 flag parameter (no value)")))
			}
		} else {
			// Only show non-parameter tokens that are not just punctuation
			if t.Type != "StringConstantExpressionAst" && t.Type != "CommandAst" && t.Type != "PipelineAst" && t.Type != "ScriptBlockAst" && t.Type != "StatementBlockAst" && t.Type != "CommandExpressionAst" {
				highlighted = append(highlighted, t.Text)
			}
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

// Helper to check if a string is numeric
func isNumeric(s string) bool {
	_, err := fmt.Sscanf(s, "%f", new(float64))
	return err == nil
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
