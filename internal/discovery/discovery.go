package discovery

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type SemanticModel struct {
	Name      string
	Path      string
	LogicalID string
}

type platformFile struct {
	Metadata struct {
		Type        string `json:"type"`
		DisplayName string `json:"displayName"`
	} `json:"metadata"`
	Config struct {
		LogicalID string `json:"logicalId"`
	} `json:"config"`
}

func DiscoverModels(basePath string) ([]SemanticModel, error) {
	var models []SemanticModel

	err := filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || info.Name() != ".platform" {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		var pf platformFile
		if json.Unmarshal(data, &pf) != nil {
			return nil
		}
		if pf.Metadata.Type != "SemanticModel" {
			return nil
		}
		name := pf.Metadata.DisplayName
		if name == "" {
			name = filepath.Base(filepath.Dir(path))
		}
		models = append(models, SemanticModel{
			Name:      name,
			Path:      filepath.Dir(path),
			LogicalID: pf.Config.LogicalID,
		})
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(models, func(i, j int) bool {
		return models[i].Name < models[j].Name
	})
	return models, nil
}

func isNonRefreshable(tmdlPath string) bool {
	f, err := os.Open(tmdlPath)
	if err != nil {
		return false
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "partition ") && strings.Contains(line, "= calculated") {
			return true
		}
		if line == "calculationGroup" || strings.HasPrefix(line, "calculationGroup\t") {
			return true
		}
	}
	return false
}

func DiscoverTables(model SemanticModel) ([]string, error) {
	tablesDir := filepath.Join(model.Path, "definition", "tables")
	entries, err := os.ReadDir(tablesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var tables []string
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".tmdl") {
			continue
		}
		fullPath := filepath.Join(tablesDir, entry.Name())
		if isNonRefreshable(fullPath) {
			continue
		}
		name := strings.TrimSuffix(entry.Name(), ".tmdl")
		tables = append(tables, name)
	}
	sort.Strings(tables)
	return tables, nil
}
