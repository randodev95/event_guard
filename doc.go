/*
Package eventcanvas provides a high-integrity telemetry engine for modeling and validating 
complex tracking plans. It supports semantic versioning of schemas, property inheritance, 
and format-agnostic normalization for production data pipelines.

Installation:
    go get github.com/randodev95/eventcanvas

Core Components:
    - pkg/ast: Core tracking plan data models and resolution logic.
    - pkg/validator: High-performance validation engine with schema caching.
    - pkg/normalization: Format-agnostic mapping (camelCase, snake_case, nested).

Example Usage:
    import "github.com/randodev95/eventcanvas/pkg/validator"
    import "github.com/randodev95/eventcanvas/pkg/parser"

    plan, _ := parser.ParseYAML(data)
    engine := validator.NewEngine(plan)
    result, _ := engine.ValidateJSON(payload)
*/
package main
