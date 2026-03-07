package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

type Customer struct {
	Path             string   `json:"path"`
	WorkspacePattern string   `json:"workspace_pattern"`
	Environments     []string `json:"environments"`
}

type Config struct {
	Customers map[string]Customer `json:"customers"`
}

func GetConfigPath() string {
	var base string
	if runtime.GOOS == "windows" {
		base = os.Getenv("APPDATA")
		if base == "" {
			base = filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Roaming")
		}
	} else {
		home, _ := os.UserHomeDir()
		base = filepath.Join(home, ".config")
	}
	return filepath.Join(base, "frefresh", "config.json")
}

func Load(path string) (Config, error) {
	cfg := Config{Customers: make(map[string]Customer)}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, err
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}
	if cfg.Customers == nil {
		cfg.Customers = make(map[string]Customer)
	}
	return cfg, nil
}

func Save(path string, cfg Config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func AddCustomer(path string, name string, customer Customer) error {
	cfg, err := Load(path)
	if err != nil {
		return err
	}
	cfg.Customers[name] = customer
	return Save(path, cfg)
}

func EditCustomer(path string, name string, customer Customer) error {
	cfg, err := Load(path)
	if err != nil {
		return err
	}
	if _, ok := cfg.Customers[name]; !ok {
		return fmt.Errorf("customer '%s' not found in config", name)
	}
	cfg.Customers[name] = customer
	return Save(path, cfg)
}

func RemoveCustomer(path string, name string) error {
	cfg, err := Load(path)
	if err != nil {
		return err
	}
	if _, ok := cfg.Customers[name]; !ok {
		return fmt.Errorf("customer '%s' not found in config", name)
	}
	delete(cfg.Customers, name)
	return Save(path, cfg)
}
