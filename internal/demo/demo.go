package demo

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/DanielAndreassen97/frefresh/internal/api"
	"github.com/DanielAndreassen97/frefresh/internal/config"
)

// MockAPIClient simulates the Power BI API for demo/recording purposes.
type MockAPIClient struct{}

func (MockAPIClient) GetAccessToken(customer string) (string, error) {
	time.Sleep(500 * time.Millisecond)
	return "demo-token", nil
}

func (MockAPIClient) GetWorkspaceID(token, workspaceName string) (string, error) {
	time.Sleep(300 * time.Millisecond)
	return "00000000-0000-0000-0000-000000000001", nil
}

func (MockAPIClient) ListDatasets(token, workspaceID string) ([]api.Dataset, error) {
	time.Sleep(200 * time.Millisecond)
	return []api.Dataset{
		{ID: "00000000-0000-0000-0000-000000000010", Name: "HR"},
		{ID: "00000000-0000-0000-0000-000000000020", Name: "Finance"},
	}, nil
}

func (MockAPIClient) QueryTables(token, workspaceID, datasetID string) ([]string, error) {
	time.Sleep(300 * time.Millisecond)
	if datasetID == "00000000-0000-0000-0000-000000000020" {
		return []string{"Dim Account", "Dim CostCenter", "Fact Actuals", "Fact Budget"}, nil
	}
	return []string{"Dim Department", "Dim Employee", "Dim Location", "Fact Absence", "Fact Overtime", "Fact Salary"}, nil
}

func (MockAPIClient) TriggerRefresh(token, workspaceID, datasetID string, tables []string) (string, error) {
	time.Sleep(300 * time.Millisecond)
	return "demo-request-42", nil
}

func (MockAPIClient) WaitForRefresh(token, workspaceID, datasetID, requestID string) (api.RefreshStatus, error) {
	time.Sleep(2 * time.Second)
	return api.RefreshStatus{Status: "Completed"}, nil
}

// SetupFixtures creates a temp directory with a config for demo mode.
// Returns the config path and a cleanup function.
func SetupFixtures() (string, func(), error) {
	tmpDir, err := os.MkdirTemp("", "frefresh-demo-*")
	if err != nil {
		return "", nil, err
	}
	cleanup := func() { os.RemoveAll(tmpDir) }

	configPath := filepath.Join(tmpDir, "config.json")

	cfg := config.Config{
		Customers: map[string]config.Customer{
			"Contoso": {
				WorkspacePattern: "DP - {env} - SemMod",
				Environments:     []string{"DEV", "TEST", "PROD"},
			},
			"Northwind": {
				WorkspacePattern: "NW - {env} - Analytics",
				Environments:     []string{"DEV", "PROD"},
			},
		},
	}
	if err := config.Save(configPath, cfg); err != nil {
		cleanup()
		return "", nil, fmt.Errorf("failed to save demo config: %w", err)
	}

	return configPath, cleanup, nil
}
