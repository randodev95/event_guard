package main

import (
	"log"
	"github.com/randodev95/eventcanvas/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
