package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfigReturnsEmptyWhenNoFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Customers) != 0 {
		t.Errorf("expected 0 customers, got %d", len(cfg.Customers))
	}
}

func TestAddAndListCustomer(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	err := AddCustomer(path, "Contoso", Customer{
		Path:             "/tmp/contoso",
		WorkspacePattern: "DP - {env} - SemMod",
		Environments:     []string{"DEV", "TEST", "PROD"},
	})
	if err != nil {
		t.Fatal(err)
	}
	cfg, _ := Load(path)
	if len(cfg.Customers) != 1 {
		t.Fatalf("expected 1 customer, got %d", len(cfg.Customers))
	}
	c := cfg.Customers["Contoso"]
	if c.Path != "/tmp/contoso" {
		t.Errorf("expected /tmp/contoso, got %s", c.Path)
	}
}

func TestRemoveCustomer(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	AddCustomer(path, "Contoso", Customer{
		Path:             "/tmp/contoso",
		WorkspacePattern: "DP - {env} - SemMod",
		Environments:     []string{"DEV"},
	})
	err := RemoveCustomer(path, "Contoso")
	if err != nil {
		t.Fatal(err)
	}
	cfg, _ := Load(path)
	if len(cfg.Customers) != 0 {
		t.Errorf("expected 0 customers, got %d", len(cfg.Customers))
	}
}

func TestRemoveNonexistentCustomerReturnsError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	err := RemoveCustomer(path, "Ghost")
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestGetConfigPath(t *testing.T) {
	p := GetConfigPath()
	if p == "" {
		t.Error("expected non-empty config path")
	}
	if filepath.Base(p) != "config.json" {
		t.Errorf("expected config.json, got %s", filepath.Base(p))
	}
}

func TestEditCustomer(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	AddCustomer(path, "Contoso", Customer{
		Path:             "/tmp/contoso",
		WorkspacePattern: "DP - {env} - SemMod",
		Environments:     []string{"DEV"},
	})
	err := EditCustomer(path, "Contoso", Customer{
		Path:             "/tmp/contoso-updated",
		WorkspacePattern: "DP - {env} - SemMod",
		Environments:     []string{"DEV", "PROD"},
	})
	if err != nil {
		t.Fatal(err)
	}
	cfg, _ := Load(path)
	c := cfg.Customers["Contoso"]
	if c.Path != "/tmp/contoso-updated" {
		t.Errorf("expected /tmp/contoso-updated, got %s", c.Path)
	}
	if len(c.Environments) != 2 {
		t.Errorf("expected 2 environments, got %d", len(c.Environments))
	}
}

func TestEditNonexistentCustomerReturnsError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	err := EditCustomer(path, "Ghost", Customer{})
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestSaveCreatesDirectory(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested", "deep")
	path := filepath.Join(dir, "config.json")
	err := AddCustomer(path, "Test", Customer{Path: "/tmp"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("expected config file to be created")
	}
}
