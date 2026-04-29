//go:build wasm
package main

import (
	"syscall/js"

	"github.com/randodev95/event_guard/internal/wasmcore"
)

var bridge = wasmcore.NewBridge()

func initEngine(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return "missing plan yaml"
	}
	res, err := bridge.HandleInit(args[0].String())
	if err != nil {
		return err.Error()
	}
	return res
}

func validate(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return "missing payload"
	}
	res, err := bridge.HandleValidate(args[0].String())
	if err != nil {
		return err.Error()
	}
	return res
}

func main() {
	if js.Global().Get("egInit").IsUndefined() {
		js.Global().Set("egInit", js.FuncOf(initEngine))
	}
	if js.Global().Get("egValidate").IsUndefined() {
		js.Global().Set("egValidate", js.FuncOf(validate))
	}

	select {}
}
