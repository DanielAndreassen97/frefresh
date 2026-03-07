package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/DanielAndreassen97/frefresh/internal/api"
	"github.com/DanielAndreassen97/frefresh/internal/config"
	"github.com/DanielAndreassen97/frefresh/internal/discovery"
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

	infoStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
)

// APIClient abstracts the Power BI API calls for testability and demo mode.
type APIClient interface {
	GetAccessToken() (string, error)
	GetWorkspaceID(token, workspaceName string) (string, error)
	GetDatasetID(token, workspaceID, datasetName string) (string, error)
	TriggerRefresh(token, workspaceID, datasetID string, tables []string) (string, error)
	WaitForRefresh(token, workspaceID, datasetID, requestID string) (api.RefreshStatus, error)
}

// RealAPIClient calls the actual Power BI REST API.
type RealAPIClient struct{}

func (RealAPIClient) GetAccessToken() (string, error) {
	return api.GetAccessToken()
}
func (RealAPIClient) GetWorkspaceID(token, workspaceName string) (string, error) {
	return api.GetWorkspaceID(token, workspaceName)
}
func (RealAPIClient) GetDatasetID(token, workspaceID, datasetName string) (string, error) {
	return api.GetDatasetID(token, workspaceID, datasetName)
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
		fmt.Println("No customers configured. Use 'frefresh add' to add one.")
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

	models, err := discovery.DiscoverModels(customer.Path)
	if err != nil {
		return fmt.Errorf("failed to discover models: %w", err)
	}
	if len(models) == 0 {
		return fmt.Errorf("no semantic models found in %s", customer.Path)
	}

	model, err := selectModel(models)
	if err != nil {
		return err
	}

	tables, err := discovery.DiscoverTables(model)
	if err != nil {
		return fmt.Errorf("failed to discover tables: %w", err)
	}
	if len(tables) == 0 {
		return fmt.Errorf("no refreshable tables found in %s", model.Name)
	}

	selectedTables, err := ui.TableCheckbox("Select tables to refresh", tables)
	if err != nil {
		return err
	}
	if len(selectedTables) == 0 {
		fmt.Println("No tables selected.")
		return nil
	}

	workspaceName := strings.ReplaceAll(customer.WorkspacePattern, "{env}", env)

	fmt.Println()
	fmt.Println(infoStyle.Render("Refresh Summary"))
	fmt.Printf("  Customer:    %s\n", customerName)
	fmt.Printf("  Environment: %s\n", env)
	fmt.Printf("  Workspace:   %s\n", workspaceName)
	fmt.Printf("  Model:       %s\n", model.Name)
	fmt.Printf("  Tables:      %s\n", strings.Join(selectedTables, ", "))
	fmt.Println()

	ok, err := ui.Confirm("Start refresh?")
	if err != nil {
		return err
	}
	if !ok {
		fmt.Println("Cancelled.")
		return nil
	}

	fmt.Println()
	fmt.Println(infoStyle.Render("Opening browser for authentication..."))
	token, err := client.GetAccessToken()
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}
	fmt.Println(infoStyle.Render("Authenticated."))

	fmt.Println(infoStyle.Render("Looking up workspace and dataset..."))
	workspaceID, err := client.GetWorkspaceID(token, workspaceName)
	if err != nil {
		return err
	}
	datasetID, err := client.GetDatasetID(token, workspaceID, model.Name)
	if err != nil {
		return err
	}

	fmt.Println(infoStyle.Render("Triggering refresh..."))
	requestID, err := client.TriggerRefresh(token, workspaceID, datasetID, selectedTables)
	if err != nil {
		return err
	}
	fmt.Printf(infoStyle.Render("Refresh triggered (request ID: %s). Waiting...")+"\n", requestID)

	status, err := client.WaitForRefresh(token, workspaceID, datasetID, requestID)
	if err != nil {
		return err
	}

	fmt.Println()
	if status.Status == "Completed" {
		fmt.Println(successStyle.Render("Refresh completed successfully!"))
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

	options := make([]ui.MenuOption, len(names))
	for i, n := range names {
		options[i] = ui.MenuOption{Label: n, Value: n}
	}

	selected, err := ui.NumberMenu("Select customer", options)
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

func selectModel(models []discovery.SemanticModel) (discovery.SemanticModel, error) {
	if len(models) == 1 {
		return models[0], nil
	}

	options := make([]ui.MenuOption, len(models))
	modelMap := map[string]discovery.SemanticModel{}
	for i, m := range models {
		options[i] = ui.MenuOption{Label: m.Name, Value: m.Name}
		modelMap[m.Name] = m
	}

	selected, err := ui.NumberMenu("Select semantic model", options)
	if err != nil {
		return discovery.SemanticModel{}, err
	}
	return modelMap[selected], nil
}
