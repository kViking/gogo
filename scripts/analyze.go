package scripts

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/formatters"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	psCommandChecker     *PowerShellCommandChecker
	psCommandCheckerInit bool
	refreshCommandsFlag  bool
)

// PowerShellCommandChecker caches the list of known PowerShell commands for efficient lookup
// and provides a method to check if a string is a known command.
type PowerShellCommandChecker struct {
	knownCommands map[string]struct{}
}

// NewPowerShellCommandChecker dynamically loads all available PowerShell command names (lowercased)
func NewPowerShellCommandChecker() *PowerShellCommandChecker {
	cmd := "pwsh -NoProfile -Command Get-Command | Select-Object -ExpandProperty Name"
	output, err := exec.Command("bash", "-c", cmd).Output()
	m := make(map[string]struct{})
	if err == nil {
		names := strings.Split(string(output), "\n")
		for _, n := range names {
			n = strings.ToLower(strings.TrimSpace(n))
			if n != "" {
				m[n] = struct{}{}
			}
		}
	}
	return &PowerShellCommandChecker{knownCommands: m}
}

func GetPowerShellCommandChecker() *PowerShellCommandChecker {
	if !psCommandCheckerInit {
		psCommandChecker = NewPowerShellCommandChecker()
		psCommandCheckerInit = true
	}
	return psCommandChecker
}

// RefreshPowerShellCommandChecker forces a reload of the known PowerShell commands
func RefreshPowerShellCommandChecker() {
	psCommandChecker = NewPowerShellCommandChecker()
	psCommandCheckerInit = true
}

func (p *PowerShellCommandChecker) IsKnownCommand(s string) bool {
	s = strings.ToLower(s)
	_, ok := p.knownCommands[s]
	return ok
}

// Analyze prompts for a PowerShell command, highlights it, and suggests likely user-input variables.
func Analyze(command ...string) error {
	if refreshCommandsFlag {
		RefreshPowerShellCommandChecker()
	}

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

	// Collect all tokens into a slice
	tokens := []chroma.Token{}
	for token := iterator(); token.Type != chroma.EOF.Type; token = iterator() {
		tokens = append(tokens, token)
	}

	// For highlighting, create a new iterator from the tokens (Chroma expects func() chroma.Token)
	highlightIter := func() func() chroma.Token {
		i := 0
		return func() chroma.Token {
			if i >= len(tokens) {
				return chroma.Token{Type: chroma.EOF.Type}
			}
			tok := tokens[i]
			i++
			return tok
		}
	}()
	formatter := formatters.Get("terminal16m")
	style := styles.Get("monokai")
	if err := formatter.Format(os.Stdout, style, highlightIter); err != nil {
		return fmt.Errorf("failed to format highlighted command: %w", err)
	}
	fmt.Println()

	// Suggest variables for string tokens using the same tokens slice
	var suggestions []string
	varCounters := map[string]int{"string": 0, "number": 0, "variable": 0, "path": 0}
	checker := GetPowerShellCommandChecker()

	// Path buffer for joining path-like tokens
	var pathBuffer []string
	flushPathBuffer := func() {
		if len(pathBuffer) > 0 {
			joined := strings.Join(pathBuffer, "")
			if isLikelyPath(joined) {
				varCounters["path"]++
				varName := "path"
				if varCounters["path"] > 1 {
					varName = fmt.Sprintf("path%d", varCounters["path"])
				}
				suggestions = append(suggestions, fmt.Sprintf("%s ← was %s", color.New(color.FgHiYellow, color.Bold).Sprint(varName), color.New(color.FgCyan).Sprint(joined)))
			}
			pathBuffer = nil
		}
	}

	for _, token := range tokens {
		typeStr := token.Type.String()
		if typeStr == "Name" || typeStr == "Punctuation" {
			// Accumulate possible path
			pathBuffer = append(pathBuffer, token.Value)
			continue
		} else {
			flushPathBuffer()
		}
		if strings.HasPrefix(typeStr, "Literal.String") {
			varCounters["string"]++
			varName := "string"
			if varCounters["string"] > 1 {
				varName = fmt.Sprintf("string%d", varCounters["string"])
			}
			suggestions = append(suggestions, fmt.Sprintf("%s ← was %s", color.New(color.FgHiYellow, color.Bold).Sprint(varName), color.New(color.FgCyan).Sprint(token.Value)))
			continue
		}
		if strings.HasPrefix(typeStr, "Literal.Number") {
			varCounters["number"]++
			varName := "number"
			if varCounters["number"] > 1 {
				varName = fmt.Sprintf("number%d", varCounters["number"])
			}
			suggestions = append(suggestions, fmt.Sprintf("%s ← was %s", color.New(color.FgHiYellow, color.Bold).Sprint(varName), color.New(color.FgCyan).Sprint(token.Value)))
			continue
		}
		if typeStr == "Name.Variable" {
			varCounters["variable"]++
			varName := "variable"
			if varCounters["variable"] > 1 {
				varName = fmt.Sprintf("variable%d", varCounters["variable"])
			}
			suggestions = append(suggestions, fmt.Sprintf("%s ← was %s", color.New(color.FgHiYellow, color.Bold).Sprint(varName), color.New(color.FgCyan).Sprint(token.Value)))
			continue
		}
		if typeStr == "Name" {
			if !checker.IsKnownCommand(token.Value) {
				varCounters["string"]++
				varName := "string"
				if varCounters["string"] > 1 {
					varName = fmt.Sprintf("string%d", varCounters["string"])
				}
				suggestions = append(suggestions, fmt.Sprintf("%s ← was %s", color.New(color.FgHiYellow, color.Bold).Sprint(varName), color.New(color.FgCyan).Sprint(token.Value)))
			}
		}
	}
	flushPathBuffer()

	if len(suggestions) > 0 {
		fmt.Println(color.New(color.FgGreen, color.Bold).Sprint("\nSuggested variables:"))
		for _, s := range suggestions {
			fmt.Println("  ", s)
		}
	} else {
		fmt.Println(color.New(color.FgYellow, color.Bold).Sprint("\nNo suggestions found."))
	}
	return nil
}

// Helper to detect likely Windows/Unix paths
func isLikelyPath(s string) bool {
	if len(s) < 3 {
		return false
	}
	if strings.Contains(s, ":\\") || strings.Contains(s, "/") {
		return true
	}
	return false
}

// NewAnalyzeCommand creates a new Cobra command for analyze.
func NewAnalyzeCommand() *cobra.Command {
	var command string
	var refreshCommands bool

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
			refreshCommandsFlag = refreshCommands
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
	cmd.Flags().BoolVar(&refreshCommands, "refresh-commands", false, "Force a refresh of the known PowerShell commands")

	return cmd
}
