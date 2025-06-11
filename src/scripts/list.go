package scripts

import (
	"fmt"
	"gogo/src/style"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

type listGadgetsModel struct {
	gadgets  []string
	cursor   int
	quitting bool
}

func InitialListGadgetsModel() tea.Model {
	scripts, _ := LoadScripts()
	var gadgets []string
	for name := range scripts {
		gadgets = append(gadgets, name)
	}
	return listGadgetsModel{gadgets: gadgets}
}

func (m listGadgetsModel) Init() tea.Cmd { return nil }

func (m listGadgetsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		}
	}
	return m, nil
}

func (m listGadgetsModel) View() string {
	header := style.HeaderStyle.Render("Your GoGoGadget gadgets")
	help := style.MenuHelpStyle.Render("↑/↓: Navigate  Enter: Select  Esc: Back")

	var s string
	s += header + "\n\n"
	if len(m.gadgets) == 0 {
		s += style.InfoStyle.Render("No gadgets found. Press Esc to return.")
		s += "\n\n" + help
		return s
	}
	for i, g := range m.gadgets {
		line := style.MenuActiveStyle.Render(g)
		if i == m.cursor {
			line = style.MenuSelectedStyle.Render(line)
		}
		s += line + "\n"
	}
	s += "\n" + help
	return s
}

// InitialListGadgetsCommand returns the cobra.Command for listing gadgets
func InitialListGadgetsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "gadgets",
		Short: "List all gadgets",
		Run: func(cmd *cobra.Command, args []string) {
			scripts, _ := LoadScripts()
			if len(scripts) == 0 {
				fmt.Println(style.InfoStyle.Render("No gadgets found. Use 'gogogadget add' to create one."))
				return
			}
			fmt.Println(style.HeaderStyle.Render("Your gadgets:"))
			for name, config := range scripts {
				fmt.Println(style.SuccessStyle.Render("  "+name+": ") + style.DescStyle.Render(config.Description))
			}
		},
	}
}
