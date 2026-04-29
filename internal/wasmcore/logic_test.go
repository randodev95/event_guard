package wasmcore

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBridge_HandleInit(t *testing.T) {
	b := NewBridge()
	yaml := `version: "1.0.0"
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
	res, err := b.HandleInit(yaml)
	require.NoError(t, err)
	assert.Equal(t, "initialized", res)
}

func TestBridge_HandleValidate(t *testing.T) {
	b := NewBridge()
	yaml := `version: "1.0.0"
taxonomy:
  events:
    "Login":
      properties:
        userId: { type: string, required: true }
flows:
  Test:
    nodes:
      Start: { type: TriggerNode, event: "Login", transitions: [{target: End}] }
      End: { type: TerminalNode }
`
	_, _ = b.HandleInit(yaml)

	t.Run("Valid Payload", func(t *testing.T) {
		res, err := b.HandleValidate(`{"event": "Login", "userId": "123"}`)
		require.NoError(t, err)
		assert.Contains(t, res.(string), "\"Valid\":true")
	})

	t.Run("Invalid Payload", func(t *testing.T) {
		res, err := b.HandleValidate(`{"event": "Login"}`)
		require.NoError(t, err)
		assert.Contains(t, res.(string), "\"Valid\":false")
	})
}
