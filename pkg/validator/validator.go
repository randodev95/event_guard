package validator

import (
	"sync"
	"github.com/eventcanvas/eventcanvas/pkg/ast"
	"github.com/eventcanvas/eventcanvas/pkg/normalization"
	"github.com/xeipuuv/gojsonschema"
)

type Result struct {
	Valid  bool
	Errors []string
}

// Engine provides a high-level SDK for validating events against a TrackingPlan.
// It uses a Mapper to normalize incoming data before validation.
type Engine struct {
	plan        *ast.TrackingPlan
	mapper      *normalization.Mapper
	schemaCache sync.Map // map[string]string (eventName -> JSONSchema)
}

func NewEngine(plan *ast.TrackingPlan) *Engine {
	return &Engine{
		plan:   plan,
		mapper: normalization.NewDefaultMapper(),
	}
}

// ValidateJSON parses a raw JSON payload and validates it against the plan.
func (e *Engine) ValidateJSON(payload []byte) (*Result, error) {
	normalized, err := e.mapper.Map(payload)
	if err != nil {
		return nil, err
	}

	// Get or Resolve Schema
	var schema string
	if val, ok := e.schemaCache.Load(normalized.Event); ok {
		schema = val.(string)
	} else {
		s, err := e.plan.ResolveEventSchema(normalized.Event)
		if err != nil {
			return nil, err
		}
		schema = s
		e.schemaCache.Store(normalized.Event, schema)
	}

	return Validate(normalized, schema)
}

func Validate(event *normalization.NormalizedEvent, schema string) (*Result, error) {
	// Senior FAANG Pattern: Canonicalize data before validation.
	// All identity and properties are merged into a single object for schema validation.
	data := make(map[string]interface{})
	
	// 1. Start with Properties
	for k, v := range event.Properties {
		data[k] = v
	}

	// 2. Overlay Identity (Canonical keys like userId, anonymousId)
	for k, v := range event.Identity {
		data[k] = v
	}

	schemaLoader := gojsonschema.NewStringLoader(schema)
	documentLoader := gojsonschema.NewGoLoader(data)

	res, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return nil, err
	}

	result := &Result{
		Valid: res.Valid(),
	}

	if !res.Valid() {
		for _, desc := range res.Errors() {
			result.Errors = append(result.Errors, desc.String())
		}
	}

	return result, nil
}
