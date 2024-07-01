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
	errorLog       *log.Logger
	infoLog        *log.Logger
	posts          *models.PostModel
	skills         *models.SkillLevelModel
	sports         *models.SportModel
	users          *models.UserModel
	contacts       *models.ContactModel
	timeslots      *models.TimeslotModel
	verifications  *models.VerificationModel
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

	infoLog.Println("Connecting to database")
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

	infoLog.Println("Connecting to SMTP")
	m, err := mailer.New(
		os.Getenv("RC_SMTP_HOST"),
		os.Getenv("RC_SMTP_PORT"),
		os.Getenv("RC_SMTP_USER"),
		os.Getenv("RC_SMTP_PASS"),
		&mail.Address{
			Name:    "Racket Connections",
			Address: "no-reply@rc.cowell.dev",
		},
	)
	if err != nil {
		errorLog.Fatal(err)
	}

	app := &application{
		isDevelopment:  *dev,
		errorLog:       errorLog,
		infoLog:        infoLog,
		posts:          models.NewPostModel(pool),
		skills:         models.NewSkillLevelModel(pool),
		sports:         models.NewSportModel(pool),
		users:          models.NewUserModel(pool),
		contacts:       models.NewContactModel(pool),
		timeslots:      models.NewTimeslotModel(pool),
		verifications:  models.NewVerificationModel(pool),
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
