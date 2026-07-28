package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTargetingChain(t *testing.T) {
	cfg := NewFlagConfig("test-app")
	flag := &Flag{
		State:          "ENABLED",
		DefaultVariant: "off",
		Variants: map[string]interface{}{
			"on":  true,
			"off": false,
		},
	}
	cfg.AddFlag("test-feature", flag)

	// Add Allowlist Rule
	err := cfg.AddTargetRule(flag, "email", "endsWith", "@company.com", "on", false)
	require.NoError(t, err)

	rules := cfg.GetTargetRules(flag)
	require.Len(t, rules, 1)
	assert.Equal(t, "email", rules[0].Attribute)
	assert.Equal(t, RuleTypeAllowlist, rules[0].Type)

	// Add Denylist Rule (Top)
	err = cfg.AddTargetRule(flag, "email", "endsWith", "@competitor.com", "off", true)
	require.NoError(t, err)

	rules = cfg.GetTargetRules(flag)
	require.Len(t, rules, 2)
	assert.Equal(t, RuleTypeDenylist, rules[0].Type)

	// Remove Rule by Index
	err = cfg.RemoveTargetRule(flag, 1)
	require.NoError(t, err)

	rules = cfg.GetTargetRules(flag)
	require.Len(t, rules, 1)

	// Add Launch Ramp
	err = cfg.AddLaunchRamp(flag, "", nil, map[string]int{"on": 50, "off": 50}, "userId")
	require.NoError(t, err)
}
