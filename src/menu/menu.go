package menu

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"gogo/src/scripts"
	"gogo/src/style"
)

type MultiColumnMenuModel struct {
	Columns      []*MenuColumn
	ActiveColumn int
	Quitting     bool
	Viewport     tea.Model // For endpoint interactions (add/analyze)
}

func (m *MultiColumnMenuModel) Init() tea.Cmd {
	if m.Viewport != nil {
		return m.Viewport.Init()
	}
	return nil
}

func (m *MultiColumnMenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.Viewport != nil {
		// Always let the viewport handle the key, including esc/q
		m.Viewport, _ = m.Viewport.Update(msg)
		// If viewport is closed (set to nil), check if quitting was confirmed
		if m.Viewport == nil && m.Quitting {
			return m, tea.Quit
		}
		// Otherwise, just return to menu
		return m, nil
	}

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if keyMsg.String() == "ctrl+c" {
			m.Quitting = true
			return m, tea.Quit
		}
		if keyMsg.String() == "esc" || keyMsg.String() == "q" || keyMsg.String() == "ctrl+[" {
			if m.ActiveColumn > 0 {
				m.ActiveColumn--
				m.Columns = m.Columns[:m.ActiveColumn+1]
				return m, nil
			}
			// If at root, confirm quit
			m.Viewport = &ConfirmPromptModel{
				Message: MenuQuitConfirm(),
				OnConfirm: func() {
					m.Quitting = true
				},
				ConfirmedMsg: "Goodbye!",
			}
			return m, nil
		}
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "left":
			if m.ActiveColumn > 0 {
				m.ActiveColumn--
				m.Columns = m.Columns[:m.ActiveColumn+1]
			}
		case "right", "enter", " ":
			col := m.Columns[m.ActiveColumn]
			if col.GetNext != nil {
				next := col.GetNext(col.Options[col.Selected])
				if next != nil {
					m.Columns = append(m.Columns[:m.ActiveColumn+1], next)
					m.ActiveColumn++
					return m, nil
				}
			}
			// Handle endpoint actions
			switch col.Options[col.Selected] {
			case "Add a new gadget":
				// Launch the add gadget TUI, then return to the menu after completion
				m.Viewport = InitialAddGadgetModelWithDone(func() {
					// After add gadget completes, re-initialize the menu
					root := BuildRootMenu()
					*m = *InitialMenuModel(root)
				})
				return m, m.Viewport.Init()
			case "Analyze a command":
				// Call AnalyzeRunAnalysis(cmdStr) when needed
				return m, nil
			case "Quit":
				m.Quitting = true
				return m, tea.Quit
				// Add more endpoint actions as needed
				// Handle gadget subcommands
			}
			if m.ActiveColumn == 2 && len(m.Columns) > 2 {
				gadgetName := m.Columns[1].Options[m.Columns[1].Selected]
				subcmd := col.Options[col.Selected]
				switch subcmd {
				case "Run":
					// If the gadget takes variables, prompt in TUI, then run
					scriptsMap, _ := scripts.LoadScripts()
					config, ok := scriptsMap[gadgetName]
					if !ok {
						m.Viewport = nil
						return m, nil
					}
					varNames := scripts.ExtractVariables(config.Command)
					if len(varNames) == 0 {
						// No variables, exit TUI and run as CLI
						m.Quitting = true
						return m, func() tea.Msg {
							os.Args = []string{os.Args[0], gadgetName}
							os.Exit(0)
							return nil
						}
					}
					// Has variables: prompt in TUI, then run
					m.Viewport = &RunGadgetWithVarsModel{
						GadgetName: gadgetName,
						VarNames:   varNames,
						Vars:       make(map[string]string),
						Index:      0,
						Config:     config,
						Parent:     m,
					}
					return m, nil
				case "Delete":
					m.Viewport = &ConfirmPromptModel{
						Message: fmt.Sprintf("Delete gadget: %s", gadgetName),
						OnConfirm: func() {
							scriptsMap, _ := scripts.LoadScripts()
							delete(scriptsMap, gadgetName)
							scripts.SaveScripts(scriptsMap)
						},
						ConfirmedMsg: "Gadget deleted!",
					}
					return m, nil
				case "Show info":
					scriptsMap, _ := scripts.LoadScripts()
					config, ok := scriptsMap[gadgetName]
					if !ok {
						m.Viewport = nil
						return m, nil
					}
					// Compose a styled info panel using style.GadgetInfoPanel
					info := style.GadgetInfoPanel(config.Command, gadgetName, config.Description, config.Variables, scripts.ExtractVariables(config.Command))
					m.Viewport = &InfoPanelModel{Content: info}
					return m, nil
				case "Edit":
					scriptsMap, _ := scripts.LoadScripts()
					config, ok := scriptsMap[gadgetName]
					if !ok {
						m.Viewport = nil
						return m, nil
					}
					model := scripts.InitialEditGadgetModel(gadgetName, config)
					m.Viewport = &model
					return m, m.Viewport.Init()
				}
			}
		case "up":
			col := m.Columns[m.ActiveColumn]
			if col.Selected > 0 {
				col.Selected--
			}
		case "down":
			col := m.Columns[m.ActiveColumn]
			if col.Selected < len(col.Options)-1 {
				col.Selected++
			}
		}
	}
	return m, nil
}

