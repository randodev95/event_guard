package main

import (
	"log"
	"github.com/eventcanvas/eventcanvas/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
