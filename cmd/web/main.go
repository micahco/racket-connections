package main

import (
	"context"
	"flag"
	"html/template"
	"log"
	"net/http"
	"net/mail"
	"net/smtp"
	"os"
	"time"

	"github.com/alexedwards/scs/pgxstore"
	"github.com/alexedwards/scs/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/micahco/racket-connections/internal/models"
)

type application struct {
	url            string
	errorLog       *log.Logger
	infoLog        *log.Logger
	posts          *models.PostModel
	users          *models.UserModel
	verifications  *models.VerificationModel
	templateCache  map[string]*template.Template
	sessionManager *scs.SessionManager
	smtpClient     *smtp.Client
	fromAddress    *mail.Address
}

func main() {
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)

	addr := flag.String("addr", ":4000", "HTTP network address")
	dsn := flag.String("dsn", "", "MySQL data source name")
	smtpHost := flag.String("smtp-host", "", "SMTP hostname")
	smtpPort := flag.String("smtp-port", "", "SMTP port")
	smtpUser := flag.String("smtp-user", "", "SMTP username")
	smtpPass := flag.String("smtp-pass", "", "SMTP password")

	flag.Parse()

	if *dsn == "" || *smtpHost == "" || *smtpPort == "" || *smtpUser == "" || *smtpPass == "" {
		errorLog.Fatal("missing required flag")
	}

	pool, err := newPool(*dsn)
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

	sc, err := newSMTPClient(*smtpHost, *smtpPort, *smtpUser, *smtpPass)
	if err != nil {
		errorLog.Fatal(err)
	}
	defer sc.Close()

	from := &mail.Address{
		Name:    "Racket Connections",
		Address: *smtpUser,
	}

	app := &application{
		url:            "https://localhost" + *addr,
		errorLog:       errorLog,
		infoLog:        infoLog,
		posts:          &models.PostModel{Pool: pool},
		users:          &models.UserModel{Pool: pool},
		verifications:  &models.VerificationModel{Pool: pool},
		templateCache:  tc,
		sessionManager: sm,
		smtpClient:     sc,
		fromAddress:    from,
	}

	srv := &http.Server{
		Addr:     *addr,
		ErrorLog: errorLog,
		Handler:  app.routes(),
	}

	infoLog.Printf("Starting server: %s", app.url)
	err = srv.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")
	errorLog.Fatal(err)
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

	if err := pool.Ping(context.Background()); err != nil {
		return nil, err
	}

	return pool, nil
}
