package cfg

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnmarshal(t *testing.T) {
	yml := `
# a comment
name: "testapp"
dependencies:
  production: &default_deps
    telephone-service: {type: soft}
    sms-service: {type: hard}
    mail-service: ~
  staging:
    << : *default_deps
    ? pigeon-service
    fax-service: {}
    messenger-service:
  sandbox: { << : *default_deps }
`

	cfg, err := Unmarshal(strings.NewReader(yml))
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, cfg.AppName, "testapp")
	require.Contains(t, cfg.Dependencies, "production")
	require.Contains(t, cfg.Dependencies, "staging")
	require.Contains(t, cfg.Dependencies, "sandbox")
	assert.Len(t, cfg.Dependencies, 3)

	prd := cfg.Dependencies["production"]
	assert.Contains(t, prd, "telephone-service")
	assert.Contains(t, prd, "sms-service")
	assert.Contains(t, prd, "mail-service")
	assert.Len(t, prd, 3)
	for app, attr := range prd {
		require.NotNilf(t, attr, "type field for production app %q is nil", app)
	}
	assert.Equal(t, "soft", prd["telephone-service"].Type)
	assert.Equal(t, "hard", prd["sms-service"].Type)
	assert.Equal(t, "hard", prd["mail-service"].Type)

	sb := cfg.Dependencies["sandbox"]
	assert.Contains(t, sb, "telephone-service")
	assert.Contains(t, sb, "sms-service")
	assert.Contains(t, sb, "mail-service")
	assert.Len(t, sb, 3)
	for app, attr := range prd {
		require.NotNilf(t, attr, "type field for sandbox app %q is nil", app)
	}
	assert.Equal(t, "soft", sb["telephone-service"].Type)
	assert.Equal(t, "hard", sb["sms-service"].Type)
	assert.Equal(t, "hard", sb["mail-service"].Type)

	stg := cfg.Dependencies["staging"]
	assert.Contains(t, stg, "telephone-service")
	assert.Contains(t, stg, "sms-service")
	assert.Contains(t, stg, "mail-service")
	assert.Contains(t, stg, "pigeon-service")
	assert.Contains(t, stg, "fax-service")
	assert.Contains(t, stg, "messenger-service")
	assert.Len(t, stg, 6)
	for app, attr := range prd {
		require.NotNilf(t, attr, "type field for staging app %q is nil", app)
	}
	assert.Equal(t, "soft", stg["telephone-service"].Type)
	assert.Equal(t, "hard", stg["sms-service"].Type)
	assert.Equal(t, "hard", stg["mail-service"].Type)
	assert.Equal(t, "hard", stg["pigeon-service"].Type)
	assert.Equal(t, "hard", stg["fax-service"].Type)
	assert.Equal(t, "hard", stg["messenger-service"].Type)
}
