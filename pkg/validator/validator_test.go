package validator

import (
	"testing"

	"github.com/randodev95/event_guard/pkg/ast"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEngine_ValidateJSON(t *testing.T) {
	plan := &ast.TrackingPlan{
		Taxonomy: ast.Taxonomy{
			Events: map[string]ast.EventV2{
				"Signup": {
					Properties: map[string]ast.PropertyV2{
						"properties.age": {Type: "number", Rules: ast.PropertyRules{Min: floatPtr(18)}},
					},
				},
			},
		},
	}
	
	engine := NewEngine(plan)
	
	t.Run("Valid Age", func(t *testing.T) {
		res, err := engine.ValidateJSON([]byte(`{"event": "Signup", "properties": {"age": 20}}`))
		require.NoError(t, err)
		assert.True(t, res.Valid)
	})
	
	t.Run("Invalid Age", func(t *testing.T) {
		res, err := engine.ValidateJSON([]byte(`{"event": "Signup", "properties": {"age": 10}}`))
		require.NoError(t, err)
		assert.False(t, res.Valid)
	})
}

func floatPtr(f float64) *float64 {
	return &f
}
