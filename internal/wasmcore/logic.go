package wasmcore

import (
	"encoding/json"
	"fmt"

	"github.com/randodev95/event_guard/pkg/parser"
	"github.com/randodev95/event_guard/pkg/validator"
)

type Bridge struct {
	engine *validator.Engine
}

func NewBridge() *Bridge {
	return &Bridge{}
}

func (b *Bridge) HandleInit(yamlStr string) (interface{}, error) {
	plan, err := parser.ParseYAML([]byte(yamlStr))
	if err != nil {
		return nil, err
	}
	b.engine = validator.NewEngine(plan)
	b.engine.Warmup()
	return "initialized", nil
}

func (b *Bridge) HandleValidate(payload string) (interface{}, error) {
	if b.engine == nil {
		return nil, fmt.Errorf("engine not initialized")
	}
	res, err := b.engine.ValidateJSON([]byte(payload))
	if err != nil {
		return nil, err
	}
	
	jsonRes, err := json.Marshal(res)
	if err != nil {
		return nil, err
	}
	return string(jsonRes), nil
}
