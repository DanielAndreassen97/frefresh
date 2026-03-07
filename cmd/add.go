package cmd

import (
	"fmt"
	"strings"

	"github.com/DanielAndreassen97/frefresh/internal/config"
	"github.com/charmbracelet/huh"
)

func Add(configPath string) error {
	var name, path, pattern, envInput string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Title("Customer name").Value(&name),
			huh.NewInput().Title("Path to folder with semantic models").Value(&path),
			huh.NewInput().Title("Workspace pattern (use {env} for environment)").Value(&pattern).Placeholder("DP - {env} - SemMod"),
			huh.NewInput().Title("Environments (comma-separated)").Value(&envInput).Placeholder("DEV, TEST, PROD"),
		),
	)

	if err := form.Run(); err != nil {
		return err
	}

	if name == "" {
		return fmt.Errorf("customer name is required")
	}
	if path == "" {
		return fmt.Errorf("path is required")
	}
	if pattern == "" {
		pattern = "DP - {env} - SemMod"
	}
	if envInput == "" {
		envInput = "DEV, TEST, PROD"
	}

	envs := parseEnvs(envInput)

	err := config.AddCustomer(configPath, name, config.Customer{
		Path:             path,
		WorkspacePattern: pattern,
		Environments:     envs,
	})
	if err != nil {
		return err
	}
	fmt.Printf("Customer '%s' added.\n", name)
	return nil
}

func parseEnvs(input string) []string {
	var envs []string
	for _, e := range strings.Split(input, ",") {
		e = strings.TrimSpace(e)
		if e != "" {
			envs = append(envs, e)
		}
	}
	return envs
}
