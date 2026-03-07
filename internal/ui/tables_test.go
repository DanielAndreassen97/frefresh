package ui

import (
	"testing"
)

func TestCategorizeTable(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{"Dim Employee", "Dim"},
		{"Fact Absence", "Fact"},
		{"Fakta Fravær", "Fact"},
		{"Log Refresh", "Log"},
		{"SomethingElse", "Other"},
	}
	for _, tt := range tests {
		got := categorizeTable(tt.name)
		if got != tt.expected {
			t.Errorf("categorizeTable(%q) = %q, want %q", tt.name, got, tt.expected)
		}
	}
}

func TestBuildItemsStructure(t *testing.T) {
	tables := []string{"Dim A", "Dim B", "Fact X", "Other Z"}
	items, groupMap, parentMap, allGroups, allItems := buildItems(tables)

	// First item is always "All tables"
	if items[0].kind != itemAll {
		t.Error("first item should be itemAll")
	}

	// Should have 3 groups: Dim, Fact, Other
	if len(allGroups) != 3 {
		t.Errorf("expected 3 groups, got %d", len(allGroups))
	}

	// Should have 4 table items
	if len(allItems) != 4 {
		t.Errorf("expected 4 table items, got %d", len(allItems))
	}

	// Each table should have a parent
	for _, ti := range allItems {
		if _, ok := parentMap[ti]; !ok {
			t.Errorf("table item %d has no parent", ti)
		}
	}

	// Each group should have children
	for _, gi := range allGroups {
		if len(groupMap[gi]) == 0 {
			t.Errorf("group %d has no children", gi)
		}
	}
}

func TestToggleAll(t *testing.T) {
	tables := []string{"Dim A", "Fact X"}
	m := newTableCheckModel("test", tables)

	m.toggle(0) // toggle "All tables"

	for _, idx := range m.allItems {
		if !m.items[idx].checked {
			t.Errorf("expected item %d to be checked after toggling all", idx)
		}
	}
	for _, idx := range m.allGroups {
		if !m.items[idx].checked {
			t.Errorf("expected group %d to be checked after toggling all", idx)
		}
	}

	selected := m.collectSelected()
	if len(selected) != 2 {
		t.Errorf("expected 2 selected, got %d", len(selected))
	}
}

func TestToggleGroup(t *testing.T) {
	tables := []string{"Dim A", "Dim B", "Fact X"}
	m := newTableCheckModel("test", tables)

	// Find the Dim group
	dimGroupIdx := -1
	for _, gi := range m.allGroups {
		if m.items[gi].group == "Dim" {
			dimGroupIdx = gi
			break
		}
	}
	if dimGroupIdx == -1 {
		t.Fatal("could not find Dim group")
	}

	m.toggle(dimGroupIdx)

	// Children should be checked
	for _, ci := range m.groupMap[dimGroupIdx] {
		if !m.items[ci].checked {
			t.Errorf("expected child %d to be checked", ci)
		}
	}

	selected := m.collectSelected()
	if len(selected) != 2 {
		t.Errorf("expected 2 selected (Dim A, Dim B), got %d: %v", len(selected), selected)
	}
}

func TestLockedWhenAllChecked(t *testing.T) {
	tables := []string{"Dim A", "Fact X"}
	m := newTableCheckModel("test", tables)

	m.toggle(0) // toggle All

	// Groups and tables should be locked
	for _, gi := range m.allGroups {
		if !m.isLocked(gi) {
			t.Errorf("group %d should be locked when All is checked", gi)
		}
	}
	for _, ti := range m.allItems {
		if !m.isLocked(ti) {
			t.Errorf("item %d should be locked when All is checked", ti)
		}
	}
}

func TestLockedWhenGroupChecked(t *testing.T) {
	tables := []string{"Dim A", "Dim B", "Fact X"}
	m := newTableCheckModel("test", tables)

	dimGroupIdx := -1
	for _, gi := range m.allGroups {
		if m.items[gi].group == "Dim" {
			dimGroupIdx = gi
			break
		}
	}

	m.toggle(dimGroupIdx)

	// Dim children should be locked
	for _, ci := range m.groupMap[dimGroupIdx] {
		if !m.isLocked(ci) {
			t.Errorf("child %d should be locked when parent group is checked", ci)
		}
	}
}

func TestToggleIndividualTable(t *testing.T) {
	tables := []string{"Dim A", "Dim B", "Fact X"}
	m := newTableCheckModel("test", tables)

	// Find first table item
	firstTable := m.allItems[0]
	m.toggle(firstTable)

	if !m.items[firstTable].checked {
		t.Error("expected individual table to be checked")
	}

	selected := m.collectSelected()
	if len(selected) != 1 {
		t.Errorf("expected 1 selected, got %d", len(selected))
	}
}

func TestEmptyGroups(t *testing.T) {
	// Only Dim tables — no Fact, Log, Other groups should appear
	tables := []string{"Dim A", "Dim B"}
	items, _, _, allGroups, _ := buildItems(tables)

	if len(allGroups) != 1 {
		t.Errorf("expected 1 group (Dim only), got %d", len(allGroups))
	}
	// Total items: 1 (All) + 1 (Dim group) + 2 (tables) = 4
	if len(items) != 4 {
		t.Errorf("expected 4 items, got %d", len(items))
	}
}
