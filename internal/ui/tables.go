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
	message   string
	items     []checkItem
	cursor    int
	selected  []string
	cancelled bool
	groupMap  map[int][]int // group index -> child indices
	parentMap map[int]int   // child index -> group index
	allGroups []int
	allItems  []int
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

func (m tableCheckModel) collectSelected() []string {
	if m.items[0].checked {
		var all []string
		for _, idx := range m.allItems {
			all = append(all, m.items[idx].tableName)
		}
		return all
	}
	var result []string
	for gIdx, children := range m.groupMap {
		if m.items[gIdx].checked {
			for _, ci := range children {
				result = append(result, m.items[ci].tableName)
			}
		} else {
			for _, ci := range children {
				if m.items[ci].checked {
					result = append(result, m.items[ci].tableName)
				}
			}
		}
	}
	sort.Strings(result)
	return result
}

func (m tableCheckModel) Init() tea.Cmd { return nil }

func (m tableCheckModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			m.cursor = (m.cursor - 1 + len(m.items)) % len(m.items)
		case "down", "j":
			m.cursor = (m.cursor + 1) % len(m.items)
		case " ":
			m.toggle(m.cursor)
		case "enter":
			m.selected = m.collectSelected()
			return m, tea.Quit
		case "esc", "ctrl+c":
			m.cancelled = true
			return m, tea.Quit
		}
	}
	return m, nil
}

var (
	pointerStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("6")).Bold(true)
	checkedStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Bold(true)
	dimmedStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	selectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Bold(true)
)

func (m tableCheckModel) View() string {
	var b strings.Builder
	fmt.Fprintf(&b, "? %s  (Space: toggle, Enter: confirm, Esc: cancel)\n\n", m.message)

	for i, item := range m.items {
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
		} else if isCursor && checked {
			label = selectedStyle.Render(label)
		}

		fmt.Fprintf(&b, "%s%s%s\n", pointer, box, label)
	}
	return b.String()
}

// TableCheckbox shows an interactive table selector with group toggle support.
// Returns selected table names, or nil + error if cancelled.
func TableCheckbox(message string, tables []string) ([]string, error) {
	model := newTableCheckModel(message, tables)
	p := tea.NewProgram(model)
	final, err := p.Run()
	if err != nil {
		return nil, err
	}
	result := final.(tableCheckModel)
	if result.cancelled {
		return nil, fmt.Errorf("cancelled")
	}
	return result.selected, nil
}
