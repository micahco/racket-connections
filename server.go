package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/go-chi/chi/v5"
)

const (
	READ_TIMEOUT  = 10 // in seconds
	WRITE_TIMEOUT = 30
)

func (app *application) serve() {
	fmt.Printf("Serving: http://localhost:%d\n", app.config.port)

	mux := chi.NewRouter()
	mux.Use(app.recoverer)

	// Static assets
	fs := http.FileServer(http.Dir(app.config.staticDir))
	p := fmt.Sprintf("/%s/", filepath.Clean(app.config.staticDir))
	mux.Handle(p, http.StripPrefix(p, fs))

	// Routes
	mux.Get("/", app.handleGetIndex)
	mux.Get("/login/", app.handleGetLogin)
	mux.Post("/login/", app.handlePostLogin)
	mux.Post("/signup/", app.handleSignupPost)

	// Create the server
	s := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.config.port),
		Handler:      mux,
		IdleTimeout:  time.Minute,
		ReadTimeout:  time.Duration(READ_TIMEOUT * int64(time.Second)),
		WriteTimeout: time.Duration(WRITE_TIMEOUT * int64(time.Second)),
	}

	s.ListenAndServe()
}

type httpError struct {
	error
	status int
}

func (app *application) recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rvr := recover(); rvr != nil {
				code := http.StatusInternalServerError
				var msg string

				switch v := rvr.(type) {
				case error:
					var httpErr httpError
					if errors.As(v, &httpErr) {
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
