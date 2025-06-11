package analyze

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/quick"
)

type DoneMsg struct {
	OrigHL      string
	ParamHL     string
	Suggestions []Suggestion
	ParamStr    string
	Error       error
}

func RunAnalysis(cmdStr string, checker func(string) bool) DoneMsg {
	lexer := lexers.Get("powershell")
	if lexer == nil {
		return DoneMsg{Error: fmt.Errorf("could not get PowerShell lexer")}
	}
	iterator, err := lexer.Tokenise(nil, cmdStr)
	if err != nil {
		return DoneMsg{Error: fmt.Errorf("failed to tokenize command: %w", err)}
	}

	tokens := []chroma.Token{}
	for token := iterator(); token.Type != chroma.EOF.Type; token = iterator() {
		tokens = append(tokens, token)
	}

	var suggestions []Suggestion
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
				varName := "path"
				if varCounters["path"] > 1 {
					varName = fmt.Sprintf("path%d", varCounters["path"])
				}
				suggestions = append(suggestions, Suggestion{varName, joined})
			} else {
				for _, t := range pathBuffer {
					if t.Type == chroma.Name {
						if !checker(t.Value) && !strings.HasPrefix(t.Value, "-") {
							varCounters["string"]++
							varName := "string"
							if varCounters["string"] > 1 {
								varName = fmt.Sprintf("string%d", varCounters["string"])
							}
							suggestions = append(suggestions, Suggestion{varName, t.Value})
						}
					}
				}
			}
			pathBuffer = nil
		}
	}

	for i, token := range tokens {
		if token.Type == chroma.Name || token.Type == chroma.Punctuation {
			pathBuffer = append(pathBuffer, token)
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
			suggestions = append(suggestions, Suggestion{varName, token.Value})
			continue
		}
		if token.Type == chroma.LiteralNumber {
			varCounters["number"]++
			varName := "number"
			if varCounters["number"] > 1 {
				varName = fmt.Sprintf("number%d", varCounters["number"])
			}
			suggestions = append(suggestions, Suggestion{varName, token.Value})
			continue
		}
		if token.Type == chroma.NameVariable {
			varCounters["variable"]++
			varName := "variable"
			if varCounters["variable"] > 1 {
				varName = fmt.Sprintf("variable%d", varCounters["variable"])
			}
			suggestions = append(suggestions, Suggestion{varName, token.Value})
			continue
		}
	}
	flushPathBuffer()

	paramStr := cmdStr
	if len(suggestions) > 0 {
		for _, s := range suggestions {
			paramStr = strings.ReplaceAll(paramStr, s.Original, "{{"+s.VarName+"}}")
		}
	}

	var buf bytes.Buffer
	_ = quick.Highlight(&buf, cmdStr, "powershell", "terminal16m", "native")
	origHL := buf.String()
	paramHL := ""
	if paramStr != cmdStr {
		buf.Reset()
		_ = quick.Highlight(&buf, paramStr, "powershell", "terminal16m", "native")
		paramHL = buf.String()
	}

	return DoneMsg{
		OrigHL:      origHL,
		ParamHL:     paramHL,
		Suggestions: suggestions,
		ParamStr:    paramStr,
		Error:       nil,
	}
}

// isLikelyPath is kept here for logic.go, but you may want to move it to a util file if reused
func isLikelyPath(s string) bool {
	if len(s) < 3 {
		return false
	}
	if strings.Contains(s, ":\\") || strings.Contains(s, "/") {
		return true
	}
	return false
}
