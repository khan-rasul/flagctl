package cmd

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func resetFlags() {
	updateState = ""
	updateDefault = ""
	updateDescription = ""
	updateRenameTo = ""
	createDefault = ""
	createVariants = ""
	createDescription = ""
	createDisabled = false
	rolloutSplits = ""
	rolloutComplete = ""
	rolloutAutoBalance = false
	targetOperator = ""
	targetValue = ""
	targetVariant = ""
	auditStrict = false
}

func executeCmd(args ...string) (string, error) {
	resetFlags()
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stdout = w
	os.Stderr = w

	rootCmd.SetArgs(args)
	err := rootCmd.Execute()

	_ = w.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	return buf.String(), err
}

func TestCmdSuite(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "flagctl-cmd-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	origWd, _ := os.Getwd()
	_ = os.Chdir(tempDir)
	defer func() { _ = os.Chdir(origWd) }()

	// 1. Version
	out, err := executeCmd("version")
	require.NoError(t, err)
	assert.Contains(t, out, "0.0.1")

	// 2. Init
	out, err = executeCmd("init", "-f", "json", "-l", "typescript")
	require.NoError(t, err, "init failed: %s", out)
	assert.FileExists(t, filepath.Join(tempDir, "flags.json"))

	// Init invalid format
	_, err = executeCmd("init", "-f", "invalid-fmt")
	assert.Error(t, err)

	// 3. Create
	_, err = executeCmd("create")
	assert.Error(t, err)

	out, err = executeCmd("create", "-k", "my-feature", "-t", "boolean", "-d", "on")
	require.NoError(t, err, "create failed: %s", out)

	// Create with string variants
	out, err = executeCmd("create", "-k", "theme-flag", "-t", "string", "-v", "dark=dark-v1,light=light-v1", "-d", "dark")
	require.NoError(t, err)

	// Create with number variants
	out, err = executeCmd("create", "-k", "limit-flag", "-t", "number", "-v", "low=10,high=100", "-d", "low")
	require.NoError(t, err)

	// Create with invalid default
	_, err = executeCmd("create", "-k", "bad-default", "-t", "boolean", "-d", "non-existent")
	assert.Error(t, err)

	// 4. List
	out, err = executeCmd("list")
	require.NoError(t, err)
	assert.Contains(t, out, "my-feature")

	// 5. Update
	_, err = executeCmd("update")
	assert.Error(t, err)

	out, err = executeCmd("update", "-k", "my-feature", "-s", "DISABLED")
	require.NoError(t, err)

	// Update invalid state
	_, err = executeCmd("update", "-k", "my-feature", "-s", "INVALID_STATE")
	assert.Error(t, err)

	// Update invalid default variant
	_, err = executeCmd("update", "-k", "my-feature", "-d", "non-existent-variant")
	assert.Error(t, err)

	// Update description
	out, err = executeCmd("update", "-k", "my-feature", "--description", "Updated desc")
	require.NoError(t, err)

	// 6. Rollout
	_, err = executeCmd("rollout")
	assert.Error(t, err)

	_, _ = executeCmd("update", "-k", "my-feature", "-s", "ENABLED")
	out, err = executeCmd("rollout", "-k", "my-feature", "-s", "on=50,off=50")
	require.NoError(t, err)

	// Rollout complete
	out, err = executeCmd("rollout", "-k", "my-feature", "--complete", "on")
	require.NoError(t, err)

	// Rollout invalid percentage sum
	_, err = executeCmd("rollout", "-k", "my-feature", "-s", "on=10,off=10")
	assert.Error(t, err)

	// 7. Target
	_, err = executeCmd("target")
	assert.Error(t, err)

	out, err = executeCmd("target", "-k", "my-feature", "-a", "email", "-v", "@company.com", "--variant", "on")
	require.NoError(t, err)

	// Target invalid variant
	_, err = executeCmd("target", "-k", "my-feature", "-a", "email", "-v", "@company.com", "--variant", "invalid")
	assert.Error(t, err)

	// 8. Validate
	out, err = executeCmd("validate")
	require.NoError(t, err)

	// 9. Generate
	out, err = executeCmd("generate", "-l", "go", "-o", "pkg/flags/flags.gen.go")
	require.NoError(t, err)

	// 10. Deprecate
	_, err = executeCmd("deprecate")
	assert.Error(t, err)

	out, err = executeCmd("deprecate", "-k", "my-feature", "-r", "old feature")
	require.NoError(t, err)

	// Deprecate already deprecated flag
	out, err = executeCmd("deprecate", "-k", "my-feature")
	require.NoError(t, err)

	// 11. Audit
	out, err = executeCmd("audit")
	require.NoError(t, err)

	// Audit strict
	_, err = executeCmd("audit", "--strict")
	assert.Error(t, err)

	// 12. Undeprecate
	_, err = executeCmd("undeprecate")
	assert.Error(t, err)

	out, err = executeCmd("undeprecate", "-k", "my-feature")
	require.NoError(t, err)

	// 13. Delete
	_, err = executeCmd("delete")
	assert.Error(t, err)

	out, err = executeCmd("delete", "-k", "my-feature", "-f")
	require.NoError(t, err)

	// Delete non-existent flag
	_, err = executeCmd("delete", "-k", "non-existent-flag", "-f")
	assert.Error(t, err)
}
