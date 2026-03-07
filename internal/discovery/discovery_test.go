package discovery

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func setupFakeRepo(t *testing.T) string {
	t.Helper()
	base := t.TempDir()

	// Semantic model
	smDir := filepath.Join(base, "HR.SemanticModel")
	tablesDir := filepath.Join(smDir, "definition", "tables")
	os.MkdirAll(tablesDir, 0o755)

	platform := map[string]any{
		"metadata": map[string]any{"type": "SemanticModel", "displayName": "HR"},
		"config":   map[string]any{"logicalId": "abc-123"},
	}
	data, _ := json.Marshal(platform)
	os.WriteFile(filepath.Join(smDir, ".platform"), data, 0o644)

	// Import tables
	os.WriteFile(filepath.Join(tablesDir, "Dim Employee.tmdl"),
		[]byte("table 'Dim Employee'\n\tpartition 'Dim Employee' = m\n\t\tmode: import\n"), 0o644)
	os.WriteFile(filepath.Join(tablesDir, "Fact Absence.tmdl"),
		[]byte("table 'Fact Absence'\n\tpartition 'Fact Absence' = m\n\t\tmode: import\n"), 0o644)

	// Calculated table (should be excluded)
	os.WriteFile(filepath.Join(tablesDir, "_Measures.tmdl"),
		[]byte("table _Measures\n\tpartition _Measures = calculated\n\t\tmode: import\n"), 0o644)

	// Calculation group (should be excluded)
	os.WriteFile(filepath.Join(tablesDir, "TimeIntelligence.tmdl"),
		[]byte("table TimeIntelligence\n\tlineageTag: abc\n\n\tcalculationGroup\n"), 0o644)

	// Report (should be ignored)
	reportDir := filepath.Join(base, "SomeReport.Report")
	os.MkdirAll(reportDir, 0o755)
	rp := map[string]any{"metadata": map[string]any{"type": "Report", "displayName": "SomeReport"}}
	rd, _ := json.Marshal(rp)
	os.WriteFile(filepath.Join(reportDir, ".platform"), rd, 0o644)

	return base
}

func TestDiscoverModels(t *testing.T) {
	base := setupFakeRepo(t)
	models, err := DiscoverModels(base)
	if err != nil {
		t.Fatal(err)
	}
	if len(models) != 1 {
		t.Fatalf("expected 1 model, got %d", len(models))
	}
	if models[0].Name != "HR" {
		t.Errorf("expected HR, got %s", models[0].Name)
	}
}

func TestDiscoverModelsIgnoresReports(t *testing.T) {
	base := setupFakeRepo(t)
	models, _ := DiscoverModels(base)
	for _, m := range models {
		if m.Name == "SomeReport" {
			t.Error("should not include reports")
		}
	}
}

func TestDiscoverTables(t *testing.T) {
	base := setupFakeRepo(t)
	models, _ := DiscoverModels(base)
	tables, err := DiscoverTables(models[0])
	if err != nil {
		t.Fatal(err)
	}
	if len(tables) != 2 {
		t.Fatalf("expected 2 tables, got %d: %v", len(tables), tables)
	}
}

func TestDiscoverTablesExcludesCalculated(t *testing.T) {
	base := setupFakeRepo(t)
	models, _ := DiscoverModels(base)
	tables, _ := DiscoverTables(models[0])
	for _, table := range tables {
		if table == "_Measures" {
			t.Error("should exclude calculated tables")
		}
	}
}

func TestDiscoverTablesExcludesCalculationGroups(t *testing.T) {
	base := setupFakeRepo(t)
	models, _ := DiscoverModels(base)
	tables, _ := DiscoverTables(models[0])
	for _, table := range tables {
		if table == "TimeIntelligence" {
			t.Error("should exclude calculation groups")
		}
	}
}
