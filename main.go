package main

import (
	"log"
	"github.com/randodev95/event_guard/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
