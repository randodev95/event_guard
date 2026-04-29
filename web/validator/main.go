//go:build wasm
package main

import (
	"encoding/json"
	"syscall/js"

	"github.com/randodev95/event_guard/pkg/parser"
	"github.com/randodev95/event_guard/pkg/validator"
)

var engine *validator.Engine

func initEngine(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return "missing plan yaml"
	}
	yamlStr := args[0].String()
	plan, err := parser.ParseYAML([]byte(yamlStr))
	if err != nil {
		return err.Error()
	}
	engine = validator.NewEngine(plan)
	engine.Warmup()
	return "initialized"
}

func validate(this js.Value, args []js.Value) interface{} {
	if engine == nil {
		return "engine not initialized"
	}
	if len(args) < 1 {
		return "missing payload"
	}
	payload := args[0].String()
	res, err := engine.ValidateJSON([]byte(payload))
	if err != nil {
		return err.Error()
	}
	
	jsonRes, _ := json.Marshal(res)
	return string(jsonRes)
}

func main() {
	// Guard against re-initialization if already set in global JS context
	if js.Global().Get("egInit").IsUndefined() {
		js.Global().Set("egInit", js.FuncOf(initEngine))
	}
	if js.Global().Get("egValidate").IsUndefined() {
		js.Global().Set("egValidate", js.FuncOf(validate))
	}

	// Keep alive for browser runtime
	select {}
}
