package analyze

import (
	bspinner "github.com/charmbracelet/bubbles/spinner"
	textinput "github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type Step int

const (
	Input Step = iota
	Spinner
	Results
	ConfirmSave
	Done
	Error
)

type Suggestion struct {
	VarName  string
	Original string
}

type Model struct {
	Step         Step
	Input        textinput.Model
	Spinner      bspinner.Model
	CmdStr       string
	OrigHL       string
	ParamHL      string
	Suggestions  []Suggestion
	Error        error
	ParamStr     string
	ConfirmInput textinput.Model
	ParentReturn func(tea.Model)
}

func NewModel(parentReturn func(tea.Model)) tea.Model {
	ti := textinput.New()
	ti.Placeholder = "Enter PowerShell command to analyze"
	ti.Focus()
	ti.CharLimit = 512
	ti.Width = 60
	ti.Prompt = "> "
	// CursorStyle should be provided by the parent package or set to default

	confirm := textinput.New()
	confirm.Placeholder = "Y/N"
	confirm.CharLimit = 3
	confirm.Width = 5
	confirm.Prompt = "> "

	sp := bspinner.New()
	sp.Spinner = bspinner.Dot

	return &Model{
		Step:         Input,
		Input:        ti,
		Spinner:      sp,
		ConfirmInput: confirm,
		ParentReturn: parentReturn,
	}
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m *Model) View() string {
	return "Analyze TUI (refactor in progress)"
}
