package main

import (
	"html/template"
	"log"
	"net/url"

	"github.com/alexedwards/scs/v2"
	"github.com/micahco/racket-connections/internal/mailer"
	"github.com/micahco/racket-connections/internal/models"
)

type application struct {
	isDevelopment  bool
	errorLog       *log.Logger
	infoLog        *log.Logger
	baseURL        *url.URL
	models         models.Models
	templateCache  map[string]*template.Template
	sessionManager *scs.SessionManager
	mailer         *mailer.Mailer
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
