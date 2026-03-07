package cmd

import (
	"fmt"

	"github.com/DanielAndreassen97/frefresh/internal/api"
	"github.com/DanielAndreassen97/frefresh/internal/config"
)

func Logout(configPath string) {
	// Clear old global tokens (pre per-customer migration)
	api.ClearAllTokens()

	// Clear tokens for each configured customer
	cfg, err := config.Load(configPath)
	if err == nil {
		for name := range cfg.Customers {
			api.ClearCachedTokens(name)
		}
	}
	fmt.Println("Cached tokens cleared.")
}
