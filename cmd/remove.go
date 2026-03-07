package cmd

import (
	"fmt"
	"sort"

	"github.com/DanielAndreassen97/frefresh/internal/config"
	"github.com/DanielAndreassen97/frefresh/internal/ui"
)

func Remove(configPath string) error {
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

	selected, err := ui.NumberMenu("Select customer to remove", options)
	if err != nil {
		return err
	}

	ok, err := ui.Confirm(fmt.Sprintf("Remove '%s'?", selected))
	if err != nil {
		return err
	}
	if !ok {
		fmt.Println("Cancelled.")
		return nil
	}

	if err := config.RemoveCustomer(configPath, selected); err != nil {
		return err
	}
	fmt.Printf("Customer '%s' removed.\n", selected)
	return nil
}

func sortedCustomerNames(cfg config.Config) []string {
	names := make([]string, 0, len(cfg.Customers))
	for name := range cfg.Customers {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
