// Package main is the entry point for the EventCanvas CLI.
package main

import (
	"github.com/randodev95/event_guard/cmd"
	"log"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
