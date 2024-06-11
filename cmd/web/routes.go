package main

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/go-chi/chi/v5"
	"github.com/micahco/racket-connections/internal/validator"
)

func (app *application) routes() http.Handler {
	r := chi.NewRouter()
	r.Use(app.recovery)
	r.Use(app.logRequests)
	r.Use(secureHeaders)

	r.NotFound(handleNotFound)

	fs := http.FileServer(http.Dir("./static/"))
	r.Handle("/static/*", http.StripPrefix("/static", fs))

	r.Get("/favicon.ico", handleFavicon)

	r.Route("/", func(r chi.Router) {
		r.Use(app.sessionManager.LoadAndSave)
		r.Use(noSurf)
		r.Use(app.authenticate)

		r.Get("/", app.handleRootGet)

		r.Route("/user", func(r chi.Router) {
			r.Get("/login", app.handleUserLoginGet)
			r.Post("/login", app.handleUserLoginPost)

			r.Post("/logout", app.handleUserLogoutPost)

			r.Get("/signup", app.handleUserSignupGet)
			r.Post("/signup", app.handleUserSignupPost)

			r.Route("/profile", func(r chi.Router) {
				r.Use(app.requireAuthentication)

				r.Get("/", app.handleUserProfileGet)
				r.Post("/", app.handleUserProfilePost)
			})
		})

		r.Route("/post", func(r chi.Router) {
			r.Use(app.requireAuthentication)

		})
	})

	return r
}

func (app *application) serverError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	app.errorLog.Output(2, trace)

	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

func unauthorizedError(w http.ResponseWriter) {
	http.Error(w, "permission denied", http.StatusUnauthorized)
}

func validationError(w http.ResponseWriter, v validator.Validator) {
	http.Error(w, v.Errors(), http.StatusUnprocessableEntity)
}
