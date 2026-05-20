package main

import (
	"log"

	"github.com/mnabil1718/taskflow/internal/bootstrap"
	"github.com/mnabil1718/taskflow/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	server := bootstrap.NewServer(cfg)
	log.Fatal(server.Run())
}
