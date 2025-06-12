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
	refreshCmdsFlag bool
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

// Analyze analyzes a PowerShell command, highlights it, and suggests likely user-input variables.
func Analyze(command ...string) error {
	if refreshCmdsFlag {
		// refreshPowerShellCommandChecker() // No-op: command checker is not used
	}
	out := colorable.NewColorableStdout()
	var cmdStr string
	if len(command) > 0 && strings.TrimSpace(strings.Join(command, " ")) != "" {
		cmdStr = strings.Join(command, " ")
	} else {
		fmt.Fprintln(out)
		fmt.Fprint(out, "Enter a PowerShell command to analyze (or use ")
		colorText.Yellow("--command")
		fmt.Fprintln(out, " for automation):")
		colorText.Bold("> ")
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
		colorText.Red("No syntax highlighter found for PowerShell. Aborting.")
		return fmt.Errorf("no lexer for PowerShell")
	}
	iterator, err := lexer.Tokenise(nil, cmdStr)
	if err != nil {
		spinner.Stop()
		colorText.Red("Failed to tokenize command: " + err.Error())
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
				suggestions = append(suggestions, struct{ VarName, Original string }{VarName: varName, Original: joined})
			}
			pathBuffer = nil
		}
	}
	for i, token := range tokens {
		if token.Type == chroma.String || token.Type == chroma.LiteralString || token.Type == chroma.LiteralStringDouble || token.Type == chroma.LiteralStringSingle {
			varCounters["string"]++
			varName := fmt.Sprintf("str%d", varCounters["string"])
			suggestions = append(suggestions, struct{ VarName, Original string }{VarName: varName, Original: token.Value})
			continue
		}
		if token.Type == chroma.LiteralNumber {
			varCounters["number"]++
			varName := fmt.Sprintf("num%d", varCounters["number"])
			suggestions = append(suggestions, struct{ VarName, Original string }{VarName: varName, Original: token.Value})
			continue
		}
		if token.Type == chroma.NameVariable {
			varCounters["variable"]++
			varName := fmt.Sprintf("var%d", varCounters["variable"])
			suggestions = append(suggestions, struct{ VarName, Original string }{VarName: varName, Original: token.Value})
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
			param := "{{" + s.VarName + "}}"
			paramStr = strings.ReplaceAll(paramStr, s.Original, param)
			suggestionReplacements = append(suggestionReplacements, struct{ Original, Replacement string }{Original: s.Original, Replacement: param})
		}
	}

	colorText.Green("\u2714 Analysis complete!")
	fmt.Fprint(out, " ")
	colorText.Bold("Original command:")
	fmt.Fprintln(out, cmdStr)
	if len(suggestionReplacements) > 0 {
		fmt.Fprintln(out)
	}
	if len(suggestions) > 0 {
		colorText.Cyan("Suggested parameterization:")
		fmt.Fprintln(out, paramStr)
		fmt.Fprintln(out)
		colorText.Yellow("Variables detected:")
		for _, s := range suggestions {
			fmt.Fprint(out, "  ")
			colorText.Magenta("%s", s.VarName)
			fmt.Fprint(out, " (from ")
			colorText.Cyan(s.Original)
			fmt.Fprintln(out, ")")
		}
	} else {
		colorText.Yellow("No variables or parameters detected. This command is ready to use as-is!")
	}

	fmt.Fprintln(out)
	colorText.Cyan("To save this as a gadget, use:")
	fmt.Fprint(out, "  ")
	colorText.Bold("GoGoGadget add --command '<your command>' --scriptname <name> --desc <description>")
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
