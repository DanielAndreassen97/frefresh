package cmd

import (
	"fmt"
	"strings"

	"github.com/DanielAndreassen97/frefresh/internal/config"
	"github.com/DanielAndreassen97/frefresh/internal/ui"
	"github.com/charmbracelet/huh"
)

func Edit(configPath string) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return err
	}
	if len(cfg.Customers) == 0 {
		fmt.Println("No customers configured.")
		return nil
	}

	names := sortedCustomerNames(cfg)
	options := make([]ui.MenuOption, len(names))
	for i, n := range names {
		options[i] = ui.MenuOption{Label: n, Value: n}
	}

	selected, err := ui.NumberMenu("Select customer to edit", options)
	if err != nil {
		return err
	}

	customer := cfg.Customers[selected]
	path := customer.Path
	pattern := customer.WorkspacePattern
	envInput := strings.Join(customer.Environments, ", ")

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Title("Path to folder with semantic models").Value(&path),
			huh.NewInput().Title("Workspace pattern (use {env} for environment)").Value(&pattern),
			huh.NewInput().Title("Environments (comma-separated)").Value(&envInput),
		),
	)

	if err := form.Run(); err != nil {
		return err
	}

	err = config.EditCustomer(configPath, selected, config.Customer{
		Path:             path,
		WorkspacePattern: pattern,
		Environments:     parseEnvs(envInput),
	})
	if err != nil {
		return err
	}
	fmt.Printf("Customer '%s' updated.\n", selected)
	return nil
}
