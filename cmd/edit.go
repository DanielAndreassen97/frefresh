package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/DanielAndreassen97/frefresh/internal/config"
	"github.com/DanielAndreassen97/frefresh/internal/ui"
	"github.com/charmbracelet/bubbles/key"
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
	options := ui.MenuOptionsFromStrings(names)

	selected, err := ui.NumberMenu("Select customer to edit", options)
	if err != nil {
		return err
	}

	customer := cfg.Customers[selected]
	pattern := customer.WorkspacePattern
	envInput := strings.Join(customer.Environments, ", ")

	km := huh.NewDefaultKeyMap()
	km.Quit = key.NewBinding(key.WithKeys("ctrl+c", "esc"))

	err = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Title("Workspace pattern (use {env} for environment)").Value(&pattern),
			huh.NewInput().Title("Environments (comma-separated)").Value(&envInput),
		),
	).WithTheme(formTheme).WithKeyMap(km).Run()
	if err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return ui.ErrGoBack
		}
		return err
	}

	err = config.EditCustomer(configPath, selected, config.Customer{
		WorkspacePattern: pattern,
		Environments:     parseEnvs(envInput),
	})
	if err != nil {
		return err
	}
	fmt.Printf("Customer '%s' updated.\n", selected)
	return nil
}
