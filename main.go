package main

import (
	"fmt"
	"os"

	"github.com/DanielAndreassen97/frefresh/cmd"
	"github.com/DanielAndreassen97/frefresh/internal/config"
	"github.com/DanielAndreassen97/frefresh/internal/demo"
	"github.com/DanielAndreassen97/frefresh/internal/ui"
)

var version = "dev"

func main() {
	ui.Version = version
	configPath := config.GetConfigPath()
	args := os.Args[1:]

	// Check for --demo flag
	isDemo := false
	filtered := args[:0]
	for _, a := range args {
		if a == "--demo" {
			isDemo = true
		} else {
			filtered = append(filtered, a)
		}
	}
	args = filtered

	if isDemo {
		demoPath, cleanup, err := demo.SetupFixtures()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to setup demo: %v\n", err)
			os.Exit(1)
		}
		defer cleanup()
		configPath = demoPath
		cmd.DefaultAPI = demo.MockAPIClient{}
	}

	if len(args) == 0 {
		cmd.MainMenu(configPath)
		return
	}

	var err error
	switch args[0] {
	case "add":
		err = cmd.Add(configPath)
	case "remove":
		err = cmd.Remove(configPath)
	case "edit":
		err = cmd.Edit(configPath)
	case "list":
		err = cmd.List(configPath)
	case "refresh":
		err = cmd.Refresh(configPath)
	case "logout":
		cmd.Logout(configPath)
	case "version", "--version", "-v":
		fmt.Printf("frefresh %s\n", version)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", args[0])
		os.Exit(1)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
