package prompt

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// ---------- SELECT MODEL ----------
type selectModel struct {
	filter      string
	cursor      int
	quit        bool
	options     []string
	filtered    []string
	message     string
	lazyMode    bool
	hasTyped    bool
	offset      int
	maxVisible  int
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
			if m.lazyMode && !m.hasTyped {
				break
			}
			if m.cursor > 0 {
				m.cursor--
				m.adjustOffset()
			}
		case "down", "ctrl+j":
			if m.lazyMode && !m.hasTyped {
				break
			}
			if m.cursor < len(m.filtered)-1 {
				m.cursor++
				m.adjustOffset()
			}
		case "enter":
			if m.lazyMode && !m.hasTyped {
				break
			}
			return m, tea.Quit
		case "ctrl+c":
			m.quit = true
			return m, tea.Quit
		case "backspace":
			runes := []rune(m.filter)
			if len(runes) > 0 {
				m.filter = string(runes[:len(runes)-1])
				if m.lazyMode {
					m.hasTyped = len(m.filter) > 0
				}
				m.cursor = 0
				m.offset = 0
			}
		default:
			if msg.Type == tea.KeyRunes {
				m.cursor = 0
				m.filter += string(msg.Runes)
				if m.lazyMode {
					m.hasTyped = true
				}
				m.offset = 0
			}
		}
	}
	if !m.lazyMode || m.hasTyped {
		m.filtered = FilterBySubstring(m.options, m.filter)
	} else {
		m.filtered = []string{}
	}

	return m, nil
}

func (m *selectModel) adjustOffset() {
	if m.cursor < m.offset {
		m.offset = m.cursor
	} else if m.cursor >= m.offset+m.maxVisible {
		m.offset = m.cursor - m.maxVisible + 1
	}
}

func (m *selectModel) View() string {
	s := fmt.Sprintf("%s\n", m.message)
	if m.lazyMode && !m.hasTyped {
		s += "Start typing to search...\n"
	} else {
		s += fmt.Sprintf("filtering by: %s\n", m.filter)
		
		// Determine which items to display
		start := m.offset
		end := start + m.maxVisible
		if end > len(m.filtered) {
			end = len(m.filtered)
		}
		
		// Display visible items
		for i := start; i < end; i++ {
			o := m.filtered[i]
			prefix := "  "
			if i == m.cursor {
				prefix = "➜ "
			}
			s += fmt.Sprintf("%s %s\n", prefix, o)
		}
		
		// Show scroll indicator if there are more items
		if len(m.filtered) > m.maxVisible {
			scrollInfo := fmt.Sprintf("(%d-%d/%d)", start+1, end, len(m.filtered))
			s += fmt.Sprintf("  ...%s\n", scrollInfo)
		}
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
	return SelectPromptWithOptions(options, message, false)
}

func SelectPromptWithOptions(options []string, message string, lazyMode bool) (int, error) {
	return SelectPromptWithVisibleLimit(options, message, lazyMode, 15) // Default 15 visible items
}

func SelectPromptWithVisibleLimit(options []string, message string, lazyMode bool, maxVisible int) (int, error) {
	m := selectModel{
		quit:        false, 
		options:     options, 
		message:     message, 
		filtered:    options, 
		filter:      "", 
		lazyMode:    lazyMode, 
		hasTyped:    false,
		offset:      0,
		maxVisible:  maxVisible,
	}
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
	textInput textinput.Model
	message   string
	quit      bool
}

func (m *inputModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m *inputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			return m, tea.Quit
		case tea.KeyCtrlC:
			m.quit = true
			return m, tea.Quit
		}
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m *inputModel) View() string {
	return fmt.Sprintf("%s\n%s", m.message, m.textInput.View())
}

// InputPrompt asks the user for a single line of text.
func InputPrompt(message string, defaultText string) (string, error) {
	ti := textinput.New()
	ti.Placeholder = ""
	ti.Focus()
	ti.SetValue(defaultText)

	m := inputModel{
		textInput: ti,
		message:   message,
		quit:      false,
	}

	p := tea.NewProgram(&m)
	if _, err := p.Run(); err != nil {
		return "", err
	}

	if m.quit {
		os.Exit(0)
	}

	return m.textInput.Value(), nil
}
