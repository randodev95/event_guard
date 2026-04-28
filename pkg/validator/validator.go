package validator

import (
	"github.com/randodev95/event_guard/pkg/ast"
	"github.com/randodev95/event_guard/pkg/normalization"
	"github.com/xeipuuv/gojsonschema"
	"sync"
)

// Result holds the outcome of a validation operation.
type Result struct {
	Valid  bool
	Errors []string
}

// Engine provides a high-level SDK for validating events against a TrackingPlan.
// It uses a Mapper to normalize incoming data before validation and caches
// resolved JSON schemas for performance.
type Engine struct {
	plan        *ast.TrackingPlan
	mapper      *normalization.Mapper
	schemaCache sync.Map // map[string]string (eventName -> JSONSchema)
}

// NewEngine initializes a new validation engine with the provided tracking plan.
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

// Validate performs a low-level validation of a normalized event against a JSON schema.
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
