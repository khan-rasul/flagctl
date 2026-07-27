package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFlagConfigCRUD(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "flagctl-config-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	jsonPath := filepath.Join(tempDir, "flags.json")
	yamlPath := filepath.Join(tempDir, "flags.yaml")

	cfg := NewFlagConfig("my-service")
	assert.Equal(t, "my-service", cfg.Metadata["flagSetId"])

	flag := &Flag{
		State:          "ENABLED",
		DefaultVariant: "on",
		Variants: map[string]interface{}{
			"on":  true,
			"off": false,
		},
	}

	err = cfg.AddFlag("test-feature", flag)
	require.NoError(t, err)

	// Duplicate add should error
	err = cfg.AddFlag("test-feature", flag)
	assert.Error(t, err)

	// Save & Load JSON
	err = SaveConfig(jsonPath, "json", cfg)
	require.NoError(t, err)

	loadedJSON, err := LoadConfig(jsonPath, "json")
	require.NoError(t, err)
	fJSON, ok := loadedJSON.GetFlag("test-feature")
	require.True(t, ok)
	assert.Equal(t, "on", fJSON.DefaultVariant)

	// Save & Load YAML
	err = SaveConfig(yamlPath, "yaml", cfg)
	require.NoError(t, err)

	loadedYAML, err := LoadConfig(yamlPath, "yaml")
	require.NoError(t, err)
	fYAML, ok := loadedYAML.GetFlag("test-feature")
	require.True(t, ok)
	assert.Equal(t, "ENABLED", fYAML.State)

	// Deprecate & Undeprecate
	assert.False(t, loadedJSON.IsDeprecated(fJSON))
	err = loadedJSON.DeprecateFlag("test-feature", "Feature complete")
	require.NoError(t, err)
	assert.True(t, loadedJSON.IsDeprecated(fJSON))

	err = loadedJSON.UndeprecateFlag("test-feature")
	require.NoError(t, err)
	assert.False(t, loadedJSON.IsDeprecated(fJSON))

	// Undeprecate non-existent
	err = loadedJSON.UndeprecateFlag("non-existent")
	assert.Error(t, err)

	// Deprecate non-existent
	err = loadedJSON.DeprecateFlag("non-existent", "reason")
	assert.Error(t, err)

	// Delete
	err = loadedJSON.DeleteFlag("test-feature")
	require.NoError(t, err)
	_, ok = loadedJSON.GetFlag("test-feature")
	assert.False(t, ok)

	// Delete non-existent
	err = loadedJSON.DeleteFlag("non-existent")
	assert.Error(t, err)
}

func TestLoadConfigErrors(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "flagctl-config-err-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// File not found
	_, err = LoadConfig(filepath.Join(tempDir, "missing.json"), "json")
	assert.Error(t, err)

	// Invalid JSON
	badJSON := filepath.Join(tempDir, "bad.json")
	_ = os.WriteFile(badJSON, []byte("{invalid json"), 0644)
	_, err = LoadConfig(badJSON, "json")
	assert.Error(t, err)

	// Invalid YAML
	badYAML := filepath.Join(tempDir, "bad.yaml")
	_ = os.WriteFile(badYAML, []byte("invalid: : yaml"), 0644)
	_, err = LoadConfig(badYAML, "yaml")
	assert.Error(t, err)
}
