package commands

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type pathItem struct {
	fullPath string
	filename string
}

func (p pathItem) FilterValue() string {
	return p.filename
}

func (p pathItem) Title() string {
	return p.filename
}

func (p pathItem) Description() string {
	return p.fullPath
}

type selectModel struct {
	list     list.Model
	choice   string
	quitting bool
}

func (m selectModel) Init() tea.Cmd {
	return nil
}

func (m selectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keyPress := msg.String(); keyPress {
		case "enter":
			if i, ok := m.list.SelectedItem().(pathItem); ok {
				m.choice = i.fullPath
			}
			return m, tea.Quit
		case "esc", "q":
			m.quitting = true
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.list.SetSize(msg.Width, msg.Height)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m selectModel) View() string {
	if m.quitting {
		return ""
	}
	return "\n" + m.list.View()
}

func SelectFromList(items []string, header string) (string, error) {
	if len(items) == 0 {
		return "", nil
	}

	var listItems []list.Item
	for _, path := range items {
		listItems = append(listItems, pathItem{
			fullPath: path,
			filename: filepath.Base(path),
		})
	}

	l := list.New(listItems, list.NewDefaultDelegate(), 0, 0)
	l.Title = header
	l.SetFilteringEnabled(true)
	l.SetShowStatusBar(false)
	l.SetShowPagination(true)
	l.SetShowHelp(true)

	m := selectModel{list: l}

	p := tea.NewProgram(m, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return "", err
	}

	finalSelectModel, ok := finalModel.(selectModel)
	if !ok {
		return "", fmt.Errorf("unexpected model type")
	}

	return finalSelectModel.choice, nil
}

type confirmModel struct {
	choices  []string
	cursor   int
	selected bool
	quitting bool
}

func (m confirmModel) Init() tea.Cmd {
	return nil
}

func (m confirmModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keyPress := msg.String(); keyPress {
		case "left", "h":
			if m.cursor > 0 {
				m.cursor--
			}
		case "right", "l", "tab":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case "enter":
			m.selected = true
			return m, tea.Quit
		case "esc", "q":
			m.quitting = true
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m confirmModel) View() string {
	if m.quitting {
		return ""
	}

	var choices []string
	for i, choice := range m.choices {
		if m.cursor == i {
			choices = append(choices, fmt.Sprintf("[ %s ]", choice))
		} else {
			choices = append(choices, fmt.Sprintf("  %s  ", choice))
		}
	}

	return fmt.Sprintf("\n  %s\n\n  %s\n", "Confirm?", strings.Join(choices, "  "))
}

func Confirm(prompt string) (bool, error) {
	m := confirmModel{
		choices: []string{"Yes", "No"},
		cursor:  1,
	}

	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return false, err
	}

	finalConfirmModel, ok := finalModel.(confirmModel)
	if !ok {
		return false, fmt.Errorf("unexpected model type")
	}

	return finalConfirmModel.selected && finalConfirmModel.cursor == 0, nil
}

type inputModel struct {
	textInput textinput.Model
	quitting  bool
}

func (m inputModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m inputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			return m, tea.Quit
		case "esc":
			m.quitting = true
			return m, tea.Quit
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m inputModel) View() string {
	if m.quitting {
		return ""
	}
	return fmt.Sprintf("\n  %s\n\n", m.textInput.View())
}

func Input(prompt string, placeholder string) (string, error) {
	ti := textinput.New()
	ti.Prompt = prompt
	ti.Placeholder = placeholder
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 50

	m := inputModel{textInput: ti}

	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return "", err
	}

	finalInputModel, ok := finalModel.(inputModel)
	if !ok {
		return "", fmt.Errorf("unexpected model type")
	}

	if finalInputModel.quitting {
		return "", nil
	}

	return strings.TrimSpace(finalInputModel.textInput.Value()), nil
}