func (m *MultiColumnMenuModel) View() string {
	if m.Viewport != nil {
		// Always render the leftmost column with faint/dimmed style when a viewport is active
		left := style.ThinBorderStyle.Render(style.MenuDimmedStyle.Render(RenderColumn(m.Columns[0], false, true)))
		// If there are more columns, render them dimmed as well
		for i := 1; i < len(m.Columns); i++ {
			dimmed := style.ThinBorderStyle.Render(style.MenuDimmedStyle.Render(RenderColumn(m.Columns[i], false, true)))
			left = lipgloss.JoinHorizontal(lipgloss.Top, left, dimmed)
		}
		vp := m.Viewport.View()
		return lipgloss.JoinHorizontal(lipgloss.Top, left, vp)
	}

	cols := m.Columns
	var rendered []string
	for i, col := range cols {
		active := (i == m.ActiveColumn)
		dim := (i < m.ActiveColumn)
		var columnBlock string
		if dim {
			columnBlock = style.ThinBorderStyle.Render(style.MenuDimmedStyle.Render(RenderColumn(col, false, true)))
		} else {
			columnBlock = style.MenuColumnStyle.Render(RenderColumn(col, active, false))
			if i < len(cols)-1 {
				columnBlock = style.ThinBorderStyle.Render(columnBlock)
			}
		}
		rendered = append(rendered, columnBlock)
	}
	// Add navigation help at the bottom
	return lipgloss.JoinHorizontal(lipgloss.Top, rendered...) + "\n" + MenuNavHelp()
}

func InitialMenuModel(root *MenuColumn) *MultiColumnMenuModel {
	return &MultiColumnMenuModel{
		Columns:      []*MenuColumn{root},
		ActiveColumn: 0,
	}
}

func RunMenuTUI() {
	root := BuildRootMenu()
	p := tea.NewProgram(InitialMenuModel(root))
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running GoGoGadget TUI:", err)
		os.Exit(1)
	}
}

// addGadgetDoneWrapper and InitialAddGadgetModelWithDone have been moved to endpoints.go

// RenderColumn has been moved to columns.go

// InfoPanelModel is a simple Bubble Tea model for displaying static info content

type InfoPanelModel struct {
	Content string
}

func (m *InfoPanelModel) Init() tea.Cmd { return nil }

func (m *InfoPanelModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if keyMsg.String() == "esc" || keyMsg.String() == "enter" || keyMsg.String() == "q" {
			return nil, nil // Close the panel
		}
	}
	return m, nil
}

func (m *InfoPanelModel) View() string {
	var sb strings.Builder
	sb.WriteString(m.Content)
	sb.WriteString("\n\nPress Esc, Enter, or Q to return.")
	return sb.String()
}
