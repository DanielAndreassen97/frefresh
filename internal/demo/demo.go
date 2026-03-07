package demo

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/DanielAndreassen97/frefresh/internal/api"
	"github.com/DanielAndreassen97/frefresh/internal/config"
)

// MockAPIClient simulates the Power BI API for demo/recording purposes.
type MockAPIClient struct{}

func (MockAPIClient) GetAccessToken() (string, error) {
	time.Sleep(500 * time.Millisecond)
	return "demo-token", nil
}

func (MockAPIClient) GetWorkspaceID(token, workspaceName string) (string, error) {
	time.Sleep(300 * time.Millisecond)
	return "demo-workspace-id", nil
}

func (MockAPIClient) GetDatasetID(token, workspaceID, datasetName string) (string, error) {
	time.Sleep(200 * time.Millisecond)
	return "demo-dataset-id", nil
}

func (MockAPIClient) TriggerRefresh(token, workspaceID, datasetID string, tables []string) (string, error) {
	time.Sleep(300 * time.Millisecond)
	return "demo-request-42", nil
}

func (MockAPIClient) WaitForRefresh(token, workspaceID, datasetID, requestID string) (api.RefreshStatus, error) {
	time.Sleep(2 * time.Second)
	return api.RefreshStatus{Status: "Completed"}, nil
}

// SetupFixtures creates a temp directory with fake semantic model files and a config
// pointing to it. Returns the config path and a cleanup function.
func SetupFixtures() (string, func(), error) {
	tmpDir, err := os.MkdirTemp("", "frefresh-demo-*")
	if err != nil {
		return "", nil, err
	}
	cleanup := func() { os.RemoveAll(tmpDir) }

	repoDir := filepath.Join(tmpDir, "repo")
	configPath := filepath.Join(tmpDir, "config.json")

	// Create fake semantic model: HR
	hrDir := filepath.Join(repoDir, "HR.SemanticModel")
	tablesDir := filepath.Join(hrDir, "definition", "tables")
	os.MkdirAll(tablesDir, 0o755)

	platform := map[string]any{
		"metadata": map[string]any{"type": "SemanticModel", "displayName": "HR"},
		"config":   map[string]any{"logicalId": "demo-hr-001"},
	}
	data, _ := json.Marshal(platform)
	os.WriteFile(filepath.Join(hrDir, ".platform"), data, 0o644)

	dimTables := []string{"Dim Employee", "Dim Department", "Dim Location"}
	factTables := []string{"Fact Absence", "Fact Salary", "Fact Overtime"}
	for _, t := range append(dimTables, factTables...) {
		content := "table '" + t + "'\n\tpartition '" + t + "' = m\n\t\tmode: import\n"
		os.WriteFile(filepath.Join(tablesDir, t+".tmdl"), []byte(content), 0o644)
	}

	// Calculated table (should be excluded)
	os.WriteFile(filepath.Join(tablesDir, "_Measures.tmdl"),
		[]byte("table _Measures\n\tpartition _Measures = calculated\n\t\tmode: import\n"), 0o644)

	// Create second model: Finance
	finDir := filepath.Join(repoDir, "Finance.SemanticModel")
	finTablesDir := filepath.Join(finDir, "definition", "tables")
	os.MkdirAll(finTablesDir, 0o755)

	finPlatform := map[string]any{
		"metadata": map[string]any{"type": "SemanticModel", "displayName": "Finance"},
		"config":   map[string]any{"logicalId": "demo-fin-001"},
	}
	fd, _ := json.Marshal(finPlatform)
	os.WriteFile(filepath.Join(finDir, ".platform"), fd, 0o644)

	finTableNames := []string{"Dim Account", "Dim CostCenter", "Fact Budget", "Fact Actuals"}
	for _, t := range finTableNames {
		content := "table '" + t + "'\n\tpartition '" + t + "' = m\n\t\tmode: import\n"
		os.WriteFile(filepath.Join(finTablesDir, t+".tmdl"), []byte(content), 0o644)
	}

	// Create config with demo customer
	cfg := config.Config{
		Customers: map[string]config.Customer{
			"Contoso": {
				Path:             repoDir,
				WorkspacePattern: "DP - {env} - SemMod",
				Environments:     []string{"DEV", "TEST", "PROD"},
			},
		},
	}
	config.Save(configPath, cfg)

	return configPath, cleanup, nil
}
