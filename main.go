package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	store, err := NewPostgresStore(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	if err := store.Init(); err != nil {
		log.Fatal(err)
	}

	server := NewAPIServer(os.Getenv("PORT"), store)
	server.Run()
}
