package parser

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseYAML(t *testing.T) {
	t.Run("Valid Plan", func(t *testing.T) {
		data := `
taxonomy:
  events:
    Signup: {}
flows:
  Onboarding:
    nodes:
      Start: { type: TriggerNode, event: Signup, transitions: [{target: End}] }
      End: { type: TerminalNode }
`
		plan, err := ParseYAML([]byte(data))
		require.NoError(t, err)
		assert.Contains(t, plan.Taxonomy.Events, "Signup")
	})

	t.Run("Invalid Plan (Orphan Event)", func(t *testing.T) {
		data := `
taxonomy:
  events:
    Signup: {}
    Orphan: {}
flows:
  Onboarding:
    nodes:
      Start: { type: TriggerNode, event: Signup }
`
		_, err := ParseYAML([]byte(data))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "orphan event")
	})
}

func TestLoadProject(t *testing.T) {
	tempDir := t.TempDir()
	mapsDir := filepath.Join(tempDir, "maps")
	require.NoError(t, os.MkdirAll(mapsDir, 0755))

	taxonomyYAML := `
version: "1.0.0"
taxonomy:
  events:
    Login: { description: "User Login" }
`
	flowYAML := `
flows:
  LoginFlow:
    nodes:
      Start: { type: TriggerNode, event: Login, transitions: [{target: End}] }
      End: { type: TerminalNode }
`
	require.NoError(t, os.WriteFile(filepath.Join(mapsDir, "taxonomy.yaml"), []byte(taxonomyYAML), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(mapsDir, "flow.yaml"), []byte(flowYAML), 0644))

	plan, err := LoadProject(mapsDir)
	require.NoError(t, err)

	assert.Equal(t, "User Login", plan.Taxonomy.Events["Login"].Description)
	assert.Contains(t, plan.Flows, "LoginFlow")
}
