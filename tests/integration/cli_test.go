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

	// Verify files created
	assert.FileExists(t, filepath.Join(tempDir, "flags.json"))
	assert.FileExists(t, filepath.Join(tempDir, ".flagctl.json"))

	// Idempotent re-init
	out, err = runFlagctl(tempDir, "init")
	require.NoError(t, err)
	assert.Contains(t, out, "Reinitialized existing flagd configuration")

	// 2. Create Boolean Flag
	out, err = runFlagctl(tempDir, "create", "--key", "new-checkout-flow", "--type", "boolean", "--default", "on", "--description", "New checkout flow")
	require.NoError(t, err, "create failed: %s", out)
	assert.Contains(t, out, "Successfully created flag 'new-checkout-flow'")

	// 3. Generate Code
	out, err = runFlagctl(tempDir, "generate", "-l", "typescript")
	require.NoError(t, err, "generate failed: %s", out)
	assert.Contains(t, out, "Successfully generated typescript")
	assert.FileExists(t, filepath.Join(tempDir, "src/flags.gen.ts"))

	// 4. Rollout 50/50
	out, err = runFlagctl(tempDir, "rollout", "--key", "new-checkout-flow", "--splits", "on=50,off=50")
	require.NoError(t, err, "rollout failed: %s", out)
	assert.Contains(t, out, "Successfully updated rollout")

	// 5. Target
	out, err = runFlagctl(tempDir, "target", "--key", "new-checkout-flow", "--attribute", "email", "-o", "endsWith", "-v", "@company.com", "--variant", "on")
	require.NoError(t, err, "target failed: %s", out)
	assert.Contains(t, out, "Successfully added targeting rule")

	// 6. Validate Schema
	out, err = runFlagctl(tempDir, "validate")
	require.NoError(t, err, "validate failed: %s", out)
	assert.Contains(t, out, "is valid according to flagd schema")

	// 7. List Flags
	out, err = runFlagctl(tempDir, "list")
	require.NoError(t, err, "list failed: %s", out)
	assert.Contains(t, out, "new-checkout-flow")

	// 8. Update Flag
	out, err = runFlagctl(tempDir, "update", "--key", "new-checkout-flow", "--state", "DISABLED")
	require.NoError(t, err, "update failed: %s", out)
	assert.Contains(t, out, "Successfully updated flag")

	// Re-enable flag
	out, err = runFlagctl(tempDir, "update", "--key", "new-checkout-flow", "--state", "ENABLED")
	require.NoError(t, err)

	// Test Immutable Key Invariant
	out, err = runFlagctl(tempDir, "update", "--key", "new-checkout-flow", "--rename-to", "renamed-flow")
	assert.Error(t, err)
	assert.Contains(t, out, "IMMUTABLE")

	// 9. Soft Deprecate
	out, err = runFlagctl(tempDir, "deprecate", "--key", "new-checkout-flow", "--reason", "Feature completed")
	require.NoError(t, err, "deprecate failed: %s", out)
	assert.Contains(t, out, "Marked flag 'new-checkout-flow' as DEPRECATED")

	// Verify Frozen State (Rollout should be blocked)
	out, err = runFlagctl(tempDir, "rollout", "--key", "new-checkout-flow", "--splits", "on=100,off=0")
	assert.Error(t, err)
	assert.Contains(t, out, "frozen")

	// Un-deprecate
	out, err = runFlagctl(tempDir, "undeprecate", "--key", "new-checkout-flow")
	require.NoError(t, err)
	assert.Contains(t, out, "Successfully un-deprecated")

	// Deprecate again for deletion test
	_, _ = runFlagctl(tempDir, "deprecate", "--key", "new-checkout-flow")

	// 10. Audit Codebase
	// Create app.ts referencing flag
	appCode := `const isNewCheckout = openFeature.getBooleanValue("new-checkout-flow", false);`
	_ = os.WriteFile(filepath.Join(tempDir, "src/app.ts"), []byte(appCode), 0644)

	out, err = runFlagctl(tempDir, "audit")
	require.NoError(t, err)
	assert.Contains(t, out, "Deprecated Flags Still in Code")

	// 11. Code-Aware Deletion Guard
	// Should fail because app.ts calls the flag
	out, err = runFlagctl(tempDir, "delete", "--key", "new-checkout-flow")
	assert.Error(t, err)
	assert.Contains(t, out, "deletion blocked by code-aware guard")

	// Remove code reference & generated file
	_ = os.Remove(filepath.Join(tempDir, "src/app.ts"))
	_ = os.Remove(filepath.Join(tempDir, "src/flags.gen.ts"))

	// Delete should now succeed
	out, err = runFlagctl(tempDir, "delete", "--key", "new-checkout-flow")
	require.NoError(t, err, "delete failed: %s", out)
	assert.Contains(t, out, "Successfully deleted flag 'new-checkout-flow'")
}

func TestRolloutAutoBalanceAndValidation(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "flagctl-rollout-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	_, _ = runFlagctl(tempDir, "init")
	_, _ = runFlagctl(tempDir, "create", "--key", "test-flag", "--type", "boolean")

	// Single variant auto-fill
	out, err := runFlagctl(tempDir, "rollout", "--key", "test-flag", "--splits", "on=25")
	require.NoError(t, err, "auto-fill failed: %s", out)

	// Invalid total without auto-balance should fail
	out, err = runFlagctl(tempDir, "rollout", "--key", "test-flag", "--splits", "on=20,off=70")
	assert.Error(t, err)
	assert.Contains(t, out, "total sum must equal 100%")

	// Auto-balance should fix it
	out, err = runFlagctl(tempDir, "rollout", "--key", "test-flag", "--splits", "on=20,off=70", "--auto-balance")
	require.NoError(t, err, "auto-balance failed: %s", out)

	// Complete rollout
	out, err = runFlagctl(tempDir, "rollout", "--key", "test-flag", "--complete", "on")
	require.NoError(t, err, "complete rollout failed: %s", out)
	assert.Contains(t, out, "Completed rollout")
}
