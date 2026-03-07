package cmd

import (
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/DanielAndreassen97/frefresh/internal/config"
	"github.com/DanielAndreassen97/frefresh/internal/ui"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

var formTheme = func() *huh.Theme {
	t := huh.ThemeBase()
	t.Focused.Title = lipgloss.NewStyle().Foreground(ui.AccentColor).Bold(true)
	t.Focused.FocusedButton = lipgloss.NewStyle().Foreground(lipgloss.Color("0")).Background(ui.AccentColor).Padding(0, 1)
	t.Focused.BlurredButton = lipgloss.NewStyle().Foreground(ui.DimColor).Padding(0, 1)
	t.Focused.TextInput.Cursor = lipgloss.NewStyle().Foreground(ui.AccentColor)
	t.Focused.TextInput.Prompt = lipgloss.NewStyle().Foreground(ui.AccentColor)
	return t
}()

// runFormStep runs a single huh form step, converting ErrUserAborted to ErrGoBack.
func runFormStep(input *huh.Input) error {
	km := huh.NewDefaultKeyMap()
	km.Quit = key.NewBinding(key.WithKeys("ctrl+c", "esc"))
	err := huh.NewForm(
		huh.NewGroup(input),
	).WithTheme(formTheme).WithKeyMap(km).Run()
	if err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return ui.ErrGoBack
		}
		return err
	}
	return nil
}

func pickFolder() (string, error) {
	if runtime.GOOS == "darwin" {
		out, err := exec.Command("osascript", "-e",
			`POSIX path of (choose folder with prompt "Select folder with semantic models")`).Output()
		if err != nil {
			return "", fmt.Errorf("folder selection cancelled")
		}
		return strings.TrimRight(strings.TrimSpace(string(out)), "/"), nil
	}
	return "", fmt.Errorf("no picker")
}

func Add(configPath string) error {
	var name, path, pattern, envInput string

	// Step 1: Customer name
	if err := runFormStep(huh.NewInput().Title("Customer name").Value(&name)); err != nil {
		return err
	}
	if name == "" {
		return fmt.Errorf("customer name is required")
	}

	// Step 2: Path — try Finder picker, fall back to text input
	titleStyle := lipgloss.NewStyle().Foreground(ui.AccentColor).Bold(true)
	fmt.Println("  " + titleStyle.Render("Path to folder with semantic models"))
	path, err := pickFolder()
	if err != nil {
		// Finder cancelled — fall back to text input (no duplicate title)
		if err := runFormStep(huh.NewInput().Title("Enter path manually").Value(&path)); err != nil {
			return err
		}
	}
	if path == "" {
		return ui.ErrGoBack
	}

	// Step 3: Workspace pattern
	pattern = "DP - {env} - SemMod"
	if err := runFormStep(huh.NewInput().Title("Workspace pattern (use {env} for environment)").Value(&pattern)); err != nil {
		return err
	}
	if pattern == "" {
		pattern = "DP - {env} - SemMod"
	}

	// Step 4: Environments
	envInput = "DEV, TEST, PROD"
	if err := runFormStep(huh.NewInput().Title("Environments (comma-separated)").Value(&envInput)); err != nil {
		return err
	}
	if envInput == "" {
		envInput = "DEV, TEST, PROD"
	}

	envs := parseEnvs(envInput)

	addErr := config.AddCustomer(configPath, name, config.Customer{
		Path:             path,
		WorkspacePattern: pattern,
		Environments:     envs,
	})
	if addErr != nil {
		return addErr
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
