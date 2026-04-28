package validator

import (
	"github.com/eventcanvas/eventcanvas/pkg/normalization"
	"github.com/xeipuuv/gojsonschema"
)

type Result struct {
	Valid  bool
	Errors []string
}

func Validate(event *normalization.NormalizedEvent, schema string) (*Result, error) {
	// Senior FAANG Pattern: Merge Envelope into Properties
	// This ensures standard fields like userId/anonymousId are validated if the plan requires them.
	data := make(map[string]interface{})
	for k, v := range event.Properties {
		data[k] = v
	}
	if event.UserID != "" {
		data["userId"] = event.UserID
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
