package main

import (
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const (
	READ_TIMEOUT        = time.Duration(10 * int64(time.Second))
	WRITE_TIMEOUT       = time.Duration(30 * int64(time.Second))
	BASE_TEMPLATE       = "base.html"
	SESSION_COOKIE_NAME = "session_token"
)

func (app *application) serve() {
	fmt.Printf("Serving: http://localhost:%d\n", app.config.port)

	mux := chi.NewRouter()
	mux.Use(middleware.StripSlashes)
	mux.Use(app.session.LoadAndSave)
	mux.Use(app.recoverer)

	// Static assets
	fs := http.FileServer(http.Dir(app.config.staticDir))
	p := fmt.Sprintf("/%s/", filepath.Clean(app.config.staticDir))
	mux.Handle(p, http.StripPrefix(p, fs))

	// Routes
	mux.Get("/", app.handleGetIndex)
	mux.Get("/login", app.handleGetLogin)
	mux.Post("/login", app.handlePostLogin)
	mux.Get("/signup", app.handleGetSignup)
	mux.Post("/signup", app.handlePostSignup)
	mux.Get("/logout", app.handleGetLogout)
	mux.Get("/forgot", app.handleGetForgot)

	// Create the server
	s := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.config.port),
		Handler:      mux,
		IdleTimeout:  time.Minute,
		ReadTimeout:  READ_TIMEOUT,
		WriteTimeout: WRITE_TIMEOUT,
	}

	s.ListenAndServe()
}

type httpError struct {
	error
	status int
}

// Recover from panic in routes and reply with proper error
func (app *application) recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rvr := recover(); rvr != nil {
				// Default status code
				code := http.StatusInternalServerError
				var msg string

				// Determine the type of error and extract the message
				switch v := rvr.(type) {
				case error:
					var httpErr httpError
					if errors.As(v, &httpErr) {
						// Update code with http error status
						code = httpErr.status
					}
					msg = v.Error()
				default:
					msg = fmt.Sprint(v)
				}

				log.Println(r.Method, r.URL, code, msg)

				http.Error(w, msg, code)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// Render nested template (filename) in config.templateDir with base template
func (app *application) renderTemplate(w http.ResponseWriter, filename string, data any) error {
	base := filepath.Join(app.config.templatesDir, BASE_TEMPLATE)
	tmpl := filepath.Join(app.config.templatesDir, filename)

	t, err := template.ParseFiles(base, tmpl)
	if err != nil {
		return fmt.Errorf("unable to parse template: %w", err)
	}

	return t.Execute(w, data)
}
