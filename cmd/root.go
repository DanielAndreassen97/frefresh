package cmd

import (
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
			{Label: "Quit", Value: "quit"},
		}

		choice, err := ui.NumberMenu("What would you like to do?", options)
		if err != nil {
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
		case "quit":
			return
		}

		if cmdErr != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", cmdErr)
		}
		fmt.Println()
	}
}
