package scanner

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/khan-rasul/flagctl/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScannerAndAudit(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "flagctl-scanner-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a sample TS file
	tsCode := `
import { openFeature } from '@openfeature/web-sdk';

const isNewCheckout = await openFeature.getBooleanValue("new-checkout-flow", false);
const theme = openFeature.getStringValue("header-theme", "dark");
const typoFlag = openFeature.getBooleanValue("missing-flag-key", false);
const accessor = flags.getNewCheckoutFlow(client);
`
	err = os.WriteFile(filepath.Join(tempDir, "app.ts"), []byte(tsCode), 0644)
	require.NoError(t, err)

	// Create FlagConfig
	cfg := config.NewFlagConfig("test-app")
	cfg.AddFlag("new-checkout-flow", &config.Flag{State: "ENABLED", DefaultVariant: "off"})
	cfg.AddFlag("header-theme", &config.Flag{State: "ENABLED", DefaultVariant: "dark"})
	cfg.AddFlag("deprecated-flag", &config.Flag{
		State:          "ENABLED",
		DefaultVariant: "off",
		Metadata:       map[string]interface{}{"deprecated": true},
	})
	cfg.AddFlag("orphaned-flag", &config.Flag{State: "ENABLED", DefaultVariant: "off"})

	// Scanner with custom pattern
	customPattern := `customFlagCall\s*\(\s*["']([^"']+)["']`
	scanner := NewScanner([]string{customPattern})
	matches, err := scanner.ScanDirectory(tempDir)
	require.NoError(t, err)
	assert.NotEmpty(t, matches)

	// Test FindFlagReferences
	refs, err := scanner.FindFlagReferences(tempDir, "new-checkout-flow")
	require.NoError(t, err)
	assert.NotEmpty(t, refs)

	// Test Audit
	report, err := scanner.Audit(tempDir, cfg)
	require.NoError(t, err)

	// Missing flag check
	assert.Len(t, report.MissingFlags, 1)
	assert.Equal(t, "missing-flag-key", report.MissingFlags[0].FlagKey)

	// Orphaned flag check
	assert.Contains(t, report.OrphanedFlags, "orphaned-flag")
}

func TestPascalToKebab(t *testing.T) {
	assert.Equal(t, "new-checkout-flow", pascalToKebab("getNewCheckoutFlow"))
	assert.Equal(t, "header-theme", pascalToKebab("isHeaderTheme"))
	assert.Equal(t, "dark-mode", pascalToKebab("DarkMode"))
}
