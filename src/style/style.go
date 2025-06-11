package style

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// Define a thin border manually
var ThinBorder = lipgloss.Border{
	Top:         " ",
	Bottom:      " ",
	Left:        " ",
	Right:       "│",
	TopLeft:     " ",
	TopRight:    " ",
	BottomLeft:  " ",
	BottomRight: " ",
}

// Menu styles
var (
	MenuHeaderStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	MenuColumnStyle   = lipgloss.NewStyle().Bold(false).Padding(0, 1)
	MenuDimmedStyle   = lipgloss.NewStyle().Padding(0, 1).Faint(true).Bold(false)
	MenuActiveStyle   = lipgloss.NewStyle().Padding(0, 1).Bold(false)
	MenuSelectedStyle = lipgloss.NewStyle().Padding(0, 1).Bold(true).Background(lipgloss.Color("57")).Foreground(lipgloss.Color("229"))
	MenuNormalStyle   = lipgloss.NewStyle().Padding(0, 1).Foreground(lipgloss.Color("36"))
	MenuHelpStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Italic(true)

	HeaderStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	ItemStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("36"))
	SelectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("229")).Background(lipgloss.Color("57")).Bold(true)
	ErrorStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	SuccessStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Bold(true)
	ConfirmStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	PromptStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("36"))
	InfoStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("14")).Bold(true)
	VarNameStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	VarWasStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("7"))
	VarValueStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("14"))

	ThinBorderStyle = lipgloss.NewStyle().BorderRight(true).BorderStyle(ThinBorder).Padding(0, 1)
	CursorStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("229")).Background(lipgloss.Color("57"))

	CommandStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("63")).Bold(true)
	NameStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("33")).Bold(true)
	DescStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("36")).Italic(true)
	VarStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Underline(true)
)

// GadgetInfoPanel returns a styled info panel for a gadget (command, name, desc, variables)
func GadgetInfoPanel(command, name, desc string, variables map[string]string, varOrder []string) string {
	header := HeaderStyle.Render("Gadget Info")
	cmd := CommandStyle.Render("PS> " + command)
	nameLine := NameStyle.Render("🔖 Name: ") + name
	descLine := DescStyle.Render("💡 Description: ") + desc
	varLines := ""
	for _, v := range varOrder {
		val := variables[v]
		varLines += VarStyle.Render(fmt.Sprintf("✏️  %s: ", v)) + val + "\n"
	}
	return header + "\n\n" + cmd + "\n" + nameLine + "\n" + descLine + "\n" + varLines
}
