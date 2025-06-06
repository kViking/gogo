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
	Type    string      `json:"Type"`
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
	for _, t := range tokens {
		contentStr := fmt.Sprintf("%v", t.Content)
		fmt.Printf("[DEBUG] Token: Type=%s, Content=%v\n", t.Type, t.Content)
		switch t.Type {
		case "String":
			highlighted = append(highlighted, highlightColor("["+contentStr+"]"))
			descriptions = append(descriptions, fmt.Sprintf("%s %s", highlightColor(contentStr), descColor("\u2190 likely a string value ("+guessStringPurpose(contentStr)+")")))
		case "Number":
			highlighted = append(highlighted, highlightColor("["+contentStr+"]"))
			descriptions = append(descriptions, fmt.Sprintf("%s %s", highlightColor(contentStr), descColor("\u2190 likely a numeric value")))
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
