package main

import (
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"

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

	fs := http.FileServer(http.Dir("./static/"))
	r.Handle("/static/*", http.StripPrefix("/static", fs))

	r.Get("/favicon.ico", handleFavicon)

	r.Route("/", func(r chi.Router) {
		r.Use(app.sessionManager.LoadAndSave)
		r.Use(noSurf)
		r.Use(app.authenticate)

		r.Get("/", app.handleIndexGet)

		r.Route("/user", func(r chi.Router) {
			r.Get("/login", app.handleUserLoginGet)
			r.Post("/login", app.handleUserLoginPost)

			r.Post("/logout", app.handleUserLogoutPost)

			r.Get("/signup", app.handleUserSignupGet)
			r.Post("/signup", app.handleUserSignupPost)

			r.Get("/create", app.handleUserCreateGet)
			r.Post("/create", app.handleUserCreatePost)

			r.Get("/forgot", app.handleUserForgotGet)
			r.Post("/forgot", app.handleUserForgotPost)

			r.Get("/reset", app.handleUserResetGet)
			r.Post("/reset", app.handleUserResetPost)

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

func handleFavicon(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./static/favicon.ico")
}

func (app *application) handleIndexGet(w http.ResponseWriter, r *http.Request) {
	app.render(w, http.StatusOK, "index.html", app.newTemplateData(r))
}

func (app *application) handleUserLoginGet(w http.ResponseWriter, r *http.Request) {
	if app.isAuthenticated(r) {
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	app.render(w, http.StatusOK, "user-login.html", app.newTemplateData(r))
}

type userLoginForm struct {
	email    string
	password string
	validator.Validator
}

func (app *application) handleUserLoginPost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		clientError(w, http.StatusBadRequest)
	}

	form := userLoginForm{
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

func (app *application) handleUserLogoutPost(w http.ResponseWriter, r *http.Request) {
	err := app.logout(r)
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.flash(r, "You have been logged out successfully!")

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) handleUserSignupGet(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

type userSignupForm struct {
	email string
	validator.Validator
}

func (app *application) handleUserSignupPost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		clientError(w, http.StatusBadRequest)

		return
	}

	form := userSignupForm{email: r.Form.Get("email")}

	form.Validate(validator.NotBlank(form.email), "invalid email: cannot be blank")
	form.Validate(validator.Matches(form.email, validator.EmailRX), "invalid email: must be a valid email address")
	form.Validate(validator.MaxChars(form.email, 254), "invalid email: must be no more than 254 characters long")
	//form.Validate(validator.PermittedEmailDomain(form.email, "oregonstate.edu"), "invalid email: must be an OSU email address")

	if !form.IsValid() {
		validationError(w, form.Validator)

		return
	}

	token, err := crypto.GenerateRandomString(32)
	if err != nil {
		app.serverError(w, err)

		return
	}

	url := fmt.Sprintf("%s/user/create?token=%s", app.url, token)
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

type userCreatePage struct {
	HasSessionEmail bool
}

func (app *application) handleUserCreateGet(w http.ResponseWriter, r *http.Request) {
	queryToken := r.URL.Query().Get("token")
	if queryToken == "" {
		unauthorizedError(w)

		return
	}

	app.sessionManager.Put(r.Context(), verificationTokenSessionKey, queryToken)
	exists := app.sessionManager.Exists(r.Context(), verificationEmailSessionKey)

	data := app.newTemplateData(r)
	data.Page = userCreatePage{HasSessionEmail: exists}
	app.render(w, http.StatusOK, "user-create.html", data)
}

type userCreateForm struct {
	name     string
	email    string
	password string
	validator.Validator
}

func (app *application) handleUserCreatePost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		clientError(w, http.StatusBadRequest)

		return
	}

	form := userCreateForm{
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

			http.Redirect(w, r, "/user/login", http.StatusSeeOther)
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

func (app *application) handleUserForgotGet(w http.ResponseWriter, r *http.Request) {
	app.render(w, http.StatusOK, "user-forgot.html", app.newTemplateData(r))
}

type userForgotForm struct {
	email string
	validator.Validator
}

func (app *application) handleUserForgotPost(w http.ResponseWriter, r *http.Request) {
	msg := "A link to reset your password has been emailed to the address provided."

	err := r.ParseForm()
	if err != nil {
		clientError(w, http.StatusBadRequest)

		return
	}

	form := userForgotForm{email: r.Form.Get("email")}

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

		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	token, err := crypto.GenerateRandomString(32)
	if err != nil {
		app.serverError(w, err)

		return
	}

	url := fmt.Sprintf("%s/user/reset?token=%s", app.url, token)
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

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

type userResetPage struct {
	HasSessionEmail bool
}

func (app *application) handleUserResetGet(w http.ResponseWriter, r *http.Request) {
	queryToken := r.URL.Query().Get("token")
	if queryToken == "" {
		unauthorizedError(w)

		return
	}

	app.sessionManager.Put(r.Context(), resetTokenSessionKey, queryToken)
	exists := app.sessionManager.Exists(r.Context(), resetEmailSessionKey)

	data := app.newTemplateData(r)
	data.Page = userResetPage{HasSessionEmail: exists}
	app.render(w, http.StatusOK, "user-reset.html", data)
}

type userResetForm struct {
	email    string
	password string
	validator.Validator
}

func (app *application) handleUserResetPost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		clientError(w, http.StatusBadRequest)

		return
	}

	form := userResetForm{
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
			app.flash(r, "Your verification token is expired. Please resubmit the form.")

			http.Redirect(w, r, "/user/forgot", http.StatusSeeOther)
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

	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

func (app *application) handleUserProfileGet(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "show user profile")
}

func (app *application) handleUserProfilePost(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "update user profile")
}
