package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/DanielAndreassen97/frefresh/internal/api"
	"github.com/DanielAndreassen97/frefresh/internal/config"
	"github.com/DanielAndreassen97/frefresh/internal/ui"
	"github.com/charmbracelet/lipgloss"
)

var (
	successStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("2")).
			Foreground(lipgloss.Color("2")).
			Padding(0, 1)

	errorStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("1")).
			Foreground(lipgloss.Color("1")).
			Padding(0, 1)

	infoStyle = lipgloss.NewStyle().Foreground(ui.AccentColor)
)

// APIClient abstracts the Power BI API calls for testability and demo mode.
type APIClient interface {
	GetAccessToken(customer string) (string, error)
	GetWorkspaceID(token, workspaceName string) (string, error)
	ListDatasets(token, workspaceID string) ([]api.Dataset, error)
	QueryTables(token, workspaceID, datasetID string) ([]string, error)
	TriggerRefresh(token, workspaceID, datasetID string, tables []string) (string, error)
	WaitForRefresh(token, workspaceID, datasetID, requestID string) (api.RefreshStatus, error)
}

// RealAPIClient calls the actual Power BI REST API.
type RealAPIClient struct{}

func (RealAPIClient) GetAccessToken(customer string) (string, error) {
	return api.GetAccessToken(customer)
}
func (RealAPIClient) GetWorkspaceID(token, workspaceName string) (string, error) {
	return api.GetWorkspaceID(token, workspaceName)
}
func (RealAPIClient) ListDatasets(token, workspaceID string) ([]api.Dataset, error) {
	return api.ListDatasets(token, workspaceID)
}
func (RealAPIClient) QueryTables(token, workspaceID, datasetID string) ([]string, error) {
	return api.QueryTables(token, workspaceID, datasetID)
}
func (RealAPIClient) TriggerRefresh(token, workspaceID, datasetID string, tables []string) (string, error) {
	return api.TriggerRefresh(token, workspaceID, datasetID, tables)
}
func (RealAPIClient) WaitForRefresh(token, workspaceID, datasetID, requestID string) (api.RefreshStatus, error) {
	return api.WaitForRefresh(token, workspaceID, datasetID, requestID, 5*time.Second, 30*time.Minute)
}

// DefaultAPI is the API client used by the refresh command. Override for demo mode.
var DefaultAPI APIClient = RealAPIClient{}

func Refresh(configPath string) error {
	return RefreshWithAPI(configPath, DefaultAPI)
}

func RefreshWithAPI(configPath string, client APIClient) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return err
	}
	if len(cfg.Customers) == 0 {
		fmt.Println("No customers configured. Add a customer first.")
		return nil
	}

	customerName, customer, err := selectCustomer(cfg)
	if err != nil {
		return err
	}

	env, err := selectEnvironment(customer)
	if err != nil {
		return err
	}

	workspaceName := strings.ReplaceAll(customer.WorkspacePattern, "{env}", env)

	// Authenticate early — we need the token to discover datasets and tables
	fmt.Println()
	fmt.Println(infoStyle.Render("Authenticating..."))
	token, err := client.GetAccessToken(customerName)
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}
	fmt.Println(infoStyle.Render("Authenticated."))
	fmt.Println()

	// Resolve workspace
	workspaceID, err := client.GetWorkspaceID(token, workspaceName)
	if err != nil {
		return err
	}

	// List datasets from the API
	datasets, err := client.ListDatasets(token, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to list datasets: %w", err)
	}
	if len(datasets) == 0 {
		return fmt.Errorf("no datasets found in workspace '%s'", workspaceName)
	}

	dataset, err := selectDataset(datasets)
	if err != nil {
		return err
	}

	// Query actual deployed tables via Fabric API
	tableSpinner := ui.NewSpinner("Retrieving tables...")
	tableSpinner.Start()

	tables, err := client.QueryTables(token, workspaceID, dataset.ID)
	tableSpinner.Stop()
	if err != nil {
		return fmt.Errorf("failed to query tables: %w", err)
	}
	if len(tables) == 0 {
		return fmt.Errorf("no refreshable tables found in %s", dataset.Name)
	}

	selection, err := ui.TableCheckbox("Select tables to refresh", tables)
	if err != nil {
		return err
	}
	if !selection.FullRefresh && len(selection.Tables) == 0 {
		fmt.Println("No tables selected.")
		return nil
	}

	fmt.Println()
	fmt.Println(infoStyle.Render("Refresh Summary"))
	fmt.Printf("  Customer:    %s\n", customerName)
	fmt.Printf("  Environment: %s\n", env)
	fmt.Printf("  Workspace:   %s\n", workspaceName)
	fmt.Printf("  Model:       %s\n", dataset.Name)
	fmt.Printf("  Tables:      %s\n", selection.Summary)
	fmt.Println()

	ok, err := ui.Confirm("Start refresh?")
	if err != nil {
		return err
	}
	if !ok {
		fmt.Println("Cancelled.")
		return nil
	}

	startTime := time.Now()

	// Run spinner while refreshing
	spinner := ui.NewSpinner("Refreshing...")
	spinner.Start()

	var refreshErr error
	var status api.RefreshStatus

	func() {
		defer spinner.Stop()

		requestID, err := client.TriggerRefresh(token, workspaceID, dataset.ID, selection.Tables)
		if err != nil {
			refreshErr = err
			return
		}

		status, err = client.WaitForRefresh(token, workspaceID, dataset.ID, requestID)
		if err != nil {
			refreshErr = err
			return
		}
	}()

	if refreshErr != nil {
		return refreshErr
	}

	duration := time.Since(startTime).Round(time.Second)

	fmt.Println()
	if status.Status == "Completed" {
		fmt.Println(successStyle.Render(fmt.Sprintf("Refresh completed successfully! (%s)", duration)))
	} else {
		msg := fmt.Sprintf("Refresh %s", status.Status)
		if len(status.Messages) > 0 {
			msg += "\n"
			for _, m := range status.Messages {
				msg += fmt.Sprintf("  • %s\n", m.Message)
			}
		}
		fmt.Println(errorStyle.Render(msg))
	}

	return nil
}

func selectCustomer(cfg config.Config) (string, config.Customer, error) {
	names := sortedCustomerNames(cfg)

	if len(names) == 1 {
		name := names[0]
		return name, cfg.Customers[name], nil
	}

	selected, err := ui.NumberMenu("Select customer", ui.MenuOptionsFromStrings(names))
	if err != nil {
		return "", config.Customer{}, err
	}
	return selected, cfg.Customers[selected], nil
}

func selectEnvironment(customer config.Customer) (string, error) {
	if len(customer.Environments) == 1 {
		return customer.Environments[0], nil
	}

	options := make([]ui.MenuOption, len(customer.Environments))
	for i, env := range customer.Environments {
		options[i] = ui.MenuOption{Label: env, Value: env}
	}

	return ui.NumberMenu("Select environment", options)
}

func selectDataset(datasets []api.Dataset) (api.Dataset, error) {
	if len(datasets) == 1 {
		return datasets[0], nil
	}

	options := make([]ui.MenuOption, len(datasets))
	dsMap := map[string]api.Dataset{}
	for i, ds := range datasets {
		options[i] = ui.MenuOption{Label: ds.Name, Value: ds.ID}
		dsMap[ds.ID] = ds
	}

	selected, err := ui.NumberMenu("Select semantic model", options)
	if err != nil {
		return api.Dataset{}, err
	}
	return dsMap[selected], nil
}
