package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/DanielAndreassen97/frefresh/internal/ui"
)

func MainMenu(configPath string) {
	fmt.Println(ui.Banner())
	fmt.Println()

	for {
		options := []ui.MenuOption{
			{Label: "Refresh tables", Value: "refresh"},
			{Label: "Add customer", Value: "add"},
			{Label: "Edit customer", Value: "edit"},
			{Label: "Remove customer", Value: "remove"},
			{Label: "List customers", Value: "list"},
			{Label: "Clear cached credentials", Value: "logout"},
			{Label: "Quit", Value: "quit"},
		}

		choice, err := ui.NumberMenu("What would you like to do?", options)
		if err != nil {
			if errors.Is(err, ui.ErrGoBack) {
				continue // Already at top level, just re-show menu
			}
			if errors.Is(err, ui.ErrQuit) {
				return
			}
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return
		}

		var cmdErr error
		switch choice {
		case "refresh":
			cmdErr = Refresh(configPath)
		case "add":
			cmdErr = Add(configPath)
		case "edit":
			cmdErr = Edit(configPath)
		case "remove":
			cmdErr = Remove(configPath)
		case "list":
			cmdErr = List(configPath)
		case "logout":
			cmdErr = Logout(configPath)
		case "quit":
			return
		}

		if cmdErr != nil {
			if errors.Is(cmdErr, ui.ErrGoBack) {
				continue
			}
			if errors.Is(cmdErr, ui.ErrQuit) {
				return
			}
			fmt.Fprintf(os.Stderr, "Error: %v\n", cmdErr)
		}
		fmt.Println()
	}
}
