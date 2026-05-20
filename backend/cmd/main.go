package main

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/mnabil1718/taskflow/internal/bootstrap"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, reading from environment")
	}

	server := bootstrap.NewServer()
	log.Fatal(server.Run())
}
