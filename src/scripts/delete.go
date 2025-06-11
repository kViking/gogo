package scripts

import (
	"fmt"
	"strings"

	"gogo/src/style"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

type deleteGadgetModel struct {
	gadgets    []string
	cursor     int
	confirm    bool
	quitting   bool
	errorMsg   string
	success    bool
	gadgetName string
}

func InitialDeleteGadgetModel() tea.Model {
	scriptsMap, _ := LoadScripts()
	var gadgets []string
	for name := range scriptsMap {
		gadgets = append(gadgets, name)
	}
	return deleteGadgetModel{gadgets: gadgets}
}

func (m deleteGadgetModel) Init() tea.Cmd { return nil }

func (m deleteGadgetModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.quitting = true
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.gadgets)-1 {
				m.cursor++
			}
		case "enter":
			if len(m.gadgets) == 0 {
				return m, nil
			}
			if !m.confirm {
				m.confirm = true
				m.gadgetName = m.gadgets[m.cursor]
				return m, nil
			}
			scriptsMap, _ := LoadScripts()
			name := m.gadgets[m.cursor]
			delete(scriptsMap, name)
			SaveScripts(scriptsMap)
			m.success = true
			m.gadgets = append(m.gadgets[:m.cursor], m.gadgets[m.cursor+1:]...)
			if m.cursor > 0 && m.cursor >= len(m.gadgets) {
				m.cursor--
			}
			return m, nil
		case "y":
			if m.confirm {
				scriptsMap, _ := LoadScripts()
				name := m.gadgets[m.cursor]
				delete(scriptsMap, name)
				SaveScripts(scriptsMap)
				m.success = true
				m.gadgets = append(m.gadgets[:m.cursor], m.gadgets[m.cursor+1:]...)
				if m.cursor > 0 && m.cursor >= len(m.gadgets) {
					m.cursor--
				}
				return m, nil
			}
		case "n":
			if m.confirm {
				m.confirm = false
				m.gadgetName = ""
				return m, nil
			}
		}
	}
	return m, nil
}

func (m deleteGadgetModel) View() string {
	header := style.HeaderStyle.Render("Delete a GoGoGadget gadget")
	errorStyle := style.ErrorStyle
	successStyle := style.SuccessStyle
	help := style.MenuHelpStyle.Render("Enter: Confirm  Esc: Cancel  ←/→: Move")

	var s string
	s += header + "\n\n"
	if m.errorMsg != "" {
		s += errorStyle.Render(m.errorMsg) + "\n\n"
	}
	if m.success {
		s += successStyle.Render("🗑️  Gadget deleted!") + "\n\n"
		s += help + "\n"
		return s
	}
	s += style.InfoStyle.Render(fmt.Sprintf("Are you sure you want to delete gadget '%s'? (Enter to confirm, Esc to cancel)", m.gadgetName))
	s += "\n\n" + help
	return s
}

// InitialDeleteGadgetCommand returns the cobra.Command for deleting a gadget
func InitialDeleteGadgetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "delete [gadget]",
		Short: "Delete a gadget",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]
			scripts, _ := LoadScripts()
			if _, ok := scripts[name]; !ok {
				fmt.Println(style.ErrorStyle.Render("Gadget not found: " + name))
				return
			}
			prompt := style.PromptStyle.Render(fmt.Sprintf("Are you sure you want to delete '%s'? (y/N): ", name))
			fmt.Print(prompt)
			var confirm string
			fmt.Scanln(&confirm)
			confirm = strings.ToLower(strings.TrimSpace(confirm))
			if confirm == "y" || confirm == "yes" {
				delete(scripts, name)
				SaveScripts(scripts)
				fmt.Println(style.SuccessStyle.Render("Deleted gadget: " + name))
			} else {
				fmt.Println(style.InfoStyle.Render("Delete cancelled."))
			}
		},
	}
}
