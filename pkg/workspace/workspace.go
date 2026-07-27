package workspace

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const WorkspaceConfigFile = ".flagctl.json"

type CodegenConfig struct {
	Enabled    bool   `json:"enabled"`
	Language   string `json:"language"`
	OutputPath string `json:"outputPath"`
}

type WorkspaceConfig struct {
	Version       string         `json:"version"`
	ConfigPath    string         `json:"configPath"`
	Format        string         `json:"format"`
	SchemaURL     string         `json:"schema,omitempty"`
	SchemaVersion string         `json:"schemaVersion,omitempty"`
	Codegen       *CodegenConfig `json:"codegen,omitempty"`
}

// FindRoot walks up from startDir to locate the directory containing .flagctl.json or flags.json/flags.yaml.
func FindRoot(startDir string) (string, error) {
	abs, err := filepath.Abs(startDir)
	if err != nil {
		return "", err
	}

	curr := abs
	for {
		// Check for .flagctl.json
		if _, err := os.Stat(filepath.Join(curr, WorkspaceConfigFile)); err == nil {
			return curr, nil
		}
		// Check for flags.json or flags.yaml
		if _, err := os.Stat(filepath.Join(curr, "flags.json")); err == nil {
			return curr, nil
		}
		if _, err := os.Stat(filepath.Join(curr, "flags.yaml")); err == nil {
			return curr, nil
		}

		parent := filepath.Dir(curr)
		if parent == curr {
			break
		}
		curr = parent
	}

	return "", errors.New("flagctl workspace not initialized (no .flagctl.json or flags.json found). Run 'flagctl init' first")
}

// LoadWorkspace loads .flagctl.json from root. If absent, returns default config.
func LoadWorkspace(root string) (*WorkspaceConfig, error) {
	cfgPath := filepath.Join(root, WorkspaceConfigFile)
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// Infer config file path
			format := "json"
			configPath := "flags.json"
			if _, errYaml := os.Stat(filepath.Join(root, "flags.yaml")); errYaml == nil {
				format = "yaml"
				configPath = "flags.yaml"
			}
			return &WorkspaceConfig{
				Version:       "1",
				ConfigPath:    configPath,
				Format:        format,
				SchemaVersion: "v0",
			}, nil
		}
		return nil, fmt.Errorf("failed to read %s: %w", WorkspaceConfigFile, err)
	}

	var cfg WorkspaceConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("invalid %s format: %w", WorkspaceConfigFile, err)
	}

	if cfg.SchemaVersion == "" {
		cfg.SchemaVersion = "v0"
	}

	return &cfg, nil
}

// SaveWorkspace writes WorkspaceConfig to .flagctl.json in root.
func SaveWorkspace(root string, cfg *WorkspaceConfig) error {
	cfgPath := filepath.Join(root, WorkspaceConfigFile)
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(cfgPath, data, 0644)
}
