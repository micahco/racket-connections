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
	isProduction   bool
	errorLog       *log.Logger
	infoLog        *log.Logger
	posts          *models.PostModel
	skills         *models.SkillModel
	sports         *models.SportModel
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
	prod := flag.Bool("prod", false, "Production mode")

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

	var sc *smtp.Client
	if *prod {
		sc, err = newSMTPClient(*smtpHost, *smtpPort, *smtpUser, *smtpPass)
		if err != nil {
			errorLog.Fatal(err)
		}
		defer sc.Close()
	}

	var url string
	if *prod {
		url = "https://localhost" + *addr
	} else {
		url = "http://localhost" + *addr
	}

	app := &application{
		url:            url,
		isProduction:   *prod,
		errorLog:       errorLog,
		infoLog:        infoLog,
		posts:          models.NewPostModel(pool),
		skills:         models.NewSkillModel(pool),
		sports:         models.NewSportModel(pool),
		users:          models.NewUserModel(pool),
		verifications:  models.NewVerificationModel(pool),
		templateCache:  tc,
		sessionManager: sm,
		smtpClient:     sc,
		fromAddress: &mail.Address{
			Name:    "Racket Connections",
			Address: *smtpUser,
		},
	}

	srv := &http.Server{
		Addr:     *addr,
		ErrorLog: errorLog,
		Handler:  app.routes(),
	}

	infoLog.Printf("Starting server: %s", app.url)
	if app.isProduction {
		err = srv.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")
	} else {
		srv.ListenAndServe()
	}
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
