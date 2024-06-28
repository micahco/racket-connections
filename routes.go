package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/micahco/racket-connections/ui"
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

		r.Get("/", app.handleRoot)

		r.Route("/auth", func(r chi.Router) {
			r.Get("/login", app.handleAuthLoginGet)
			r.Post("/login", app.handleAuthLoginPost)

			r.Post("/logout", app.handleAuthLogoutPost)

			r.Get("/signup", app.handleAuthSignupGet)
			r.Post("/signup", app.handleAuthSignupPost)

			r.Get("/register", app.handleAuthRegisterGet)
			r.Post("/register", app.handleAuthRegisterPost)

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

			r.Get("/edit", nil)
			r.Post("/edit", nil)

			r.Post("/delete", nil)
		})

		r.Route("/posts", func(r chi.Router) {
			//r.Use(app.requireAuthentication)

			r.Get("/", app.handlePostsGet)
			r.Post("/", app.handlePostsPost)

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

func refresh(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
}

func (app *application) handleStatic() http.Handler {
	if app.isProduction {
		return http.FileServer(http.FS(ui.Files))
	}

	fs := http.FileServer(http.Dir("./ui/static/"))
	return http.StripPrefix("/static", fs)
}

func (app *application) handleFavicon(w http.ResponseWriter, r *http.Request) {
	if app.isProduction {
		http.ServeFileFS(w, r, ui.Files, "static/favicon.ico")

		return
	}

	http.ServeFile(w, r, "./ui/static/favicon.ico")
}

func (app *application) handleRoot(w http.ResponseWriter, r *http.Request) {
	if app.isAuthenticated(r) {
		http.Redirect(w, r, "/posts", http.StatusSeeOther)

		return
	}

	app.render(w, r, http.StatusOK, "login.html", nil)
}
