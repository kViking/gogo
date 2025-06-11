package scripts

import (
	"regexp"
	"strings"
)

// ExtractVariables parses a command string and returns a slice of variable names used as {{var}}.
func ExtractVariables(command string) []string {
	var vars []string
	seen := map[string]bool{}
	re := regexp.MustCompile(`\{\{([A-Za-z0-9_]+)\}\}`)
	matches := re.FindAllStringSubmatch(command, -1)
	for _, m := range matches {
		if !seen[m[1]] {
			vars = append(vars, m[1])
			seen[m[1]] = true
		}
	}
	return vars
}

// GetSimpleVariableDescription returns a human-friendly description for a variable name (standalone, no config).
func GetSimpleVariableDescription(varName string) string {
	descriptions := map[string]string{
		"USER": "The current system user",
		"HOME": "The user's home directory",
		"PWD":  "The present working directory",
		// Add more known variable descriptions here
	}
	if desc, ok := descriptions[varName]; ok {
		return desc
	}
	return "No description available."
}

// RenameVariable renames all instances of a variable in a command string.
func RenameVariable(command, oldName, newName string) string {
	re := regexp.MustCompile(`\{\{` + regexp.QuoteMeta(oldName) + `\}\}`)
	return re.ReplaceAllString(command, "{{"+newName+"}}")
}

// ValidateVariableName checks if a variable name is valid (alphanumeric/underscore, not empty).
func ValidateVariableName(varName string) bool {
	if varName == "" {
		return false
	}
	re := regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)
	return re.MatchString(varName)
}

// FormatVariable returns the variable in {{var}} format.
func FormatVariable(varName string) string {
	return "{{" + varName + "}}"
}

// ReplaceVariables replaces all variables in a command string with their values from the map.
func ReplaceVariables(command string, values map[string]string) string {
	re := regexp.MustCompile(`\{\{([A-Za-z0-9_]+)\}\}`)
	return re.ReplaceAllStringFunc(command, func(match string) string {
		m := re.FindStringSubmatch(match)
		if len(m) == 2 {
			if val, ok := values[m[1]]; ok {
				return val
			}
		}
		return match
	})
}

// ListVariableDescriptions returns a formatted string listing variables and their descriptions.
func ListVariableDescriptions(vars []string) string {
	var b strings.Builder
	for _, v := range vars {
		desc := GetSimpleVariableDescription(v)
		b.WriteString("  - ")
		b.WriteString(v)
		b.WriteString(": ")
		b.WriteString(desc)
		b.WriteString("\n")
	}
	return b.String()
}
