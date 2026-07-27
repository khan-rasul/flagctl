package generator

import (
	"testing"

	"github.com/khan-rasul/flagctl/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerators(t *testing.T) {
	cfg := config.NewFlagConfig("test-app")

	// Boolean flag
	cfg.AddFlag("new-checkout-flow", &config.Flag{
		State:          "ENABLED",
		DefaultVariant: "on",
		Variants: map[string]interface{}{
			"on":  true,
			"off": false,
		},
		Metadata: map[string]interface{}{"description": "Enable checkout UI"},
	})

	// String flag
	cfg.AddFlag("header-theme", &config.Flag{
		State:          "ENABLED",
		DefaultVariant: "dark",
		Variants: map[string]interface{}{
			"dark":  "dark-theme",
			"light": "light-theme",
		},
	})

	// Number flag
	cfg.AddFlag("max-connections", &config.Flag{
		State:          "ENABLED",
		DefaultVariant: "standard",
		Variants: map[string]interface{}{
			"standard": 10.0,
			"high":     100.0,
		},
	})

	// Object flag
	cfg.AddFlag("theme-config", &config.Flag{
		State:          "ENABLED",
		DefaultVariant: "v1",
		Variants: map[string]interface{}{
			"v1": map[string]interface{}{"color": "blue"},
		},
	})

	// Test TypeScript Generator
	tsGen, err := GetGenerator("typescript")
	require.NoError(t, err)
	tsCode, err := tsGen.Generate(cfg)
	require.NoError(t, err)
	assert.Contains(t, tsCode, "getNewCheckoutFlow")
	assert.Contains(t, tsCode, "new-checkout-flow")
	assert.Contains(t, tsCode, "getHeaderTheme")
	assert.Contains(t, tsCode, "getMaxConnections")
	assert.Contains(t, tsCode, "getThemeConfig")

	// Test Go Generator
	goGen, err := GetGenerator("go")
	require.NoError(t, err)
	goCode, err := goGen.Generate(cfg)
	require.NoError(t, err)
	assert.Contains(t, goCode, "NewCheckoutFlow")
	assert.Contains(t, goCode, "HeaderTheme")
	assert.Contains(t, goCode, "MaxConnections")
	assert.Contains(t, goCode, "ThemeConfig")

	// Test Invalid Language
	_, err = GetGenerator("invalid-lang")
	assert.Error(t, err)
}
