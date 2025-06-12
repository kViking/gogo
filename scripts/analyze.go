// analyze.go: Analyze PowerShell commands for variable suggestions and parameterization.
package scripts

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/lexers"
	"github.com/mattn/go-colorable"
	"github.com/spf13/cobra"
)

var (
	psCmdChecker     *PowerShellCommandChecker
	psCmdCheckerInit bool
	refreshCmdsFlag  bool
)

// PowerShellCommandChecker caches the list of known PowerShell commands for efficient lookup
// and provides a method to check if a string is a known command.
type PowerShellCommandChecker struct {
	known map[string]struct{}
}

// NewPowerShellCommandChecker dynamically loads all available PowerShell command names (lowercased)
func NewPowerShellCommandChecker() *PowerShellCommandChecker {
	cmd := "pwsh -NoProfile -Command Get-Command | Select-Object -ExpandProperty Name"
	output, err := exec.Command("bash", "-c", cmd).Output()
	m := make(map[string]struct{})
	if err == nil {
		for _, line := range strings.Split(string(output), "\n") {
			line = strings.TrimSpace(strings.ToLower(line))
			if line != "" {
				m[line] = struct{}{}
			}
		}
	}
	return &PowerShellCommandChecker{known: m}
}

func getPowerShellCommandChecker() *PowerShellCommandChecker {
	if !psCmdCheckerInit {
		psCmdChecker = NewPowerShellCommandChecker()
		psCmdCheckerInit = true
	}
	return psCmdChecker
}

func refreshPowerShellCommandChecker() {
	psCmdChecker = NewPowerShellCommandChecker()
	psCmdCheckerInit = true
}

func (p *PowerShellCommandChecker) IsKnown(cmd string) bool {
	cmd = strings.ToLower(cmd)
	_, ok := p.known[cmd]
	return ok
}

// Analyze prompts for a PowerShell command, highlights it, and suggests likely user-input variables.
func Analyze(command ...string) error {
	if refreshCmdsFlag {
		refreshPowerShellCommandChecker()
	}
	out := colorable.NewColorableStdout()
	var cmdStr string
	if len(command) > 0 && strings.TrimSpace(strings.Join(command, " ")) != "" {
		cmdStr = strings.Join(command, " ")
	} else {
		fmt.Fprintln(out)
		fmt.Fprint(out, "Enter a PowerShell command to analyze: ")
		reader := bufio.NewReader(os.Stdin)
		c, _ := reader.ReadString('\n')
		cmdStr = strings.TrimSpace(c)
	}

	fmt.Fprintln(out)
	spinner := NewSpinner("Analyzing command...")
	spinner.Start()

	lexer := lexers.Get("powershell")
	if lexer == nil {
		spinner.Stop()
		return fmt.Errorf("no lexer for PowerShell")
	}
	iterator, err := lexer.Tokenise(nil, cmdStr)
	if err != nil {
		spinner.Stop()
		return err
	}

	tokens := []chroma.Token{}
	for token := iterator(); token.Type != chroma.EOF.Type; token = iterator() {
		tokens = append(tokens, token)
	}

	var suggestions []struct{ VarName, Original string }
	varCounters := map[string]int{"string": 0, "number": 0, "variable": 0, "path": 0}
	var pathBuffer []chroma.Token
	flushPathBuffer := func() {
		if len(pathBuffer) > 0 {
			joined := ""
			for _, t := range pathBuffer {
				joined += t.Value
			}
			if isLikelyPath(joined) {
				varCounters["path"]++
				varName := fmt.Sprintf("path%d", varCounters["path"])
				suggestions = append(suggestions, struct{ VarName, Original string }{varName, joined})
			}
			pathBuffer = nil
		}
	}
	for i, token := range tokens {
		if token.Type == chroma.String || token.Type == chroma.LiteralString {
			varCounters["string"]++
			varName := fmt.Sprintf("str%d", varCounters["string"])
			suggestions = append(suggestions, struct{ VarName, Original string }{varName, token.Value})
			continue
		}
		if token.Type == chroma.LiteralNumber {
			varCounters["number"]++
			varName := fmt.Sprintf("num%d", varCounters["number"])
			suggestions = append(suggestions, struct{ VarName, Original string }{varName, token.Value})
			continue
		}
		if token.Type == chroma.NameVariable {
			varCounters["variable"]++
			varName := fmt.Sprintf("var%d", varCounters["variable"])
			suggestions = append(suggestions, struct{ VarName, Original string }{varName, token.Value})
			continue
		}
		if token.Type == chroma.LiteralString || token.Type == chroma.LiteralStringDouble || token.Type == chroma.LiteralStringSingle {
			pathBuffer = append(pathBuffer, token)
			if i == len(tokens)-1 {
				flushPathBuffer()
			}
			continue
		}
		flushPathBuffer()
	}
	flushPathBuffer()

	fmt.Fprintln(out)
	spinner.Stop()
	fmt.Fprintln(out)

	paramStr := cmdStr
	var suggestionReplacements []struct{ Original, Replacement string }
	if len(suggestions) > 0 {
		for _, s := range suggestions {
			paramStr = strings.ReplaceAll(paramStr, s.Original, fmt.Sprintf("{{%s}}", s.VarName))
			suggestionReplacements = append(suggestionReplacements, struct{ Original, Replacement string }{s.Original, fmt.Sprintf("{{%s}}", s.VarName)})
		}
	}

	fmt.Fprintf(out, "\x1b[1;32mOriginal command:\x1b[0m\n")
	fmt.Fprintln(out, cmdStr)
	if len(suggestionReplacements) > 0 {
		fmt.Fprintln(out)
		fmt.Fprintf(out, "\x1b[1;36mSuggested parameterization:\x1b[0m\n")
		fmt.Fprintln(out, paramStr)
	}
	if len(suggestions) > 0 {
		fmt.Fprintln(out)
		fmt.Fprintf(out, "\x1b[1;33mSuggested variables:\x1b[0m\n")
		for _, s := range suggestions {
			fmt.Fprintf(out, "  %s: %s\n", s.VarName, s.Original)
		}
	} else {
		fmt.Fprintln(out, "\x1b[33mNo variables suggested.\x1b[0m")
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Fprint(out, "\nWould you like to save this parameterization as a gadget? Y/N: ")
	resp, _ := reader.ReadString('\n')
	resp = strings.TrimSpace(strings.ToLower(resp))
	if resp == "y" || resp == "yes" {
		// TODO: Call the add process (reuse NewAddCommand logic)
		// Simulate: GoGoGadget add --command <paramStr>
	}
	return nil
}

// isLikelyPath detects likely Windows/Unix paths
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
		Use:   "analyze",
		Short: "Analyze a PowerShell command for variables and parameterization",
		RunE: func(cmd *cobra.Command, args []string) error {
			refreshCmdsFlag = refreshCommands
			if command != "" {
				return Analyze(command)
			}
			return Analyze()
		},
	}
	cmd.Flags().StringVar(&command, "command", "", "PowerShell command to analyze")
	cmd.Flags().BoolVar(&refreshCommands, "refresh", false, "Refresh PowerShell command cache")
	return cmd
}
