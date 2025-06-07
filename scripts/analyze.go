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

	// Use PowerShell AST and Tokenizer to get full token list and AST analysis
	psScript := fmt.Sprintf(`
$ErrorActionPreference = 'Stop'
$source = @'
%s
'@
$tokens = [System.Management.Automation.PSParser]::Tokenize($source, [ref]$null)
$ast = [System.Management.Automation.Language.Parser]::ParseInput($source, [ref]$null, [ref]$null)
$cmdAsts = $ast.FindAll({$args[0] -is [System.Management.Automation.Language.CommandAst]}, $true)
[PSCustomObject]@{
    Tokens = @($tokens | Select-Object Type, Content)
    Commands = @($cmdAsts | ForEach-Object {
        [PSCustomObject]@{
            CommandName = if ($_.CommandElements.Count -gt 0) { $_.CommandElements[0].Value } else { $null }
            Elements = @($_.CommandElements | Select-Object @{n='Type';e={ $_.GetType().Name }}, @{n='Text';e={ $_.ToString() }})
        }
    })
} | ConvertTo-Json -Depth 5
`, strings.ReplaceAll(cmdStr, "'", "''"))
	cmd := exec.Command("pwsh", "-NoProfile", "-Command", psScript)

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "PowerShell error output: %s\n", stderr.String())
		return fmt.Errorf("failed to run PowerShell: %w", err)
	}

	jsonOut := strings.TrimSpace(out.String())
	if jsonOut == "" {
		fmt.Fprintf(os.Stderr, "PowerShell error output: %s\n", stderr.String())
		return fmt.Errorf("PowerShell AST output was empty")
	}

	// Check if output is valid JSON (starts with '{' and ends with '}')
	if !strings.HasPrefix(jsonOut, "{") || !strings.HasSuffix(jsonOut, "}") {
		fmt.Fprintf(os.Stderr, "PowerShell output was not valid JSON. Raw output:\n%s\n", jsonOut)
		return fmt.Errorf("PowerShell output was not valid JSON")
	}

	// Parse the combined output
	type Token struct {
		Type    int    `json:"Type"`
		Content string `json:"Content"`
	}
	type CommandAstResult struct {
		CommandName string `json:"CommandName"`
		Elements    []struct {
			Type string `json:"Type"`
			Text string `json:"Text"`
		} `json:"Elements"`
	}
	type CombinedResult struct {
		Tokens   []Token            `json:"Tokens"`
		Commands []CommandAstResult `json:"Commands"`
	}
	var result CombinedResult
	if err := json.Unmarshal([]byte(jsonOut), &result); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to parse PowerShell output: %v\nRaw output: %s\n", err, jsonOut)
		return fmt.Errorf("failed to parse PowerShell output")
	}

	// Helper: map PSToken type int to string name for highlighting
	psTokenTypeName := func(t int) string {
		switch t {
		case 0:
			return "Unknown"
		case 1:
			return "Command"
		case 2:
			return "CommandParameter"
		case 3:
			return "Number"
		case 4:
			return "String"
		case 5:
			return "Variable"
		case 6:
			return "Member"
		case 7:
			return "LoopLabel"
		case 8:
			return "Attribute"
		case 9:
			return "Type"
		case 10:
			return "Keyword"
		case 11:
			return "Comment"
		case 12:
			return "StatementSeparator"
		case 13:
			return "GroupStart"
		case 14:
			return "GroupEnd"
		case 15:
			return "CurlyBracketStart"
		case 16:
			return "CurlyBracketEnd"
		case 17:
			return "SquareBracketStart"
		case 18:
			return "SquareBracketEnd"
		case 19:
			return "LineContinuation"
		case 20:
			return "NewLine"
		case 21:
			return "Whitespace"
		case 22:
			return "StringExpandable"
		case 23:
			return "HereStringExpandable"
		case 24:
			return "HereString"
		case 25:
			return "Operator"
		case 26:
			return "VariableExpandable"
		case 27:
			return "EmbeddedCommand"
		default:
			return "Other"
		}
	}

	// Build a set of AST argument values for highlighting
	astArgs := make(map[string]struct{})
	debugTypeColor := color.New(color.FgMagenta).SprintFunc()
	debugValColor := color.New(color.FgYellow).SprintFunc()
	for _, cmd := range result.Commands {
		for _, el := range cmd.Elements {
			// Compact, colored debug output: [T:Type]V:Value
			fmt.Fprintf(os.Stderr, "%s%s%s%s ", debugTypeColor("[T:"), debugTypeColor(el.Type), debugTypeColor("]V:"), debugValColor(el.Text))
			if (el.Type == "StringConstantExpressionAst" || el.Type == "ExpandableStringExpressionAst") && !isKnownCommand(el.Text) {
				astArgs[el.Text] = struct{}{}
			}
		}
	}
	fmt.Fprintln(os.Stderr) // Newline after debug output

	// Print the user's original command, syntax highlighted using the tokenizer and AST
	fmt.Println(color.New(color.FgGreen, color.Bold).Sprint("\nOriginal command:"))
	var origHighlighted []string
	for _, t := range result.Tokens {
		if _, isArg := astArgs[t.Content]; isArg {
			origHighlighted = append(origHighlighted, highlightColor("{{"+t.Content+"}}"))
		} else if isKnownCommand(t.Content) {
			origHighlighted = append(origHighlighted, highlightColor(t.Content))
		} else if psTokenTypeName(t.Type) == "CommandParameter" {
			origHighlighted = append(origHighlighted, color.New(color.FgCyan, color.Bold).Sprint(t.Content))
		} else {
			origHighlighted = append(origHighlighted, t.Content)
		}
	}
	fmt.Println(strings.Join(origHighlighted, " "))

	// Print parameterization for each command as before
	var descriptions []string
	varCounters := map[string]int{"string": 0, "path": 0}
	for _, cmdAst := range result.Commands {
		astTokens := cmdAst.Elements
		var cmdHighlighted []string
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
				cmdHighlighted = append(cmdHighlighted, highlightColor(t.Text))
				continue
			}
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
					cmdHighlighted = append(cmdHighlighted, highlightColor("{{"+varName+"}}"))
					descriptions = append(descriptions, fmt.Sprintf("%s %s", highlightColor(varName), descColor(fmt.Sprintf("\u2190 was '%s', positional argument (type: %s)", t.Text, t.Type))))
					continue
				}
				cmdHighlighted = append(cmdHighlighted, color.New(color.FgHiWhite).Sprint(t.Text))
				continue
			}
			if t.Type == "CommandParameterAst" {
				cmdHighlighted = append(cmdHighlighted, color.New(color.FgCyan, color.Bold).Sprint(t.Text))
				continue
			}
			cmdHighlighted = append(cmdHighlighted, t.Text)
		}
		fmt.Println(color.New(color.FgGreen, color.Bold).Sprint("\nSuggested parameterization:"))
		fmt.Println(strings.Join(cmdHighlighted, " "))
	}

	if len(descriptions) > 0 {
		fmt.Println(color.New(color.FgGreen, color.Bold).Sprint("\nDescriptions of highlighted sections:"))
		for _, desc := range descriptions {
			fmt.Println("  " + desc)
		}
	}
	return nil
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
