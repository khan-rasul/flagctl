package schema

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRegistryAndValidate(t *testing.T) {
	r, err := NewRegistry()
	require.NoError(t, err)
	require.NotNil(t, r)

	validJSON := `{
		"$schema": "https://flagd.dev/schema/v0/flags.json",
		"flags": {
			"test-flag": {
				"state": "ENABLED",
				"defaultVariant": "on",
				"variants": {
					"on": true,
					"off": false
				}
			}
		}
	}`

	err = r.ValidateJSON("v0", []byte(validJSON))
	assert.NoError(t, err)

	invalidJSON := `{
		"flags": {
			"bad-flag": {
				"state": "INVALID_STATE",
				"variants": {}
			}
		}
	}`

	err = r.ValidateJSON("v0", []byte(invalidJSON))
	assert.Error(t, err)

	err = r.ValidateJSON("v99", []byte(validJSON))
	assert.ErrorContains(t, err, "unsupported schema version")
}

func TestDetectVersion(t *testing.T) {
	assert.Equal(t, "v0", DetectVersion(`"$schema": "https://flagd.dev/schema/v0/flags.json"`))
	assert.Equal(t, "v1", DetectVersion(`"$schema": "https://flagd.dev/schema/v1/flags.json"`))
}
