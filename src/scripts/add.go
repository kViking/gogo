package scripts

import (
	"fmt"
	"regexp"
	"strings"

	"gogo/src/style"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

// This file provides the add gadget TUI model and the CLI command for adding gadgets.
// It does NOT import or reference the menu package, and is UI-agnostic.
//
// - If called from the CLI (via InitialAddGadgetCommand), the program returns to CLI after add completes.
// - If called from the TUI menu, the menu system is responsible for returning to the menu after add completes.
//
// Do not add any menu-specific logic or imports here.

type addGadgetModel struct {
	step      int
	steps     []string
	command   string
	name      string
	desc      string
	variables map[string]string
	varOrder  []string
	varIndex  int
	input     string
	errorMsg  string
	success   bool
	quitting  bool
}

// Custom message to signal add gadget completion
// This can be used to return control to the main menu
type addGadgetDoneMsg struct{}

func initialAddGadgetModel() addGadgetModel {
	return addGadgetModel{
		step:      0,
		steps:     []string{"command", "name", "desc", "variables", "confirm"},
		variables: map[string]string{},
		varOrder:  []string{},
	}
}

// Exported for use in main.go
func InitialAddGadgetModel() tea.Model {
	return initialAddGadgetModel()
}

// InitialAddGadgetCommand returns the cobra.Command for adding a gadget
func InitialAddGadgetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "add",
		Short: "Add a new gadget",
		Run: func(cmd *cobra.Command, args []string) {
			model := initialAddGadgetModel()
			p := tea.NewProgram(model)
			if _, err := p.Run(); err != nil {
				fmt.Println("Error running add gadget TUI:", err)
			} else {
				// After add gadget TUI completes, return to main menu
				fmt.Println("Returning to main menu...")
			}
		},
	}
}

func (m addGadgetModel) Init() tea.Cmd {
	return nil
}

func (m addGadgetModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Only treat Esc/q as quit if not in an input/entry state (step 0-4 are entry states)
		if (msg.String() == "esc" || msg.String() == "q") && (m.step < 0 || m.step > 4) {
			m.quitting = true
			return m, tea.Quit
		}
		switch msg.String() {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "enter":
			return m.handleEnter()
		case "backspace":
			if len(m.input) > 0 {
				m.input = m.input[:len(m.input)-1]
			}
		case "tab":
			// ignore for now
		default:
			if len(msg.String()) == 1 {
				m.input += msg.String()
			}
		}
	}
	if m.success {
		// Instead of quitting, send a custom message to parent
		return m, func() tea.Msg { return addGadgetDoneMsg{} }
	}
	if m.quitting {
		return m, tea.Quit
	}
	return m, nil
}

func (m addGadgetModel) handleEnter() (tea.Model, tea.Cmd) {
	m.errorMsg = ""
	if m.step == 0 { // command
		m.command = m.input
		m.input = ""
		m.step++
		return m, nil
	}
	if m.step == 1 { // name
		nameRe := regexp.MustCompile(`^[A-Za-z0-9_\-]+$`)
		if !nameRe.MatchString(m.input) {
			m.errorMsg = "Gadget names cannot contain spaces or punctuation. Use only letters, numbers, dashes, or underscores."
			return m, nil
		}
		m.name = m.input
		m.input = ""
		m.step++
		return m, nil
	}
	if m.step == 2 { // desc
		m.desc = m.input
		m.input = ""
		// Extract variables from command
		m.varOrder = ExtractVariables(m.command)
		if len(m.varOrder) > 0 {
			m.step++
			m.varIndex = 0
			return m, nil
		}
		m.step = 4 // skip to confirm
		return m, nil
	}
	if m.step == 3 { // variables
		if m.varIndex < len(m.varOrder) {
			m.variables[m.varOrder[m.varIndex]] = m.input
			m.input = ""
			m.varIndex++
			if m.varIndex >= len(m.varOrder) {
				m.step++
			}
			return m, nil
		}
	}
	if m.step == 4 { // confirm
		// Save gadget
		if m.name == "" || m.command == "" {
			m.errorMsg = "Gadget name and command are required."
			return m, nil
		}
		scriptsMap, _ := LoadScripts()
		scriptsMap[m.name] = ScriptConfig{
			Description: m.desc,
			Command:     m.command,
			Variables:   m.variables,
		}
		if err := SaveScripts(scriptsMap); err != nil {
			m.errorMsg = "Error saving gadget: " + err.Error()
			return m, nil
		}
		m.success = true
		// Do not set m.quitting here; let Update handle completion
		return m, nil
	}
	return m, nil
}

func (m addGadgetModel) View() string {
	header := style.HeaderStyle.Render(fmt.Sprintf("Add a new GoGoGadget gadget  •  Step %d/%d", m.step+1, len(m.steps)))
	errorStyle := style.ErrorStyle
	successStyle := style.SuccessStyle
	cursor := style.CursorStyle.Render("_")
	help := style.MenuHelpStyle.Render("Enter: Confirm  Esc: Cancel  ←/→: Move")

	var sb strings.Builder
	sb.WriteString(header + "\n\n")
	sb.WriteString(style.GadgetInfoPanel(m.command, m.name, m.desc, m.variables, m.varOrder))

	if m.errorMsg != "" {
		sb.WriteString("\n" + errorStyle.Render(m.errorMsg) + "\n")
	}
	if m.success {
		sb.WriteString("\n" + successStyle.Render("✅ Gadget added!") + "\n\n")
		sb.WriteString(help + "\n")
		return sb.String()
	}

	sb.WriteString("\n")
	switch m.step {
	case 0:
		sb.WriteString(style.InfoStyle.Render("📝 Enter the PowerShell command this gadget will run (use {{variable}} for variables): "))
		sb.WriteString(style.MenuActiveStyle.Render(m.input) + cursor)
	case 1:
		sb.WriteString(style.InfoStyle.Render("🔖 Enter gadget name: "))
		sb.WriteString(style.MenuActiveStyle.Render(m.input) + cursor)
	case 2:
		sb.WriteString(style.InfoStyle.Render("💡 Enter gadget description: "))
		sb.WriteString(style.MenuActiveStyle.Render(m.input) + cursor)
	case 3:
		if m.varIndex < len(m.varOrder) {
			v := m.varOrder[m.varIndex]
			sb.WriteString(style.InfoStyle.Render(fmt.Sprintf("✏️  Describe variable '%s': ", v)))
			sb.WriteString(style.MenuActiveStyle.Render(m.input) + cursor)
		}
	case 4:
		sb.WriteString(style.ConfirmStyle.Render("Press Enter to confirm and save this gadget, or Esc to cancel."))
	}

	sb.WriteString("\n\n" + help)
	return sb.String()
}
