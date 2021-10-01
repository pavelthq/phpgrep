package main

import (
	"log"
	"os"

	"github.com/pavelthq/phpgrep/internal/phpgrep"
)

func main() {
	exitCode, err := phpgrep.Main()
	if err != nil {
		log.Printf("error: %+v", err)
		return
	}
	os.Exit(exitCode)
}
