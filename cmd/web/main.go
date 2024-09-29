package main

import (
	"context"
	"encoding/gob"
	"flag"
	"log"
	"net/http"
	"net/mail"
	"net/url"
	"os"
	"time"

	"github.com/alexedwards/scs/pgxstore"
	"github.com/alexedwards/scs/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/micahco/racket-connections/internal/env"
	"github.com/micahco/racket-connections/internal/mailer"
	"github.com/micahco/racket-connections/internal/models"
)

func main() {
	// Parse CLI flags
	dev := flag.Bool("dev", false, "Development mode")
	port := flag.String("port", "8080", "Listening address")
	flag.Parse()

	// Loggers
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)

	// Load .env for development
	if *dev {
		err := godotenv.Load()
		if err != nil {
			errorLog.Fatal("Error loading .env file")
		}
	}

	// Create base URL
	rawURL, err := env.Get("RC_BASE_URL")
	if err != nil {
		errorLog.Fatal(err)
	}

	baseURL, err := url.Parse(rawURL)
	if err != nil {
		errorLog.Fatal(err)
	}

	// PostgreSQL
	pool, err := newPool()
	if err != nil {
		errorLog.Fatal(err)
	}
	defer pool.Close()

	// HTML template cache
	tc, err := newTemplateCache()
	if err != nil {
		errorLog.Fatal(err)
	}

	// Session manager
	sm := scs.New()
	sm.Store = pgxstore.New(pool)
	sm.Lifetime = 12 * time.Hour

	// SMTP mailer
	m, err := newMailer()
	if err != nil {
		errorLog.Fatal(err)
	}

	// New app
	app := &application{
		isDevelopment:  *dev,
		errorLog:       errorLog,
		infoLog:        infoLog,
		baseURL:        baseURL,
		models:         models.New(pool),
		templateCache:  tc,
		sessionManager: sm,
		mailer:         m,
	}

	// Required to encode/decode session flash messages
	gob.Register(FlashMessage{})

	// Listen and serve
	srv := &http.Server{
		Addr:     ":" + *port,
		ErrorLog: errorLog,
		Handler:  app.routes(),
	}

	infoLog.Println("listening on", *port)
	err = srv.ListenAndServe()
	if err != nil {
		errorLog.Fatal(err)
	}
}

func newPool() (*pgxpool.Pool, error) {
	connString, err := env.Get("RC_DB_URL")
	if err != nil {
		return nil, err
	}

	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, err
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, err
	}

	err = pool.Ping(context.Background())
	if err != nil {
		return nil, err
	}

	return pool, nil
}

func newMailer() (*mailer.Mailer, error) {
	host, err := env.Get("RC_SMTP_HOST")
	if err != nil {
		return nil, err
	}

	port, err := env.Get("RC_SMTP_PORT")
	if err != nil {
		return nil, err
	}

	user, err := env.Get("RC_SMTP_USER")
	if err != nil {
		return nil, err
	}

	pass, err := env.Get("RC_SMTP_PASS")
	if err != nil {
		return nil, err
	}

	return mailer.New(host, port, user, pass,
		&mail.Address{
			Name:    "Racket Connections",
			Address: "no-reply@cowell.dev",
		},
	)
}
