package ui

import (
	"fmt"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func categorizeTable(name string) string {
	switch {
	case strings.HasPrefix(name, "Dim"):
		return "Dim"
	case strings.HasPrefix(name, "Fact"), strings.HasPrefix(name, "Fakta"):
		return "Fact"
	case strings.HasPrefix(name, "Log"):
		return "Log"
	default:
		return "Other"
	}
}

type itemType int

const (
	itemAll itemType = iota
	itemGroup
	itemTable
)

type checkItem struct {
	kind      itemType
	label     string
	tableName string
	group     string
	checked   bool
}

type tableCheckModel struct {
	message    string
	items      []checkItem
	cursor     int
	selection  TableSelection
	goBack     bool
	quit       bool
	done       bool
	groupMap   map[int][]int // group index -> child indices
	parentMap  map[int]int   // child index -> group index
	allGroups  []int
	allItems   []int
	termHeight int // terminal height for viewport scrolling
}

func buildItems(tables []string) ([]checkItem, map[int][]int, map[int]int, []int, []int) {
	groups := map[string][]string{"Dim": {}, "Fact": {}, "Log": {}, "Other": {}}
	for _, t := range tables {
		g := categorizeTable(t)
		groups[g] = append(groups[g], t)
	}

	var items []checkItem
	groupMap := map[int][]int{}
	parentMap := map[int]int{}
	var allGroups, allItems []int

	items = append(items, checkItem{kind: itemAll, label: fmt.Sprintf("All tables (%d)", len(tables))})

	for _, gName := range []string{"Dim", "Fact", "Log", "Other"} {
		gTables := groups[gName]
		if len(gTables) == 0 {
			continue
		}
		sort.Strings(gTables)
		gIdx := len(items)
		items = append(items, checkItem{kind: itemGroup, label: fmt.Sprintf("All %s (%d)", gName, len(gTables)), group: gName})
		allGroups = append(allGroups, gIdx)

		var children []int
		for _, t := range gTables {
			cIdx := len(items)
			items = append(items, checkItem{kind: itemTable, label: t, tableName: t, group: gName})
			children = append(children, cIdx)
			allItems = append(allItems, cIdx)
			parentMap[cIdx] = gIdx
		}
		groupMap[gIdx] = children
	}

	return items, groupMap, parentMap, allGroups, allItems
}

func newTableCheckModel(message string, tables []string) tableCheckModel {
	items, groupMap, parentMap, allGroups, allItems := buildItems(tables)
	return tableCheckModel{
		message:   message,
		items:     items,
		groupMap:  groupMap,
		parentMap: parentMap,
		allGroups: allGroups,
		allItems:  allItems,
	}
}

func (m tableCheckModel) isLocked(idx int) bool {
	item := m.items[idx]
	if item.kind == itemAll {
		return false
	}
	if m.items[0].checked {
		return true
	}
	if item.kind == itemTable {
		parentIdx := m.parentMap[idx]
		return m.items[parentIdx].checked
	}
	return false
}

func (m *tableCheckModel) toggle(idx int) {
	if m.isLocked(idx) {
		return
	}
	item := &m.items[idx]
	item.checked = !item.checked

	if item.kind == itemAll {
		for _, gi := range m.allGroups {
			m.items[gi].checked = item.checked
		}
		for _, ti := range m.allItems {
			m.items[ti].checked = item.checked
		}
	} else if item.kind == itemGroup {
		for _, ci := range m.groupMap[idx] {
			m.items[ci].checked = item.checked
		}
	}
}

// TableSelection holds the result of the table checkbox.
type TableSelection struct {
	FullRefresh bool     // true = refresh entire model, no specific tables
	Tables      []string // specific table names (only when FullRefresh is false)
	Summary     string   // human-readable description for display
}

func (m tableCheckModel) collectSelection() TableSelection {
	if m.items[0].checked {
		return TableSelection{FullRefresh: true, Summary: "Full model refresh"}
	}

	var tables []string
	var summaryParts []string

	for _, gIdx := range m.allGroups {
		group := m.items[gIdx]
		children := m.groupMap[gIdx]

		if group.checked {
			// Whole group selected — add tables but use group label in summary
			for _, ci := range children {
				tables = append(tables, m.items[ci].tableName)
			}
			summaryParts = append(summaryParts, fmt.Sprintf("All %s (%d)", group.group, len(children)))
		} else {
			for _, ci := range children {
				if m.items[ci].checked {
					tables = append(tables, m.items[ci].tableName)
					summaryParts = append(summaryParts, m.items[ci].tableName)
				}
			}
		}
	}

	sort.Strings(tables)
	return TableSelection{Tables: tables, Summary: strings.Join(summaryParts, ", ")}
}

func (m tableCheckModel) Init() tea.Cmd { return nil }

const jumpSize = 5

func (m tableCheckModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.termHeight = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			m.cursor = (m.cursor - 1 + len(m.items)) % len(m.items)
		case "down", "j":
			m.cursor = (m.cursor + 1) % len(m.items)
		case "alt+up", "alt+k":
			m.cursor -= jumpSize
			if m.cursor < 0 {
				m.cursor = 0
			}
		case "alt+down", "alt+j":
			m.cursor += jumpSize
			if m.cursor >= len(m.items) {
				m.cursor = len(m.items) - 1
			}
		case " ":
			m.toggle(m.cursor)
		case "enter":
			m.selection = m.collectSelection()
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
		}
	}
	return m, nil
}

