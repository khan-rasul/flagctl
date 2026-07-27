package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Flag struct {
	State          string                 `json:"state" yaml:"state"`
	DefaultVariant string                 `json:"defaultVariant" yaml:"defaultVariant"`
	Variants       map[string]interface{} `json:"variants" yaml:"variants"`
	Targeting      interface{}            `json:"targeting,omitempty" yaml:"targeting,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

type FlagConfig struct {
	Schema     string                 `json:"$schema,omitempty" yaml:"$schema,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	Evaluators map[string]interface{} `json:"$evaluators,omitempty" yaml:"$evaluators,omitempty"`
	Flags      map[string]*Flag       `json:"flags" yaml:"flags"`
}

func NewFlagConfig(flagSetID string) *FlagConfig {
	return &FlagConfig{
		Schema: "https://flagd.dev/schema/v0/flags.json",
		Metadata: map[string]interface{}{
			"flagSetId": flagSetID,
			"version":   "1.0.0",
		},
		Evaluators: make(map[string]interface{}),
		Flags:      make(map[string]*Flag),
	}
}

func LoadConfig(path string, format string) (*FlagConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read flag config '%s': %w", path, err)
	}

	var cfg FlagConfig
	format = strings.ToLower(format)

	if format == "yaml" || format == "yml" || strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml") {
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("invalid YAML in '%s': %w", path, err)
		}
	} else {
		if err := json.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("invalid JSON in '%s': %w", path, err)
		}
	}

	if cfg.Flags == nil {
		cfg.Flags = make(map[string]*Flag)
	}

	return &cfg, nil
}

func SaveConfig(path string, format string, cfg *FlagConfig) error {
	format = strings.ToLower(format)
	var data []byte
	var err error

	if format == "yaml" || format == "yml" || strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml") {
		data, err = yaml.Marshal(cfg)
	} else {
		data, err = json.MarshalIndent(cfg, "", "  ")
	}

	if err != nil {
		return fmt.Errorf("failed to serialize flag config: %w", err)
	}

	dir := filepath.Dir(path)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	return os.WriteFile(path, data, 0644)
}

func (c *FlagConfig) GetFlag(key string) (*Flag, bool) {
	flag, ok := c.Flags[key]
	return flag, ok
}

func (c *FlagConfig) AddFlag(key string, flag *Flag) error {
	if _, exists := c.Flags[key]; exists {
		return fmt.Errorf("flag '%s' already exists in configuration", key)
	}
	c.Flags[key] = flag
	return nil
}

func (c *FlagConfig) IsDeprecated(flag *Flag) bool {
	if flag.Metadata == nil {
		return false
	}
	dep, ok := flag.Metadata["deprecated"]
	if !ok {
		return false
	}
	if b, ok := dep.(bool); ok {
		return b
	}
	return false
}

func (c *FlagConfig) DeprecateFlag(key string, reason string) error {
	flag, ok := c.Flags[key]
	if !ok {
		return fmt.Errorf("flag '%s' not found", key)
	}

	if flag.Metadata == nil {
		flag.Metadata = make(map[string]interface{})
	}

	flag.Metadata["deprecated"] = true
	flag.Metadata["deprecatedAt"] = time.Now().Format("2006-01-02")
	if reason != "" {
		flag.Metadata["deprecationReason"] = reason
	}

	return nil
}

func (c *FlagConfig) UndeprecateFlag(key string) error {
	flag, ok := c.Flags[key]
	if !ok {
		return fmt.Errorf("flag '%s' not found", key)
	}

	if flag.Metadata != nil {
		delete(flag.Metadata, "deprecated")
		delete(flag.Metadata, "deprecatedAt")
		delete(flag.Metadata, "deprecationReason")
	}

	return nil
}

func (c *FlagConfig) DeleteFlag(key string) error {
	if _, ok := c.Flags[key]; !ok {
		return fmt.Errorf("flag '%s' not found", key)
	}
	delete(c.Flags, key)
	return nil
}
