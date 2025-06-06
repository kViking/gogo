package scripts

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
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

	// Prepare PowerShell command to tokenize and output as JSON
	psScript := fmt.Sprintf(`$input = '%s'; [System.Management.Automation.PSParser]::Tokenize($input, [ref]$null) | Select-Object Content,Type | ConvertTo-Json`, strings.ReplaceAll(cmdStr, "'", "''"))
	fmt.Println("[DEBUG] PowerShell command to tokenize:", psScript)
	cmd := exec.Command("pwsh", "-NoProfile", "-Command", psScript)

	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		fmt.Println("[DEBUG] PowerShell execution error:", err)
		return fmt.Errorf("failed to run PowerShell: %w", err)
	}

	fmt.Println("[DEBUG] Raw PowerShell output:", out.String())

	var tokens []PSToken
	if err := json.Unmarshal(out.Bytes(), &tokens); err != nil {
		fmt.Println("[DEBUG] JSON unmarshal error:", err)
		return fmt.Errorf("failed to parse PowerShell output: %w", err)
	}

	fmt.Printf("[DEBUG] Parsed tokens: %+v\n", tokens)

	// Highlight likely user-input tokens and describe them
	var highlighted []string
	var descriptions []string
	highlightColor := color.New(color.FgHiYellow, color.Bold).SprintFunc()
	descColor := color.New(color.FgCyan).SprintFunc()
	// For unique variable names
	varCounters := map[string]int{"string": 0, "number": 0, "path": 0}
	for _, t := range tokens {
		contentStr := fmt.Sprintf("%v", t.Content)
		typeStr := psTokenTypeToString(t.Type)
		fmt.Printf("[DEBUG] Token: Type=%s, Content=%v\n", typeStr, t.Content)
		var varName string
		switch typeStr {
		case "String":
			if isLikelyPath(contentStr) {
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
			highlighted = append(highlighted, highlightColor("{{"+varName+"}}"))
			descriptions = append(descriptions, fmt.Sprintf("%s %s", highlightColor(varName), descColor(fmt.Sprintf("\u2190 was '%s', likely a string value (%s)", contentStr, guessStringPurpose(contentStr)))))
		case "Number":
			varCounters["number"]++
			if varCounters["number"] == 1 {
				varName = "integer"
			} else {
				varName = fmt.Sprintf("integer%d", varCounters["number"])
			}
			highlighted = append(highlighted, highlightColor("{{"+varName+"}}"))
			descriptions = append(descriptions, fmt.Sprintf("%s %s", highlightColor(varName), descColor(fmt.Sprintf("\u2190 was '%s', likely a numeric value", contentStr))))
		default:
			highlighted = append(highlighted, contentStr)
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

// psTokenTypeToString converts a PowerShell token type to its string representation.
func psTokenTypeToString(t interface{}) string {
	switch v := t.(type) {
	case string:
		return v
	case float64:
		switch int(v) {
		case 0:
			return "None"
		case 1:
			return "Command"
		case 2:
			return "CommandArgument"
		case 3:
			return "String"
		case 4:
			return "Number"
		case 5:
			return "Variable"
		case 6:
			return "Parameter"
		case 7:
			return "StringExpandable"
		case 8:
			return "Operator"
		case 9:
			return "GroupStart"
		case 10:
			return "GroupEnd"
		case 11:
			return "Keyword"
		case 12:
			return "Comment"
		case 13:
			return "StatementSeparator"
		case 14:
			return "NewLine"
		case 15:
			return "LineContinuation"
		case 16:
			return "Position"
		default:
			return fmt.Sprintf("Unknown(%v)", v)
		}
	default:
		return fmt.Sprintf("Unknown(%v)", v)
	}
}

// guessStringPurpose tries to guess what a string is used for.
func guessStringPurpose(s string) string {
	// Simple heuristics
	if isLikelyPath(s) {
		return "file or folder path"
	}
	if isLikelyEncoding(s) {
		return "encoding type"
	}
	if isLikelyGuid(s) {
		return "GUID"
	}
	return "text or parameter"
}

func isLikelyPath(s string) bool {
	// Windows or Unix path
	return strings.Contains(s, `\`) || strings.Contains(s, `/`)
}

func isLikelyEncoding(s string) bool {
	encodings := []string{"UTF8", "ASCII", "Unicode", "UTF7", "UTF32", "BigEndianUnicode", "Default", "OEM"}
	for _, e := range encodings {
		if strings.EqualFold(s, e) {
			return true
		}
	}
	return false
}

func isLikelyGuid(s string) bool {
	guidRegex := regexp.MustCompile(`^[{(]?[0-9A-Fa-f]{8}(-[0-9A-Fa-f]{4}){3}-[0-9A-Fa-f]{12}[)}]?$`)
	return guidRegex.MatchString(s)
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
