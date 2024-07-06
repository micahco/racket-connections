package main

import (
	"context"
	"encoding/gob"
	"flag"
	"html/template"
	"log"
	"net/http"
	"net/mail"
	"os"
	"time"

	"github.com/alexedwards/scs/pgxstore"
	"github.com/alexedwards/scs/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/micahco/racket-connections/internal/mailer"
	"github.com/micahco/racket-connections/internal/models"
)

type application struct {
	isDevelopment  bool
	baseURL        string
	errorLog       *log.Logger
	infoLog        *log.Logger
	models         models.Models
	templateCache  map[string]*template.Template
	sessionManager *scs.SessionManager
	mailer         *mailer.Mailer
}

func main() {
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)

	dev := flag.Bool("dev", false, "Development mode")
	port := flag.String("port", "4000", "Listening address")
	flag.Parse()

	if *dev {
		err := godotenv.Load()
		if err != nil {
			log.Fatal("Error loading .env file")
		}
	}

	pool, err := newPool(os.Getenv("RC_DB_DSN"))
	if err != nil {
		errorLog.Fatal(err)
	}
	defer pool.Close()

	tc, err := newTemplateCache()
	if err != nil {
		errorLog.Fatal(err)
	}

	sm := scs.New()
	sm.Store = pgxstore.New(pool)
	sm.Lifetime = 12 * time.Hour

	m, err := mailer.New(
		os.Getenv("RC_SMTP_HOST"),
		os.Getenv("RC_SMTP_PORT"),
		os.Getenv("RC_SMTP_USER"),
		os.Getenv("RC_SMTP_PASS"),
		&mail.Address{
			Name:    "Racket Connections",
			Address: "no-reply@cowell.dev",
		},
	)
	if err != nil {
		errorLog.Fatal(err)
	}

	app := &application{
		isDevelopment:  *dev,
		baseURL:        os.Getenv("RC_BASE_URL"),
		errorLog:       errorLog,
		infoLog:        infoLog,
		models:         models.New(pool),
		templateCache:  tc,
		sessionManager: sm,
		mailer:         m,
	}

	gob.Register(FlashMessage{})

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

func newPool(dsn string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(dsn)
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

func (app *application) background(fn func()) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				app.infoLog.Println("BACKGROUND:", err)
			}
		}()
		fn()
	}()
}
