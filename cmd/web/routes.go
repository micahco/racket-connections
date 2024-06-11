package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/micahco/racket-connections/internal/models"
	"github.com/micahco/racket-connections/internal/validator"
)

func (app *application) routes() http.Handler {
	r := chi.NewRouter()
	r.Use(app.recovery)
	r.Use(app.logRequests)
	r.Use(secureHeaders)

	r.NotFound(app.handleNotFound)

	fs := http.FileServer(http.Dir("./static/"))
	r.Handle("/static/*", http.StripPrefix("/static", fs))
	r.HandleFunc("/favicon.ico", handleFavicon)

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

func (app *application) handleNotFound(w http.ResponseWriter, r *http.Request) {
	app.notFound(w)
}

func handleFavicon(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./static/favicon.ico")
}

func (app *application) handleRootGet(w http.ResponseWriter, r *http.Request) {
	app.render(w, http.StatusOK, "home.html", app.newTemplateData(r))
}

func (app *application) handleUserLoginGet(w http.ResponseWriter, r *http.Request) {
	if app.isAuthenticated(r) {
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	app.render(w, http.StatusOK, "login.html", app.newTemplateData(r))
}

type userLoginForm struct {
	email    string
	password string
	validator.Validator
}

func (app *application) handleUserLoginPost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
	}

	form := userLoginForm{
		email:    r.Form.Get("email"),
		password: r.Form.Get("password"),
	}

	form.Validate(validator.NotBlank(form.email), "invalid email: cannot be blank")
	form.Validate(validator.Matches(form.email, validator.EmailRX), "invalid email: must be a valid email address")
	form.Validate(validator.NotBlank(form.password), "invalid password: cannot be blank")

	if !form.IsValid() {
		app.invalidRequest(w, form.Validator)

		return
	}

	id, err := app.users.Authenticate(form.email, form.password)
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			app.permissionDenied(w)
		} else if errors.Is(err, models.ErrNotVerified) {
			app.flash(r, "Please verify your email address")

			http.Redirect(w, r, "/", http.StatusSeeOther)
		} else {
			app.serverError(w, err)
		}

		return
	}

	err = app.sessionLogin(r, id)
	if err != nil {
		app.serverError(w, err)

		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) handleUserLogoutPost(w http.ResponseWriter, r *http.Request) {
	err := app.sessionLogout(r)
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.flash(r, "You have been logged out successfully!")

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

type userSignupForm struct {
	name     string
	email    string
	password string
	validator.Validator
}

func (app *application) handleUserSignupGet(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

func (app *application) handleUserSignupPost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)

		return
	}

	form := userSignupForm{
		name:     r.Form.Get("name"),
		email:    r.Form.Get("email"),
		password: r.Form.Get("password"),
	}

	form.Validate(validator.NotBlank(form.name), "invalid name: cannot be blank")
	form.Validate(validator.NotBlank(form.email), "invalid email: cannot be blank")
	form.Validate(validator.Matches(form.email, validator.EmailRX), "invalid email: must be a valid email address")
	form.Validate(validator.PermittedEmailDomain(form.email, "oregonstate.edu"), "invalid email: must be an OSU email address")
	form.Validate(validator.NotBlank(form.password), "invalid password: cannot be blank")
	form.Validate(validator.MinChars(form.password, 8), "invalid password: must be at least 8 characters long")
	form.Validate(validator.MaxChars(form.password, 72), "invalid password: must be no more than 72 characters long")

	if !form.IsValid() {
		app.invalidRequest(w, form.Validator)

		return
	}

	err = app.users.Insert(form.name, form.email, form.password)
	if err != nil {
		if errors.Is(err, models.ErrDuplicateEmail) {
			app.permissionDenied(w)
		} else {
			app.serverError(w, err)
		}

		return
	}

	// send verification email

	msg := fmt.Sprintf("Verification email sent to %s", form.email)
	app.flash(r, msg)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) handleUserProfileGet(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "show user profile")
}

func (app *application) handleUserProfilePost(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "update user profile")
}
