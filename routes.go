package main

import (
	"embed"
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/micahco/racket-connections/internal/crypto"
	"github.com/micahco/racket-connections/internal/models"
	"github.com/micahco/racket-connections/internal/validator"
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
		r.Use(noSurf)
		r.Use(app.authenticate)

		r.Get("/", app.handleIndexGet)

		r.Route("/auth", func(r chi.Router) {
			r.Get("/login", app.handleAuthLoginGet)
			r.Post("/login", app.handleAuthLoginPost)

			r.Post("/logout", app.handleAuthLogoutPost)

			r.Get("/signup", app.handleAuthSignupGet)
			r.Post("/signup", app.handleAuthSignupPost)

			r.Get("/create", app.handleAuthCreateGet)
			r.Post("/create", app.handleAuthCreatePost)

			r.Get("/forgot", app.handleAuthForgotGet)
			r.Post("/forgot", app.handleAuthForgotPost)

			r.Get("/reset", app.handleAuthResetGet)
			r.Post("/reset", app.handleAuthResetPost)
		})

		r.Route("/profile", func(r chi.Router) {
			r.Use(app.requireAuthentication)

			r.Get("/", app.handleAuthProfileGet)

			r.Get("/edit", app.handleAuthProfileGet)
			r.Post("/edit", app.handleAuthProfilePost)

			r.Post("/delete", app.handleAuthProfilePost)
		})

		r.Route("/post", func(r chi.Router) {
			r.Use(app.requireAuthentication)

		})
	})

	return r
}

// Errors

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

// Handlers

func handleNotFound(w http.ResponseWriter, r *http.Request) {
	clientError(w, http.StatusNotFound)
}

//go:embed static
var staticFS embed.FS

func (app *application) handleStatic() http.Handler {
	if app.isProduction {
		return http.FileServer(http.FS(staticFS))
	}

	fs := http.FileServer(http.Dir("./static/"))
	return http.StripPrefix("/static", fs)
}

func (app *application) handleFavicon(w http.ResponseWriter, r *http.Request) {
	if app.isProduction {
		http.ServeFileFS(w, r, staticFS, "./static/favicon.ico")

		return
	}

	http.ServeFile(w, r, "./static/favicon.ico")
}

func (app *application) handleIndexGet(w http.ResponseWriter, r *http.Request) {
	app.render(w, http.StatusOK, "home.html", app.newTemplateData(r))
}

func (app *application) handleAuthLoginGet(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

type authLoginForm struct {
	email    string
	password string
	validator.Validator
}

func (app *application) handleAuthLoginPost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		clientError(w, http.StatusBadRequest)
	}

	form := authLoginForm{
		email:    r.Form.Get("email"),
		password: r.Form.Get("password"),
	}

	form.Validate(validator.NotBlank(form.email), "invalid email: cannot be blank")
	form.Validate(validator.Matches(form.email, validator.EmailRX), "invalid email: must be a valid email address")
	form.Validate(validator.NotBlank(form.password), "invalid password: cannot be blank")

	if !form.IsValid() {
		validationError(w, form.Validator)

		return
	}

	id, err := app.users.Authenticate(form.email, form.password)
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			unauthorizedError(w)
		} else {
			app.serverError(w, err)
		}

		return
	}

	err = app.login(r, id)
	if err != nil {
		app.serverError(w, err)

		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) handleAuthLogoutPost(w http.ResponseWriter, r *http.Request) {
	err := app.logout(r)
	if err != nil {
		app.serverError(w, err)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) handleAuthSignupGet(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

type authSignupForm struct {
	email string
	validator.Validator
}

func (app *application) handleAuthSignupPost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		clientError(w, http.StatusBadRequest)

		return
	}

	form := authSignupForm{email: r.Form.Get("email")}

	form.Validate(validator.NotBlank(form.email), "invalid email: cannot be blank")
	form.Validate(validator.Matches(form.email, validator.EmailRX), "invalid email: must be a valid email address")
	form.Validate(validator.MaxChars(form.email, 254), "invalid email: must be no more than 254 characters long")
	form.Validate(validator.PermittedEmailDomain(form.email, "oregonstate.edu"), "invalid email: must be an OSU email address")

	if !form.IsValid() {
		validationError(w, form.Validator)

		return
	}

	// Check if link verification has already been created
	v, err := app.verifications.Get(form.email)
	if err != nil && err != models.ErrNoRecord {
		app.serverError(w, err)

		return
	}

	// Don't send a new link if less than 5 minutes since last
	if v != nil {
		min := 5 * time.Minute
		if time.Since(v.CreatedAt) < min {
			app.flash(r, "A link to activate your account has been emailed to the address provided.")

			http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)

			return
		}
	}

	token, err := crypto.GenerateRandomString(32)
	if err != nil {
		app.serverError(w, err)

		return
	}

	url := fmt.Sprintf("%s/auth/create?token=%s", app.url, token)
	html := fmt.Sprintf("<p>Please follow the link below to activate your account:<p>"+
		"<a href=\"%s\">%s</a>", url, url)

	err = app.sendMail(form.email, "Email verification", html)
	if err != nil {
		app.serverError(w, err)

		return
	}

	err = app.verifications.Insert(token, form.email)
	if err != nil {
		app.serverError(w, err)

		return
	}

	app.sessionManager.Clear(r.Context())
	app.sessionManager.RenewToken(r.Context())
	app.sessionManager.Put(r.Context(), verificationEmailSessionKey, form.email)

	app.flash(r, "A link to activate your account has been emailed to the address provided.")

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

