package cmd

import "fmt"

func Help() {
	fmt.Println(`Usage: frefresh [command]

Commands:
  refresh   Select a customer, model, and tables to refresh via Power BI API
  add       Add a new customer configuration
  edit      Edit an existing customer's workspace pattern or environments
  remove    Remove a customer configuration
  list      Show all configured customers
  logout    Clear cached OAuth tokens from the OS keychain
  help      Show this help message
  version   Print the current version

Run without arguments to use the interactive menu.`)
}
