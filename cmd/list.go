package cmd

import (
	"fmt"

	"github.com/DanielAndreassen97/frefresh/internal/config"
	"github.com/DanielAndreassen97/frefresh/internal/ui"
	"github.com/charmbracelet/lipgloss"
)

var (
	listHeaderStyle = lipgloss.NewStyle().Bold(true).Foreground(ui.AccentColor)
	listLabelStyle  = lipgloss.NewStyle().Foreground(ui.DimColor)
)

func List(configPath string) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return err
	}
	if len(cfg.Customers) == 0 {
		fmt.Println("No customers configured. Add a customer first.")
		return nil
	}

	names := sortedCustomerNames(cfg)
	for i, name := range names {
		customer := cfg.Customers[name]
		if i > 0 {
			fmt.Println()
		}
		fmt.Println(listHeaderStyle.Render(name))
		fmt.Printf("  %s %s\n", listLabelStyle.Render("Pattern:"), customer.WorkspacePattern)
		fmt.Printf("  %s %v\n", listLabelStyle.Render("Environments:"), customer.Environments)
	}
	return nil
}