type authCreatePage struct {
	HasSessionEmail bool
}

func (app *application) handleAuthCreateGet(w http.ResponseWriter, r *http.Request) {
	queryToken := r.URL.Query().Get("token")
	if queryToken == "" {
		unauthorizedError(w)

		return
	}

	app.sessionManager.Put(r.Context(), verificationTokenSessionKey, queryToken)
	exists := app.sessionManager.Exists(r.Context(), verificationEmailSessionKey)

	data := app.newTemplateData(r)
	data.Page = authCreatePage{HasSessionEmail: exists}
	app.render(w, http.StatusOK, "auth-create.html", data)
}

type authCreateForm struct {
	name     string
	email    string
	password string
	validator.Validator
}

func (app *application) handleAuthCreatePost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		clientError(w, http.StatusBadRequest)

		return
	}

	form := authCreateForm{
		name:     r.Form.Get("name"),
		email:    r.Form.Get("email"),
		password: r.Form.Get("password"),
	}

	email := app.sessionManager.GetString(r.Context(), verificationEmailSessionKey)
	if form.email != "" {
		email = form.email
	}

	form.Validate(validator.NotBlank(form.name), "invalid name: cannot be blank")
	form.Validate(validator.NotBlank(email), "invalid email: cannot be blank")
	form.Validate(validator.Matches(email, validator.EmailRX), "invalid email: must be a valid email address")
	form.Validate(validator.MaxChars(email, 254), "invalid email: must be no more than 254 characters long")
	form.Validate(validator.NotBlank(form.password), "invalid password: cannot be blank")
	form.Validate(validator.MinChars(form.password, 8), "invalid password: must be at least 8 characters long")
	form.Validate(validator.MaxChars(form.password, 72), "invalid password: must be no more than 72 characters long")

	if !form.IsValid() {
		validationError(w, form.Validator)

		return
	}

	token := app.sessionManager.GetString(r.Context(), verificationTokenSessionKey)
	if token == "" {
		unauthorizedError(w)

		return
	}

	err = app.verifications.Verify(token, email)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			unauthorizedError(w)
		} else if errors.Is(err, models.ErrExpiredVerification) {
			app.flash(r, "Your verification token is expired. Please signup for a new account.")

			http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		} else {
			app.serverError(w, err)
		}

		return
	}

	err = app.verifications.Purge(email)
	if err != nil {
		app.serverError(w, err)

		return
	}

	id, err := app.users.Insert(form.name, email, form.password)
	if err != nil {
		if errors.Is(err, models.ErrDuplicateEmail) {
			unauthorizedError(w)
		} else {
			app.serverError(w, err)
		}

		return
	}

	app.sessionManager.Clear(r.Context())
	err = app.login(r, id)
	if err != nil {
		app.serverError(w, err)

		return
	}

	app.flash(r, "Successfully created account. Welcome!")

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) handleAuthForgotGet(w http.ResponseWriter, r *http.Request) {
	app.render(w, http.StatusOK, "auth-forgot.html", app.newTemplateData(r))
}

type authForgotForm struct {
	email string
	validator.Validator
}

