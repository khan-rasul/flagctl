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
	auditStrict = false

	launchAddKey = ""
	launchAddPercent = 0
	launchAddVariant = ""
	launchAddSplits = ""
	launchAddBucketBy = "userId"
	launchAddAttribute = ""
	launchAddValue = ""
	launchListKey = ""
	launchRampKey = ""
	launchRampPercent = 0
	launchRampVariant = ""
	launchRampIndex = 0
	launchRemoveKey = ""
	launchRemoveIndex = 0

	targetAddKey = ""
	targetAddAttribute = ""
	targetAddOperator = "=="
	targetAddValue = ""
	targetAddVariant = ""
	targetAddTop = false
	targetListKey = ""
	targetRemoveKey = ""
	targetRemoveIndex = 0
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

	// 3. Create
	out, err = executeCmd("create", "-k", "my-feature", "-t", "boolean", "-d", "on")
	require.NoError(t, err, "create failed: %s", out)

	// 4. Target Add (Allowlist & Denylist)
	out, err = executeCmd("target", "add", "-k", "my-feature", "-a", "email", "-o", "endsWith", "-v", "@company.com", "--variant", "on")
	require.NoError(t, err)

	out, err = executeCmd("target", "add", "-k", "my-feature", "-a", "email", "-o", "endsWith", "-v", "@competitor.com", "--variant", "off", "--top")
	require.NoError(t, err)

	// 5. Target List
	out, err = executeCmd("target", "list", "-k", "my-feature")
	require.NoError(t, err)
	assert.Contains(t, out, "@competitor.com")
	assert.Contains(t, out, "@company.com")

	// 6. Launch Add & Ramp
	out, err = executeCmd("launch", "add", "-k", "my-feature", "-p", "20", "-v", "on")
	require.NoError(t, err)

	out, err = executeCmd("launch", "ramp", "-k", "my-feature", "-p", "50", "-v", "on")
	require.NoError(t, err)

	out, err = executeCmd("launch", "list", "-k", "my-feature")
	require.NoError(t, err)
	assert.Contains(t, out, "fractional rollout")

	// 7. Launch Ramp Complete (100%)
	out, err = executeCmd("launch", "ramp", "-k", "my-feature", "-p", "100", "-v", "on")
	require.NoError(t, err)

	// Re-add target rule & remove rule by index
	_, _ = executeCmd("target", "add", "-k", "my-feature", "-a", "email", "-v", "@company.com", "--variant", "on")
	out, err = executeCmd("target", "remove", "-k", "my-feature", "-i", "1")
	require.NoError(t, err)

	// 8. Validate
	out, err = executeCmd("validate")
	require.NoError(t, err)

	// 9. Deprecate
	out, err = executeCmd("deprecate", "-k", "my-feature", "-r", "old feature")
	require.NoError(t, err)

	// 10. Audit
	out, err = executeCmd("audit")
	require.NoError(t, err)

	// 11. Undeprecate
	out, err = executeCmd("undeprecate", "-k", "my-feature")
	require.NoError(t, err)

	// 12. Delete
	out, err = executeCmd("delete", "-k", "my-feature", "-f")
	require.NoError(t, err)
}