var (
	pointerStyle  = lipgloss.NewStyle().Foreground(AccentColor).Bold(true)
	checkedStyle      = lipgloss.NewStyle().Foreground(AccentColor).Bold(true)
	dimmedStyle       = lipgloss.NewStyle().Foreground(DimColor)
	cursorStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff")).Bold(true)
	checkedLabelStyle = lipgloss.NewStyle().Foreground(AccentColor)
)

func (m tableCheckModel) renderItem(i int) string {
	item := m.items[i]
	locked := m.isLocked(i)
	checked := item.checked || locked
	isCursor := i == m.cursor

	pointer := "  "
	if isCursor {
		pointer = pointerStyle.Render("❯ ")
	}

	box := "□ "
	if checked {
		box = checkedStyle.Render("■ ")
	}

	var label string
	if item.kind == itemAll || item.kind == itemGroup {
		label = fmt.Sprintf("── %s ──", item.label)
	} else {
		label = fmt.Sprintf("    %s", item.label)
	}

	if locked {
		box = dimmedStyle.Render("■ ")
		label = dimmedStyle.Render(label)
	} else if isCursor {
		label = cursorStyle.Render(label)
	} else if checked {
		label = checkedLabelStyle.Render(label)
	}

	return fmt.Sprintf("%s%s%s", pointer, box, label)
}

func (m tableCheckModel) View() string {
	if m.done {
		if m.goBack || m.quit {
			return ""
		}
		return checkedStyle.Render(fmt.Sprintf("  Tables: %s", m.selection.Summary)) + "\n"
	}

	var b strings.Builder
	header := fmt.Sprintf("  %s  (space to toggle, alt+↑↓ to jump)\n\n", m.message)
	b.WriteString(header)

	// header takes 2 lines, reserve 1 for safety
	maxVisible := m.termHeight - 3
	if maxVisible <= 0 || maxVisible >= len(m.items) {
		// Terminal height unknown or everything fits — render all
		for i := range m.items {
			fmt.Fprintf(&b, "%s\n", m.renderItem(i))
		}
		return b.String()
	}

	// Reserve 2 lines for potential ↑/↓ indicators
	itemSlots := maxVisible - 2
	if itemSlots < 1 {
		itemSlots = 1
	}

	start := m.cursor - itemSlots/2
	if start < 0 {
		start = 0
	}
	end := start + itemSlots
	if end > len(m.items) {
		end = len(m.items)
		start = end - itemSlots
		if start < 0 {
			start = 0
		}
	}

	if start > 0 {
		fmt.Fprintf(&b, "  %s\n", dimmedStyle.Render(fmt.Sprintf("↑ %d more above", start)))
	}

	for i := start; i < end; i++ {
		fmt.Fprintf(&b, "%s\n", m.renderItem(i))
	}

	if end < len(m.items) {
		fmt.Fprintf(&b, "  %s\n", dimmedStyle.Render(fmt.Sprintf("↓ %d more below", len(m.items)-end)))
	}

	return b.String()
}

// TableCheckbox shows an interactive table selector with group toggle support.
// Returns a TableSelection describing what was chosen, or error if cancelled.
func TableCheckbox(message string, tables []string) (TableSelection, error) {
	model := newTableCheckModel(message, tables)
	p := tea.NewProgram(model)
	final, err := p.Run()
	if err != nil {
		return TableSelection{}, err
	}
	result := final.(tableCheckModel)
	if result.quit {
		return TableSelection{}, ErrQuit
	}
	if result.goBack {
		return TableSelection{}, ErrGoBack
	}
	return result.selection, nil
}
