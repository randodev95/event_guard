package server

import (
	"testing"

	"github.com/randodev95/event_guard/pkg/ast"
	"github.com/stretchr/testify/assert"
)

func TestServer(t *testing.T) {
	plan := &ast.TrackingPlan{
		Taxonomy: ast.Taxonomy{
			Events: map[string]ast.EventV2{
				"Login": {Properties: map[string]ast.PropertyV2{"user": {Type: "string"}}},
			},
		},
	}

	t.Run("DLQ Path", func(t *testing.T) {
		dlq := make(chan []byte, 1)
		srv := &Server{DLQ: dlq}
		
		// This is just a skeletal check of the server struct
		assert.NotNil(t, srv)
	})

	t.Run("Reload Path", func(t *testing.T) {
		// Mock server reload logic
		assert.NotEmpty(t, plan.Taxonomy.Events)
	})
}
