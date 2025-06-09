package scripts

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/quick"
	"github.com/mattn/go-colorable"
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

	out := colorable.NewColorableStdout()

	var cmdStr string
	if len(command) > 0 && strings.TrimSpace(strings.Join(command, " ")) != "" {
		cmdStr = strings.Join(command, " ")
	} else {
		fmt.Fprintln(out) // Ensure a blank line before the prompt
		fmt.Fprint(out, "Enter your PowerShell command: ")
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		cmdStr = strings.TrimSpace(input)
	}

	fmt.Fprintln(out) // Blank line before spinner
	spinner := GetSpinner("Analyzing command...")
	spinner.Start()

	lexer := lexers.Get("powershell")
	if lexer == nil {
		spinner.Stop()
		return fmt.Errorf("could not get PowerShell lexer")
	}
	iterator, err := lexer.Tokenise(nil, cmdStr)
	if err != nil {
		spinner.Stop()
		return fmt.Errorf("failed to tokenize command: %w", err)
	}

	// Collect all tokens into a slice
	tokens := []chroma.Token{}
	for token := iterator(); token.Type != chroma.EOF.Type; token = iterator() {
		tokens = append(tokens, token)
	}

	// Suggest variables for string tokens using the same tokens slice
	var suggestions []struct{ VarName, Original string }
	varCounters := map[string]int{"string": 0, "number": 0, "variable": 0, "path": 0}
	checker := GetPowerShellCommandChecker()

	// Path buffer for joining path-like tokens
	var pathBuffer []chroma.Token
	flushPathBuffer := func() {
		if len(pathBuffer) > 0 {
			joined := ""
			for _, t := range pathBuffer {
				joined += t.Value
			}
			if isLikelyPath(joined) {
				varCounters["path"]++
				varName := "path"
				if varCounters["path"] > 1 {
					varName = fmt.Sprintf("path%d", varCounters["path"])
				}
				suggestions = append(suggestions, struct{ VarName, Original string }{varName, joined})
			} else {
				// Not a path: suggest each Name token in the buffer
				for _, t := range pathBuffer {
					if t.Type == chroma.Name {
						if !checker.IsKnownCommand(t.Value) && !strings.HasPrefix(t.Value, "-") {
							varCounters["string"]++
							varName := "string"
							if varCounters["string"] > 1 {
								varName = fmt.Sprintf("string%d", varCounters["string"])
							}
							suggestions = append(suggestions, struct{ VarName, Original string }{varName, t.Value})
						}
					}
				}
			}
			pathBuffer = nil
		}
	}

	for i, token := range tokens {
		if token.Type == chroma.Name || token.Type == chroma.Punctuation {
			// Accumulate possible path
			pathBuffer = append(pathBuffer, token)
			// If this is the last token, flush the buffer
			if i == len(tokens)-1 {
				flushPathBuffer()
			}
			continue
		} else {
			flushPathBuffer()
		}

		if token.Type == chroma.LiteralString {
			varCounters["string"]++
			varName := "string"
			if varCounters["string"] > 1 {
				varName = fmt.Sprintf("string%d", varCounters["string"])
			}
			suggestions = append(suggestions, struct{ VarName, Original string }{varName, token.Value})
			continue
		}
		if token.Type == chroma.LiteralNumber {
			varCounters["number"]++
			varName := "number"
			if varCounters["number"] > 1 {
				varName = fmt.Sprintf("number%d", varCounters["number"])
			}
			suggestions = append(suggestions, struct{ VarName, Original string }{varName, token.Value})
			continue
		}
		if token.Type == chroma.NameVariable {
			varCounters["variable"]++
			varName := "variable"
			if varCounters["variable"] > 1 {
				varName = fmt.Sprintf("variable%d", varCounters["variable"])
			}
			suggestions = append(suggestions, struct{ VarName, Original string }{varName, token.Value})
			continue
		}
	}
	flushPathBuffer()

	fmt.Fprintln(out) // Blank line after spinner
	spinner.Stop()
	fmt.Fprintln(out) // Ensure a blank line before the output

	// Parameterization (no spinner needed)
	var paramStr = cmdStr
	var suggestionReplacements []struct{ Original, Replacement string }
	if len(suggestions) > 0 {
		for _, s := range suggestions {
			paramStr = strings.Replace(paramStr, s.Original, "{{"+s.VarName+"}}", 1)
			suggestionReplacements = append(suggestionReplacements, struct{ Original, Replacement string }{s.Original, "{{" + s.VarName + "}}"})
		}
	}

	// Print original command
	fmt.Fprintf(out, "\x1b[1;32mOriginal command (syntax highlighted):\x1b[0m\n")
	if err := quick.Highlight(out, cmdStr, "powershell", "terminal16m", "native"); err != nil {
		return fmt.Errorf("failed to highlight command: %w", err)
	}
	fmt.Fprintln(out)

	if len(suggestionReplacements) > 0 {
		fmt.Fprintln(out) // Blank line between commands
		fmt.Fprintf(out, "\x1b[1;32mParameterized version (syntax highlighted):\x1b[0m\n")
		if err := quick.Highlight(out, paramStr, "powershell", "terminal16m", "native"); err != nil {
			return fmt.Errorf("failed to highlight parameterized command: %w", err)
		}
		fmt.Fprintln(out)
	}

	if len(suggestions) > 0 {
		fmt.Fprintln(out) // Blank line before suggestions
		fmt.Fprintln(out, "Suggested variables:")
		for _, s := range suggestions {
			fmt.Fprintf(out, "  \x1b[1;33m%s\x1b[0m\x1b[1;37m ← was \x1b[0m\x1b[1;36m%s\x1b[0m\n", s.VarName, s.Original)
		}
	} else {
		fmt.Fprintf(out, "\n\x1b[1;33mNo suggestions found.\x1b[0m\n")
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
