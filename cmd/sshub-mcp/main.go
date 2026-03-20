package main

import (
	"log"
	"os"

	"sshub-mcp/internal/app"
	"sshub-mcp/internal/config"
)

func main() {
	a, err := app.New(config.Load())
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	if err := a.Run(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
