package prompt

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// ---------- SELECT MODEL ----------
type selectModel struct {
	filter   string
	cursor   int
	quit     bool
	options  []string
	filtered []string
	message  string
}

func FilterBySubstring(src []string, needle string) []string {
	var result []string
	if needle == "" {
		return src
	}

	for _, v := range src {
		if strings.Contains(strings.ToLower(v), strings.ToLower(needle)) {
			result = append(result, v)
		}
	}
	return result
}

func (m *selectModel) Init() tea.Cmd { return nil }

func (m *selectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "ctrl+k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "ctrl+j":
			if m.cursor < len(m.filtered)-1 {
				m.cursor++
			}
		case "enter":
			return m, tea.Quit
		case "ctrl+c":
			m.quit = true
			return m, tea.Quit
		case "backspace":
			runes := []rune(m.filter)
			if len(runes) > 0 {
				m.filter = string(runes[:len(runes)-1])
			}
		default:
			if msg.Type == tea.KeyRunes {
				m.cursor = 0
				m.filter += string(msg.Runes)
			}
		}
	}
	m.filtered = FilterBySubstring(m.options, m.filter)

	return m, nil
}

func (m *selectModel) View() string {
	s := fmt.Sprintf("%s\n", m.message)
	s += fmt.Sprintf("filtering by: %s\n", m.filter)
	for i, o := range m.filtered {
		prefix := "  "
		if i == m.cursor {
			prefix = "➜ "
		}
		s += fmt.Sprintf("%s %s\n", prefix, o)
	}
	return s
}

func IndexOf(arr []string, target string) int {
	for i, v := range arr {
		if v == target {
			return i
		}
	}
	return -1 // not found
}

func SelectPrompt(options []string, message string) (int, error) {
	m := selectModel{quit: false, options: options, message: message, filtered: options, filter: ""}
	p := tea.NewProgram(&m)
	if _, err := p.Run(); err != nil {
		return -1, err
	}

	if m.quit {
		os.Exit(0)
	}

	return IndexOf(m.options, m.filtered[m.cursor]), nil
}

// ---------- INPUT MODEL ----------
type inputModel struct {
	quit     bool
	cursor   int
	input    string
	message  string
	accepted bool
}

func (m *inputModel) Init() tea.Cmd { return nil }

func (m *inputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			m.accepted = true
			return m, tea.Quit
		case "ctrl+c":
			m.quit = true
			return m, tea.Quit
		case "backspace":
			if len(m.input) > 0 {
				m.input = m.input[:len(m.input)-1]
			}
		case " ":
			// Explicitly handle space character
			m.input += " "
		default:
			if msg.Type == tea.KeyRunes {
				m.input += string(msg.Runes)
			}
		}
	}
	return m, nil
}

func (m *inputModel) View() string {
	prompt := fmt.Sprintf("%s\n> %s", m.message, m.input)
	return prompt
}

// InputPrompt asks the user for a single line of text.
func InputPrompt(message string, defaultText string) (string, error) {
	m := inputModel{quit: false, message: message, input: defaultText}
	p := tea.NewProgram(&m)
	if _, err := p.Run(); err != nil {
		return "", err
	}

	if m.quit {
		os.Exit(0)
	}

	return m.input, nil
}
