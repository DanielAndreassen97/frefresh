package ui

import (
	"errors"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

// ErrGoBack is returned when the user presses Esc/b to go back.
var ErrGoBack = errors.New("go back")

// ErrQuit is returned when the user presses Ctrl+C/q to quit.
var ErrQuit = errors.New("quit")

type MenuOption struct {
	Label string
	Value string
}

type menuModel struct {
	message  string
	options  []MenuOption
	cursor   int
	selected string
	goBack   bool
	quit     bool
	done     bool
}

func (m menuModel) Init() tea.Cmd { return nil }

func (m menuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			m.cursor = (m.cursor - 1 + len(m.options)) % len(m.options)
		case "down", "j":
			m.cursor = (m.cursor + 1) % len(m.options)
		case "enter":
			m.selected = m.options[m.cursor].Value
			m.done = true
			return m, tea.Quit
		case "esc", "b":
			m.goBack = true
			m.done = true
			return m, tea.Quit
		case "ctrl+c", "q":
			m.quit = true
			m.done = true
			return m, tea.Quit
		default:
			if len(msg.String()) == 1 && msg.String()[0] >= '1' && msg.String()[0] <= '9' {
				idx := int(msg.String()[0] - '1')
				if idx < len(m.options) {
					m.selected = m.options[idx].Value
					m.done = true
					return m, tea.Quit
				}
			}
		}
	}
	return m, nil
}

var (
	menuPointerStyle  = lipgloss.NewStyle().Foreground(AccentColor).Bold(true)
	menuNumberStyle   = lipgloss.NewStyle().Foreground(DimColor)
	menuSelectedStyle = lipgloss.NewStyle().Foreground(AccentColor)
)

func (m menuModel) selectedLabel() string {
	for _, opt := range m.options {
		if opt.Value == m.selected {
			return opt.Label
		}
	}
	return ""
}

func (m menuModel) View() string {
	// After selection: show one-line summary instead of full menu
	if m.done {
		if m.goBack || m.quit {
			return ""
		}
		return menuSelectedStyle.Render(fmt.Sprintf("  %s: %s", m.message, m.selectedLabel())) + "\n"
	}

	var b strings.Builder
	fmt.Fprintf(&b, "  %s\n\n", m.message)

	for i, opt := range m.options {
		pointer := "  "
		if i == m.cursor {
			pointer = menuPointerStyle.Render("❯ ")
		}

		num := menuNumberStyle.Render(fmt.Sprintf("%d)", i+1))
		fmt.Fprintf(&b, "%s%s %s\n", pointer, num, opt.Label)
	}

	return b.String()
}

// MenuOptionsFromStrings creates MenuOptions where Label and Value are both the string value.
func MenuOptionsFromStrings(values []string) []MenuOption {
	opts := make([]MenuOption, len(values))
	for i, v := range values {
		opts[i] = MenuOption{Label: v, Value: v}
	}
	return opts
}

// NumberMenu shows a select menu with digit key support.
// Returns ErrGoBack if the user presses Esc/b.
func NumberMenu(message string, options []MenuOption) (string, error) {
	model := menuModel{message: message, options: options}
	p := tea.NewProgram(model)
	final, err := p.Run()
	if err != nil {
		return "", err
	}
	result := final.(menuModel)
	if result.quit {
		return "", ErrQuit
	}
	if result.goBack {
		return "", ErrGoBack
	}
	return result.selected, nil
}

// Confirm shows a yes/no prompt. Returns true for yes.
func Confirm(message string) (bool, error) {
	var result bool
	theme := huh.ThemeBase()
	theme.Focused.FocusedButton = lipgloss.NewStyle().Foreground(lipgloss.Color("0")).Background(AccentColor).Padding(0, 1)
	theme.Focused.BlurredButton = lipgloss.NewStyle().Foreground(DimColor).Padding(0, 1)
	theme.Focused.Title = lipgloss.NewStyle().Foreground(AccentColor).Bold(true)

	km := huh.NewDefaultKeyMap()
	km.Quit = key.NewBinding(key.WithKeys("ctrl+c", "esc"))

	err := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(message).
				Affirmative("Yes").
				Negative("No").
				Value(&result),
		),
	).WithTheme(theme).WithKeyMap(km).Run()
	if err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return false, ErrGoBack
		}
		return false, err
	}
	return result, nil
}
