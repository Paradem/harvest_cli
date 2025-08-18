package prompt

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// ---------- SELECT MODEL ----------
type selectModel struct {
	cursor  int
	options []string
	message string
}

func (m selectModel) Init() tea.Cmd { return nil }

func (m selectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.options)-1 {
				m.cursor++
			}
		case "enter":
			return m, tea.Quit
		case "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m selectModel) View() string {
	s := fmt.Sprintf("%s\n", m.message)
	for i, o := range m.options {
		prefix := "  "
		if i == m.cursor {
			prefix = "➜ "
		}
		s += fmt.Sprintf("%s %s\n", prefix, o)
	}
	return s
}

// SelectPrompt shows a list of options and returns the index of the chosen one.
func SelectPrompt(options []string, message string) (int, error) {
	m := selectModel{options: options, message: message}
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		return -1, err
	}
	return m.cursor, nil
}

// ---------- INPUT MODEL ----------
type inputModel struct {
	cursor   int
	input    string
	message  string
	accepted bool
}

func (m inputModel) Init() tea.Cmd { return nil }

func (m inputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			m.accepted = true
			return m, tea.Quit
		case "ctrl+c":
			return m, tea.Quit
		case "backspace":
			if len(m.input) > 0 {
				m.input = m.input[:len(m.input)-1]
			}
		default:
			if msg.Type == tea.KeyRunes {
				m.input += string(msg.Runes)
			}
		}
	}
	return m, nil
}

func (m inputModel) View() string {
	prompt := fmt.Sprintf("%s\n> %s", m.message, m.input)
	return prompt
}

// InputPrompt asks the user for a single line of text.
func InputPrompt(message string, defaultText string) (string, error) {
	m := inputModel{message: message, input: defaultText}
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		return "", err
	}
	return m.input, nil
}
