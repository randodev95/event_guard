package validator

import (
	"fmt"
	"github.com/randodev95/event_guard/pkg/normalization"
	"github.com/xeipuuv/gojsonschema"
	"sync"
	"time"
)

// Contract defines the interface for retrieving schemas and identity rules.
// This allows the engine to be decoupled from the concrete AST implementation.
type Contract interface {
	ResolveEventSchema(eventName string) (string, error)
	GetIdentityProperties() []string
	GetEventNames() []string
}

// Result holds the outcome of a validation operation.
type Result struct {
	Valid  bool
	Errors []string
}

// ObservationHandler allows production systems to hook into validation telemetry.
type ObservationHandler interface {
	OnValidationStart(eventName string)
	OnValidationEnd(eventName string, duration int64, valid bool, err error)
}

// Engine provides a high-level SDK for validating events against a Data Contract.
type Engine struct {
	contract           Contract
	mapper             *normalization.Mapper
	compiledCache      sync.Map // map[string]*gojsonschema.Schema
	observationHandler ObservationHandler
}

// NewEngine initializes a new validation engine with the provided contract.
func NewEngine(contract Contract) *Engine {
	return &Engine{
		contract: contract,
		mapper:   normalization.NewDefaultMapper(),
	}
}

// GetMapper returns the normalization mapper used by the engine.
func (e *Engine) GetMapper() *normalization.Mapper {
	return e.mapper
}

// SetObservationHandler sets the telemetry hook for the engine.
func (e *Engine) SetObservationHandler(h ObservationHandler) {
	e.observationHandler = h
}

// ResetCache clears the compiled schema cache.
func (e *Engine) ResetCache() {
	e.compiledCache = sync.Map{}
}

// Warmup pre-compiles all schemas in the tracking plan.
func (e *Engine) Warmup() error {
	for _, name := range e.contract.GetEventNames() {
		_, err := e.getOrCompileSchema(name)
		if err != nil {
			return fmt.Errorf("warmup failed for event [%s]: %w", name, err)
		}
	}
	return nil
}

// ValidateJSON parses a raw JSON payload and validates it against the plan.
func (e *Engine) ValidateJSON(payload []byte) (*Result, error) {
	startTime := time.Now()
	var err error
	var res *Result

	normalized, err := e.mapper.Map(payload)
	if err != nil {
		return nil, fmt.Errorf("normalization failed: %w", err)
	}

	eventName := normalized.Event
	if e.observationHandler != nil {
		e.observationHandler.OnValidationStart(eventName)
	}

	if !e.hasValidIdentity(normalized) {
		res = &Result{
			Valid:  false,
			Errors: []string{"identity_required: no recognized identity properties found"},
		}
		if e.observationHandler != nil {
			e.observationHandler.OnValidationEnd(eventName, time.Since(startTime).Microseconds(), false, nil)
		}
		return res, nil
	}

	schema, err := e.getOrCompileSchema(eventName)
	if err != nil {
		if e.observationHandler != nil {
			e.observationHandler.OnValidationEnd(eventName, time.Since(startTime).Microseconds(), false, err)
		}
		return nil, fmt.Errorf("schema resolution failed: %w", err)
	}

	res, err = e.validateWithSchema(normalized, schema)
	if e.observationHandler != nil {
		e.observationHandler.OnValidationEnd(eventName, time.Since(startTime).Microseconds(), res != nil && res.Valid, err)
	}
	return res, err
}

func (e *Engine) hasValidIdentity(event *normalization.NormalizedEvent) bool {
	idProps := e.contract.GetIdentityProperties()
	if len(idProps) == 0 {
		return true
	}

	for _, idProp := range idProps {
		if val, ok := event.Identity[idProp]; ok && val != "" {
			return true
		}
	}
	return false
}

func (e *Engine) getOrCompileSchema(eventName string) (*gojsonschema.Schema, error) {
	if val, ok := e.compiledCache.Load(eventName); ok {
		return val.(*gojsonschema.Schema), nil
	}

	// Resolve from AST via Contract interface
	s, err := e.contract.ResolveEventSchema(eventName)
	if err != nil {
		return nil, fmt.Errorf("contract violation: %w", err)
	}

	loader := gojsonschema.NewStringLoader(s)
	schema, err := gojsonschema.NewSchema(loader)
	if err != nil {
		return nil, fmt.Errorf("schema compilation failed for [%s]: %w", eventName, err)
	}

	e.compiledCache.Store(eventName, schema)
	return schema, nil
}

func (e *Engine) validateWithSchema(event *normalization.NormalizedEvent, schema *gojsonschema.Schema) (*Result, error) {
	data := make(map[string]interface{}, 16)
	for k, v := range event.Properties {
		data[k] = v
	}
	for k, v := range event.Identity {
		data[k] = v
	}

	documentLoader := gojsonschema.NewGoLoader(data)
	jsonschemaResult, err := schema.Validate(documentLoader)
	if err != nil {
		return nil, fmt.Errorf("internal validation error: %w", err)
	}

	return newResult(jsonschemaResult), nil
}

func newResult(res *gojsonschema.Result) *Result {
	result := &Result{
		Valid: res.Valid(),
	}
	if !res.Valid() {
		for _, desc := range res.Errors() {
			result.Errors = append(result.Errors, desc.String())
		}
	}
	return result
}

// Validate performs a low-level validation of a normalized event against a JSON schema.
func Validate(event *normalization.NormalizedEvent, schema string) (*Result, error) {
	data := make(map[string]interface{})

	for k, v := range event.Properties {
		data[k] = v
	}

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
