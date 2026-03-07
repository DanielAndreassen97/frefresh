package cmd

import (
	"fmt"

	"github.com/DanielAndreassen97/frefresh/internal/config"
	"github.com/charmbracelet/lipgloss"
)

var (
	listHeaderStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#e8712a"))
	listLabelStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
)

func List(configPath string) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return err
	}
	if len(cfg.Customers) == 0 {
		fmt.Println("No customers configured. Use 'frefresh add' to add one.")
		return nil
	}

	names := sortedCustomerNames(cfg)
	for i, name := range names {
		customer := cfg.Customers[name]
		if i > 0 {
			fmt.Println()
		}
		fmt.Println(listHeaderStyle.Render(name))
		fmt.Printf("  %s %s\n", listLabelStyle.Render("Path:"), customer.Path)
		fmt.Printf("  %s %s\n", listLabelStyle.Render("Pattern:"), customer.WorkspacePattern)
		fmt.Printf("  %s %v\n", listLabelStyle.Render("Environments:"), customer.Environments)
	}
	return nil
}
