package scripts

import (
	"fmt"
	"gogo/src/style"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

type editState int

const (
	editMenu editState = iota
	editField
	editVarSubmenu
)

type editGadgetModel struct {
	gadgetName string
	command    string
	desc       string
	variables  map[string]string // variable name -> description
	varOrder   []string

	// TUI state
	state      editState
	menuIndex  int
	varIndex   int
	varSubmenu int // 0: edit name, 1: edit description
	input      string
	quitting   bool
}

func InitialEditGadgetModel(name string, config ScriptConfig) editGadgetModel {
	varOrder := ExtractVariables(config.Command)
	return editGadgetModel{
		gadgetName: name,
		command:    config.Command,
		desc:       config.Description,
		variables:  config.Variables,
		varOrder:   varOrder,
		state:      editMenu,
	}
}

func (m editGadgetModel) Init() tea.Cmd { return nil }

func (m editGadgetModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Only treat Esc/q as quit if not in an input/entry state (editMenu and editVarSubmenu are navigation states)
		if (msg.String() == "esc" || msg.String() == "q") && (m.state == editMenu || m.state == editVarSubmenu) {
			m.quitting = true
			return m, tea.Quit
		}
		switch m.state {
		case editMenu:
			switch msg.String() {
			case "up":
				if m.menuIndex > 0 {
					m.menuIndex--
				}
			case "down":
				max := 3 + len(m.varOrder) // name, command, desc, then each variable
				if m.menuIndex < max-1 {
					m.menuIndex++
				}
			case "enter":
				if m.menuIndex < 3 {
					m.state = editField
					m.input = ""
				} else {
					m.state = editVarSubmenu
					m.varIndex = m.menuIndex - 3
					m.varSubmenu = 0
				}
			}
		case editField:
			switch msg.String() {
			case "enter":
				if m.menuIndex == 0 && m.input != "" {
					m.gadgetName = m.input
					m.state = editMenu
				} else if m.menuIndex == 1 && m.input != "" {
					m.command = m.input
					m.varOrder = ExtractVariables(m.command)
					m.state = editMenu
				} else if m.menuIndex == 2 && m.input != "" {
					m.desc = m.input
					m.state = editMenu
				} else if m.menuIndex >= 3 {
					if m.varSubmenu == 0 && m.input != "" {
						// Edit variable name
						oldName := m.varOrder[m.varIndex]
						newName := m.input
						if newName != "" && newName != oldName {
							// Update varOrder
							m.varOrder[m.varIndex] = newName
							// Update variables map
							m.variables[newName] = m.variables[oldName]
							delete(m.variables, oldName)
							m.state = editVarSubmenu
						}
					} else if m.varSubmenu == 1 && m.input != "" {
						// Edit variable description
						v := m.varOrder[m.varIndex]
						m.variables[v] = m.input
						m.state = editVarSubmenu // Return to variable submenu after editing
					}
				} else {
					m.state = editMenu
				}
				m.input = ""
			case "esc":
				m.state = editMenu
				m.input = ""
			default:
				if len(msg.String()) == 1 {
					m.input += msg.String()
				} else if msg.String() == "backspace" && len(m.input) > 0 {
					m.input = m.input[:len(m.input)-1]
				}
			}
		case editVarSubmenu:
			switch msg.String() {
			case "up", "down":
				m.varSubmenu = 1 - m.varSubmenu // toggle between name/desc
			case "enter":
				m.state = editField
				m.input = ""
			}
		}
	}
	return m, nil
}

func (m editGadgetModel) View() string {
	help := style.MenuHelpStyle.Render("↑/↓: Move  Enter: Edit  Esc: Back/Cancel")
	cursor := style.CursorStyle.Render("_")

	var sb strings.Builder
	sb.WriteString(style.GadgetInfoPanel(m.command, m.gadgetName, m.desc, m.variables, m.varOrder))
	sb.WriteString("\n")

	switch m.state {
	case editMenu:
		menu := []string{
			style.NameStyle.Render("Edit name"),
			style.CommandStyle.Render("Edit command"),
			style.DescStyle.Render("Edit description"),
		}
		for _, v := range m.varOrder {
			menu = append(menu, style.VarStyle.Render(fmt.Sprintf("Edit variable '%s'", v)))
		}
		for i, item := range menu {
			if i == m.menuIndex {
				sb.WriteString(style.MenuSelectedStyle.Render("> "+item) + "\n")
			} else {
				sb.WriteString("  " + item + "\n")
			}
		}
		sb.WriteString("\n" + help)
	case editField:
		prompt := ""
		if m.menuIndex == 0 {
			prompt = style.NameStyle.Render("Edit name: ")
		} else if m.menuIndex == 1 {
			prompt = style.CommandStyle.Render("Edit command: ")
		} else if m.menuIndex == 2 {
			prompt = style.DescStyle.Render("Edit description: ")
		} else {
			if m.varSubmenu == 0 {
				prompt = style.VarStyle.Render(fmt.Sprintf("Edit variable name for '%s': ", m.varOrder[m.varIndex]))
			} else {
				prompt = style.VarStyle.Render(fmt.Sprintf("Edit description for variable '%s': ", m.varOrder[m.varIndex]))
			}
		}
		sb.WriteString(prompt + m.input + cursor + "\n\n" + help)
	case editVarSubmenu:
		v := m.varOrder[m.varIndex]
		items := []string{
			style.VarStyle.Render(fmt.Sprintf("Edit variable name for '%s'", v)),
			style.VarStyle.Render(fmt.Sprintf("Edit description for variable '%s'", v)),
		}
		for i, item := range items {
			if i == m.varSubmenu {
				sb.WriteString(style.MenuSelectedStyle.Render("> "+item) + "\n")
			} else {
				sb.WriteString("  " + item + "\n")
			}
		}
		sb.WriteString("\n" + help)
	}
	return sb.String()
}

// InitialEditGadgetCommand returns the cobra.Command for editing a gadget
func InitialEditGadgetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "edit [gadget]",
		Short: "Edit a gadget (opens TUI)",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]
			scripts, _ := LoadScripts()
			config, ok := scripts[name]
			if !ok {
				errorText("Gadget not found: " + name)
				return
			}
			model := InitialEditGadgetModel(name, config)
			p := tea.NewProgram(model)
			if _, err := p.Run(); err != nil {
				fmt.Println("Error running edit gadget TUI:", err)
			}
		},
	}
}