func (app *application) handleAuthForgotPost(w http.ResponseWriter, r *http.Request) {
	msg := "A link to reset your password has been emailed to the address provided."

	err := r.ParseForm()
	if err != nil {
		clientError(w, http.StatusBadRequest)

		return
	}

	form := authForgotForm{email: r.Form.Get("email")}

	form.Validate(validator.NotBlank(form.email), "invalid email: cannot be blank")
	form.Validate(validator.MaxChars(form.email, 254), "invalid email: must be no more than 254 characters long")

	if !form.IsValid() {
		validationError(w, form.Validator)

		return
	}

	exists, err := app.users.ExistsEmail(form.email)
	if err != nil {
		app.serverError(w, err)
		return
	}

	if !exists {
		app.flash(r, msg)

		http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)

		return
	}

	token, err := crypto.GenerateRandomString(32)
	if err != nil {
		app.serverError(w, err)

		return
	}

	url := fmt.Sprintf("%s/auth/reset?token=%s", app.url, token)
	html := fmt.Sprintf("<p>Please follow the link below to reset your password:<p>"+
		"<a href=\"%s\">%s</a>", url, url)

	err = app.sendMail(form.email, "Reset password", html)
	if err != nil {
		app.serverError(w, err)

		return
	}

	err = app.verifications.Insert(token, form.email)
	if err != nil {
		app.serverError(w, err)

		return
	}

	app.sessionManager.Clear(r.Context())
	app.sessionManager.RenewToken(r.Context())
	app.sessionManager.Put(r.Context(), resetEmailSessionKey, form.email)

	app.flash(r, msg)

	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
}

type authResetPage struct {
	HasSessionEmail bool
}

func (app *application) handleAuthResetGet(w http.ResponseWriter, r *http.Request) {
	queryToken := r.URL.Query().Get("token")
	if queryToken == "" {
		unauthorizedError(w)

		return
	}

	app.sessionManager.Put(r.Context(), resetTokenSessionKey, queryToken)
	exists := app.sessionManager.Exists(r.Context(), resetEmailSessionKey)

	data := app.newTemplateData(r)
	data.Page = authResetPage{HasSessionEmail: exists}
	app.render(w, http.StatusOK, "auth-reset.html", data)
}

type authResetForm struct {
	email    string
	password string
	validator.Validator
}

func (app *application) handleAuthResetPost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		clientError(w, http.StatusBadRequest)

		return
	}

	form := authResetForm{
		email:    r.Form.Get("email"),
		password: r.Form.Get("password"),
	}

	email := app.sessionManager.GetString(r.Context(), resetEmailSessionKey)
	app.infoLog.Println("sessionEmail:", email)
	if form.email != "" {
		email = form.email
	}

	form.Validate(validator.NotBlank(email), "invalid email: cannot be blank")
	form.Validate(validator.Matches(email, validator.EmailRX), "invalid email: must be a valid email address")
	form.Validate(validator.MaxChars(email, 254), "invalid email: must be no more than 254 characters long")
	form.Validate(validator.NotBlank(form.password), "invalid password: cannot be blank")
	form.Validate(validator.MinChars(form.password, 8), "invalid password: must be at least 8 characters long")
	form.Validate(validator.MaxChars(form.password, 72), "invalid password: must be no more than 72 characters long")

	if !form.IsValid() {
		validationError(w, form.Validator)

		return
	}

	token := app.sessionManager.GetString(r.Context(), resetTokenSessionKey)
	if token == "" {
		unauthorizedError(w)

		return
	}

	err = app.verifications.Verify(token, email)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			unauthorizedError(w)
		} else if errors.Is(err, models.ErrExpiredVerification) {
			app.flash(r, "Expired verification token.")

			http.Redirect(w, r, "/auth/forgot", http.StatusSeeOther)
		} else {
			app.serverError(w, err)
		}

		return
	}

	err = app.verifications.Purge(email)
	if err != nil {
		app.serverError(w, err)

		return
	}

	err = app.users.UpdatePassword(email, form.password)
	if err != nil {
		app.serverError(w, err)

		return
	}

	app.sessionManager.Clear(r.Context())

	app.flash(r, "Successfully reset password. Please login.")

	http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
}

func (app *application) handleAuthProfileGet(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "show user profile")
}

func (app *application) handleAuthProfilePost(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "update user profile")
}