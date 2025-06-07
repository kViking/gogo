package scripts

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/formatters"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// Analyze prompts for a PowerShell command, highlights it, and suggests likely user-input variables.
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

	fmt.Println(color.New(color.FgGreen, color.Bold).Sprint("\nOriginal command (syntax highlighted):"))
	lexer := lexers.Get("powershell")
	if lexer == nil {
		return fmt.Errorf("could not get PowerShell lexer")
	}
	iterator, err := lexer.Tokenise(nil, cmdStr)
	if err != nil {
		return fmt.Errorf("failed to tokenize command: %w", err)
	}
	formatter := formatters.Get("terminal16m")
	style := styles.Get("monokai")
	if err := formatter.Format(os.Stdout, style, iterator); err != nil {
		return fmt.Errorf("failed to format highlighted command: %w", err)
	}
	fmt.Println()

	// Suggest variables for string tokens
	iterator, err = lexer.Tokenise(nil, cmdStr)
	if iterator == nil {
		return fmt.Errorf("failed to tokenize command for variable suggestion")
	}
	var suggestions []string
	varCounters := map[string]int{"string": 0, "number": 0, "variable": 0}
	for token := iterator(); token.Type != chroma.EOF.Type; token = iterator() {
		typeStr := token.Type.String()
		if strings.HasPrefix(typeStr, "Literal.String") {
			varCounters["string"]++
			varName := "string"
			if varCounters["string"] > 1 {
				varName = fmt.Sprintf("string%d", varCounters["string"])
			}
			suggestions = append(suggestions, fmt.Sprintf("%s ← was %s", color.New(color.FgHiYellow, color.Bold).Sprint(varName), color.New(color.FgCyan).Sprint(token.Value)))
		}
		if strings.HasPrefix(typeStr, "Literal.Number") {
			varCounters["number"]++
			varName := "number"
			if varCounters["number"] > 1 {
				varName = fmt.Sprintf("number%d", varCounters["number"])
			}
			suggestions = append(suggestions, fmt.Sprintf("%s ← was %s", color.New(color.FgHiYellow, color.Bold).Sprint(varName), color.New(color.FgCyan).Sprint(token.Value)))
		}
		if typeStr == "Name.Variable" {
			varCounters["variable"]++
			varName := "variable"
			if varCounters["variable"] > 1 {
				varName = fmt.Sprintf("variable%d", varCounters["variable"])
			}
			suggestions = append(suggestions, fmt.Sprintf("%s ← was %s", color.New(color.FgHiYellow, color.Bold).Sprint(varName), color.New(color.FgCyan).Sprint(token.Value)))
		}
	}

	if len(suggestions) > 0 {
		fmt.Println(color.New(color.FgGreen, color.Bold).Sprint("\nSuggested variables:"))
		for _, s := range suggestions {
			fmt.Println("  ", s)
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

If you don't know what parts of your command to make into variables, this is where to start. This tool uses syntax highlighting to make educated guesses about which parts of your command might work. Drop in your working command, and it will suggest how to parameterize it.

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
