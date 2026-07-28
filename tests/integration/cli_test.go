package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func runFlagctl(dir string, args ...string) (string, error) {
	binaryPath, err := filepath.Abs("../../flagctl")
	if err != nil {
		return "", err
	}

	cmd := exec.Command(binaryPath, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func TestCLIEndToEndWorkflow(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "flagctl-integration-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// 1. Init
	out, err := runFlagctl(tempDir, "init", "--flag-set-id", "integration-test", "-l", "typescript")
	require.NoError(t, err, "init failed: %s", out)
	assert.Contains(t, out, "Initialized empty flagd configuration")

	// 2. Create Boolean Flag
	out, err = runFlagctl(tempDir, "create", "--key", "new-checkout-flow", "--type", "boolean", "--default", "on", "--description", "New checkout flow")
	require.NoError(t, err, "create failed: %s", out)
	assert.Contains(t, out, "Successfully created flag 'new-checkout-flow'")

	// 3. Launch Add (20%)
	out, err = runFlagctl(tempDir, "launch", "add", "--key", "new-checkout-flow", "--percent", "20", "--variant", "on")
	require.NoError(t, err, "launch add failed: %s", out)
	assert.Contains(t, out, "Added global launch ramp")

	// 4. Launch Ramp (50%)
	out, err = runFlagctl(tempDir, "launch", "ramp", "--key", "new-checkout-flow", "--percent", "50", "--variant", "on")
	require.NoError(t, err, "launch ramp failed: %s", out)
	assert.Contains(t, out, "Ramped launch percentage")

	// 5. Target Add (Allowlist & Denylist)
	out, err = runFlagctl(tempDir, "target", "add", "--key", "new-checkout-flow", "--attribute", "email", "-o", "endsWith", "-v", "@company.com", "--variant", "on")
	require.NoError(t, err, "target add failed: %s", out)

	out, err = runFlagctl(tempDir, "target", "add", "--key", "new-checkout-flow", "--attribute", "email", "-o", "endsWith", "-v", "@competitor.com", "--variant", "off", "--top")
	require.NoError(t, err)

	// 6. Target List
	out, err = runFlagctl(tempDir, "target", "list", "--key", "new-checkout-flow")
	require.NoError(t, err)
	assert.Contains(t, out, "@competitor.com")
	assert.Contains(t, out, "@company.com")

	// 7. Target Remove
	out, err = runFlagctl(tempDir, "target", "remove", "--key", "new-checkout-flow", "--index", "1")
	require.NoError(t, err)

	// 8. Validate Schema
	out, err = runFlagctl(tempDir, "validate")
	require.NoError(t, err, "validate failed: %s", out)
	assert.Contains(t, out, "is valid according to flagd schema")

	// 9. List Flags
	out, err = runFlagctl(tempDir, "list")
	require.NoError(t, err, "list failed: %s", out)
	assert.Contains(t, out, "new-checkout-flow")

	// 10. Soft Deprecate
	out, err = runFlagctl(tempDir, "deprecate", "--key", "new-checkout-flow", "--reason", "Feature completed")
	require.NoError(t, err, "deprecate failed: %s", out)

	// Deprecated flag should block launch ramp
	out, err = runFlagctl(tempDir, "launch", "ramp", "--key", "new-checkout-flow", "--percent", "100")
	assert.Error(t, err)
	assert.Contains(t, out, "frozen")

	// Un-deprecate
	out, err = runFlagctl(tempDir, "undeprecate", "--key", "new-checkout-flow")
	require.NoError(t, err)

	// Delete flag
	out, err = runFlagctl(tempDir, "delete", "--key", "new-checkout-flow", "--force")
	require.NoError(t, err)
	assert.Contains(t, out, "Successfully deleted flag")
}
