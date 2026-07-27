package workspace

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindRootAndWorkspace(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "flagctl-workspace-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	subDir := filepath.Join(tempDir, "src", "sub")
	require.NoError(t, os.MkdirAll(subDir, 0755))

	// Should fail before init
	_, err = FindRoot(subDir)
	assert.Error(t, err)

	// Save workspace config in tempDir
	wsCfg := &WorkspaceConfig{
		Version:       "1",
		ConfigPath:    "flags.json",
		Format:        "json",
		SchemaVersion: "v0",
	}
	err = SaveWorkspace(tempDir, wsCfg)
	require.NoError(t, err)

	// FindRoot should now discover tempDir from subDir
	foundRoot, err := FindRoot(subDir)
	require.NoError(t, err)
	assert.Equal(t, tempDir, foundRoot)

	// LoadWorkspace should retrieve saved config
	loaded, err := LoadWorkspace(tempDir)
	require.NoError(t, err)
	assert.Equal(t, "flags.json", loaded.ConfigPath)
	assert.Equal(t, "v0", loaded.SchemaVersion)
}

func TestLoadWorkspaceDefaultFallback(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "flagctl-fallback-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Touch flags.yaml
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "flags.yaml"), []byte("flags: {}"), 0644))

	loaded, err := LoadWorkspace(tempDir)
	require.NoError(t, err)
	assert.Equal(t, "flags.yaml", loaded.ConfigPath)
	assert.Equal(t, "yaml", loaded.Format)

	// FindRoot finding flags.json directly
	jsonDir := filepath.Join(tempDir, "jsondir")
	_ = os.MkdirAll(jsonDir, 0755)
	_ = os.WriteFile(filepath.Join(jsonDir, "flags.json"), []byte("{}"), 0644)
	r, err := FindRoot(jsonDir)
	require.NoError(t, err)
	assert.Equal(t, jsonDir, r)
}

func TestLoadWorkspaceErrors(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "flagctl-ws-err-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Invalid .flagctl.json
	_ = os.WriteFile(filepath.Join(tempDir, WorkspaceConfigFile), []byte("{invalid json"), 0644)
	_, err = LoadWorkspace(tempDir)
	assert.Error(t, err)
}
