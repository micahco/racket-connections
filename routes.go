package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (app *application) routes() http.Handler {
	r := chi.NewRouter()
	r.Use(app.recovery)
	r.Use(app.logRequests)
	r.Use(secureHeaders)

	r.NotFound(handleNotFound)

	r.Handle("/static/*", app.handleStatic())

	r.Get("/favicon.ico", app.handleFavicon)

	r.Route("/", func(r chi.Router) {
		r.Use(app.sessionManager.LoadAndSave)
		r.Use(app.noSurf)
		r.Use(app.authenticate)

		r.Get("/", app.handleRootGet)

		r.Route("/auth", func(r chi.Router) {
			r.Get("/login", app.handleAuthLoginGet)
			r.Post("/login", app.handleAuthLoginPost)

			r.Post("/logout", app.handleAuthLogoutPost)

			r.Get("/signup", app.handleAuthSignupGet)
			r.Post("/signup", app.handleAuthSignupPost)

			r.Get("/create", app.handleAuthCreateGet)
			r.Post("/create", app.handleAuthCreatePost)

			r.Route("/reset", func(r chi.Router) {
				r.Get("/", app.handleAuthResetGet)
				r.Post("/", app.handleAuthResetPost)

				r.Get("/update", app.handleAuthResetUpdateGet)
				r.Post("/update", app.handleAuthResetUpdatePost)
			})
		})

		r.Route("/profile", func(r chi.Router) {
			r.Use(app.requireAuthentication)

			r.Get("/", app.handleProfileGet)

			r.Get("/edit", app.handleProfileEditGet)
			r.Post("/edit", app.handleProfileEditPost)

			r.Post("/delete", app.handleProfileDeletePost)

			r.Route("/contact", func(r chi.Router) {
				r.Get("/new", app.handleProfileContactNewGet)
				r.Post("/new", app.handleProfileContactNewPost)

				r.Get("/list", nil)

				r.Post("/delete", nil)
			})
		})

		r.Route("/posts", func(r chi.Router) {
			r.Use(app.requireAuthentication)

			r.Get("/", app.handlePostsGet)
			r.Post("/", app.handlePostsPost)

			r.Get("/latest", app.handlePostsLatestGet)

			r.Route("/{id}", func(r chi.Router) {
				r.Get("/*", app.handlePostsIdGet)

				r.Get("/edit", app.handlePostsIdEditGet)
				r.Post("/edit", app.handlePostsIdEditPost)

				r.Get("/delete", app.handlePostsIdDeleteGet)
				r.Post("/delete", app.handlePostsIdDeletePost)
			})

			r.Get("/new", app.handlePostsNewGet)
			r.Post("/new", app.handlePostsNewPost)
		})
	})

	return r
}
