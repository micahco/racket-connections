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

	r.NotFound(app.handleNotFound)
	r.Handle("/static/*", app.handleStatic())
	r.Get("/favicon.ico", app.handleFavicon)

	r.Route("/", func(r chi.Router) {
		r.Use(app.sessionManager.LoadAndSave)
		r.Use(app.noSurf)
		r.Use(app.authenticate)

		r.Get("/", app.handleRoot)
		r.Get("/about", app.handleAbout)
		r.NotFound(app.handleNotFound)

		r.Route("/auth", func(r chi.Router) {
			r.Get("/login", app.handleRedirectToRoot)
			r.Post("/login", app.handleAuthLoginPost)
			r.Get("/logout", app.handleRedirectToRoot)
			r.Post("/logout", app.handleAuthLogoutPost)
			r.Get("/signup", app.handleRedirectToRoot)
			r.Post("/signup", app.handleAuthSignupPost)
			r.Get("/register", app.handleAuthRegisterGet)
			r.Post("/register", app.handleAuthRegisterPost)
			r.Get("/reset", app.handleAuthResetGet)
			r.Post("/reset", app.handleAuthResetPost)
			r.Get("/reset/update", app.handleAuthResetUpdateGet)
			r.Post("/reset/update", app.handleAuthResetUpdatePost)
			r.NotFound(app.handleNotFound)
		})

		r.Route("/profile", func(r chi.Router) {
			r.Use(app.requireAuthentication)

			r.Get("/", app.handleProfileGet)
			r.Get("/contacts", app.handleProfileContactsGet)
			r.Post("/contacts", app.handleProfileContactsPost)
			r.Post("/contacts/delete", app.handleProfileContactsDeletePost)
			r.Get("/availability", app.handleProfileAvailabilityGet)
			r.Post("/availability", app.handleProfileAvailabilityPost)
			r.Get("/delete", app.handleProfileDeleteGet)
			r.Post("/delete", app.handleProfileDeletePost)
			r.NotFound(app.handleNotFound)
		})

		r.Route("/posts", func(r chi.Router) {
			r.Use(app.requireAuthentication)

			r.Get("/", app.handlePostsGet)
			r.Post("/", app.handlePostsPost)
			r.Get("/available", app.handlePostsAvailableGet)
			r.Get("/new", app.handlePostsNewGet)
			r.Post("/new", app.handlePostsNewPost)
			r.NotFound(app.handleNotFound)

			r.Route("/{id}", func(r chi.Router) {
				r.Get("/*", app.handlePostsIdGet)
				r.Get("/delete", app.handlePostsIdDeleteGet)
				r.Post("/delete", app.handlePostsIdDeletePost)
				r.NotFound(app.handleNotFound)
			})
		})
	})

	return r
}

func refresh(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
}

func (app *application) handleNotFound(w http.ResponseWriter, r *http.Request) {
	app.renderError(w, r, http.StatusNotFound, "")
}

func (app *application) handleStatic() http.Handler {
	if app.isDevelopment {
		fs := http.FileServer(http.Dir("./ui/static/"))

		return http.StripPrefix("/static", fs)
	}

	return http.FileServer(http.FS(ui.Files))
}

func (app *application) handleFavicon(w http.ResponseWriter, r *http.Request) {
	if app.isDevelopment {
		http.ServeFile(w, r, "./ui/static/favicon.ico")

		return
	}
	http.ServeFileFS(w, r, ui.Files, "static/favicon.ico")
}

func (app *application) handleRoot(w http.ResponseWriter, r *http.Request) {
	if app.isAuthenticated(r) {
		http.Redirect(w, r, "/posts", http.StatusSeeOther)

		return
	}

	app.render(w, r, http.StatusOK, "login.html", nil)
}

func (app *application) handleRedirectToRoot(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) handleAbout(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, http.StatusOK, "about.html", nil)
}
