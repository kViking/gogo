package menu

import (
	"fmt"
	"gogo/src/scripts"
	"gogo/src/style"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// Local wrapper for add gadget endpoint with done callback
func InitialAddGadgetModelWithDone(done func()) tea.Model {
	model := scripts.InitialAddGadgetModel()
	return &addGadgetDoneWrapper{model: model, done: done}
}

type addGadgetDoneWrapper struct {
	model tea.Model
	done  func()
}

func (w *addGadgetDoneWrapper) Init() tea.Cmd {
	return w.model.Init()
}

func (w *addGadgetDoneWrapper) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	model, cmd := w.model.Update(msg)
	w.model = model
	// If the add gadget model is finished (success or quit), call done
	if m, ok := model.(interface{ View() string }); ok {
		view := m.View()
		if view != "" && (strings.Contains(view, "Gadget added!") || strings.Contains(view, "Returning to menu")) && w.done != nil {
			w.done()
		}
	}
	return w, cmd
}

func (w *addGadgetDoneWrapper) View() string {
	return w.model.(interface{ View() string }).View()
}

type ConfirmPromptModel struct {
	Message      string
	Confirmed    bool
	Quitting     bool
	OnConfirm    func()
	ConfirmedMsg string // Optional custom message when confirmed
}

func (m ConfirmPromptModel) Init() tea.Cmd { return nil }

func (m ConfirmPromptModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "y", "Y", "enter":
			if m.OnConfirm != nil {
				m.OnConfirm()
			}
			m.Confirmed = true
			return m, func() tea.Msg { return EndpointDoneMsg{} }
		case "n", "N", "esc", "ctrl+[":
			m.Quitting = true
			return m, func() tea.Msg { return EndpointDoneMsg{} }
		}
	}
	return m, nil
}

func (m ConfirmPromptModel) View() string {
	if m.Confirmed {
		msg := m.ConfirmedMsg
		if msg == "" {
			msg = "Confirmed!"
		}
		return style.SuccessStyle.Render(msg) + "\nReturning to menu..."
	}
	if m.Quitting {
		return "Cancelled. Returning to menu..."
	}
	return style.HeaderStyle.Render(m.Message) +
		"\nAre you sure? (y/n)"
}

type EndpointDoneMsg struct{}

type RunGadgetWithVarsModel struct {
	GadgetName string
	VarNames   []string
	Vars       map[string]string
	Index      int
	Config     scripts.ScriptConfig
	Parent     *MultiColumnMenuModel
	Input      string
	Quitting   bool
}

func (m *RunGadgetWithVarsModel) Init() tea.Cmd { return nil }

func (m *RunGadgetWithVarsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "ctrl+[":
			m.Quitting = true
			if m.Parent != nil {
				m.Parent.Viewport = nil
			}
			return m, func() tea.Msg { return EndpointDoneMsg{} }
		case "enter":
			m.Vars[m.VarNames[m.Index]] = m.Input
			m.Input = ""
			m.Index++
			if m.Index >= len(m.VarNames) {
				args := []string{os.Args[0], m.GadgetName}
				for _, v := range m.VarNames {
					args = append(args, m.Vars[v])
				}
				m.Quitting = true
				return m, func() tea.Msg {
					os.Args = args
					os.Exit(0)
					return nil
				}
			}
		default:
			if len(msg.String()) == 1 {
				m.Input += msg.String()
			}
		}
	}
	return m, nil
}

func (m *RunGadgetWithVarsModel) View() string {
	if m.Quitting {
		return "Returning to menu..."
	}
	header := style.HeaderStyle.Render("Run gadget: " + m.GadgetName)
	promptStyle := style.PromptStyle
	var s string
	s += header + "\n\n"
	if m.Index < len(m.VarNames) {
		v := m.VarNames[m.Index]
		desc := scripts.GetVariableDescription(v, m.Config)
		s += promptStyle.Render(fmt.Sprintf("Enter %s: ", desc)) + m.Input + "_"
	}
	return s
}
