package main

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type config struct {
	port         int
	staticDir    string
	templatesDir string
	dsn          string
}

type application struct {
	config config
	store  store
}

func main() {
	// Load .env vars
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	// Parse env vars to config
	var cfg config

	cfg.port, err = strconv.Atoi(os.Getenv("RC_PORT"))
	if err != nil {
		log.Fatal("RC_PORT:", err)
	}

	cfg.staticDir = os.Getenv("RC_STATIC_DIR")
	cfg.templatesDir = os.Getenv("RC_TEMPLATES_DIR")
	cfg.dsn = os.Getenv("RC_DSN")

	// Create store
	pg, err := newPgStore(cfg.dsn)
	if err != nil {
		log.Fatal(err)
	}
	if err := pg.init(); err != nil {
		log.Fatal(err)
	}

	// Create the app and start the server
	app := &application{
		config: cfg,
		store:  pg,
	}

	app.serve()
}
