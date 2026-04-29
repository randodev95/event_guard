package generator

import (
	"testing"

	"github.com/randodev95/event_guard/pkg/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockPlan(t *testing.T) string {
	return `version: "1.0.0"
taxonomy:
  events:
    "Test Event":
      properties:
        userId: { type: string, required: true }
flows:
  Test:
    nodes:
      Start: { type: TriggerNode, event: "Test Event", transitions: [{target: End}] }
      End: { type: TerminalNode }
`
}

func TestGenerateDBT(t *testing.T) {
	plan, err := parser.ParseYAML([]byte(mockPlan(t)))
	require.NoError(t, err)

	out, err := GenerateDBT(plan)
	require.NoError(t, err)
	assert.Contains(t, out, "name: Test Event")
	assert.Contains(t, out, "not_null")
}

func TestGenerateSQLMesh(t *testing.T) {
	plan, err := parser.ParseYAML([]byte(mockPlan(t)))
	require.NoError(t, err)

	out, err := GenerateSQLMesh(plan)
	require.NoError(t, err)
	assert.Contains(t, out, "MODEL")
}

func TestGenerateHTML(t *testing.T) {
	plan, err := parser.ParseYAML([]byte(mockPlan(t)))
	require.NoError(t, err)

	out, err := GenerateHTML(plan)
	require.NoError(t, err)
	assert.Contains(t, out, "Test Event")
}

func TestGenerateMermaid(t *testing.T) {
	plan, err := parser.ParseYAML([]byte(mockPlan(t)))
	require.NoError(t, err)

	out, err := GenerateMermaid(plan)
	require.NoError(t, err)
	assert.Contains(t, out, "graph TD")
}
