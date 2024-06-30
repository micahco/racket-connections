package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/micahco/racket-connections/internal/crypto"
	"github.com/micahco/racket-connections/internal/models"
	"github.com/micahco/racket-connections/internal/validator"
)

func (app *application) handleAuthLoginGet(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

type authLoginForm struct {
	email    string
	password string
	validator.Validator
}

func (app *application) handleAuthLoginPost(w http.ResponseWriter, r *http.Request) {
	if app.isAuthenticated(r) {
		clientError(w, http.StatusBadRequest)

		return
	}

	err := r.ParseForm()
	if err != nil {
		clientError(w, http.StatusBadRequest)

		return
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

	f := FlashMessage{
		Type:    FlashSuccess,
		Message: "Successfully logged out.",
	}
	app.flash(r, f)

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
	if app.isAuthenticated(r) {
		clientError(w, http.StatusBadRequest)

		return
	}

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

	// Consistent flash message
	f := FlashMessage{
		Type:    FlashInfo,
		Message: "A link to activate your account has been sent to the email address provided.",
	}

	// Don't send a new link if less than 5 minutes since last
	if v != nil {
		min := 5 * time.Minute
		if time.Since(v.CreatedAt) < min {
			app.flash(r, f)

			refresh(w, r)

			return
		}
	}

	token, err := crypto.GenerateRandomString(32)
	if err != nil {
		app.serverError(w, err)

		return
	}

	url := fmt.Sprintf("%s/auth/register?token=%s", app.url, token)
	html := fmt.Sprintf("<p>Please follow the link below to create your account:</p>"+
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

	app.flash(r, f)

	refresh(w, r)
}

type authRegisterData struct {
	HasSessionEmail bool
	ContactMethods  []*models.ContactMethod
	Days            []*models.DayOfWeek
	Times           []*models.TimeOfDay
}

func (app *application) handleAuthRegisterGet(w http.ResponseWriter, r *http.Request) {
	if app.isAuthenticated(r) {
		clientError(w, http.StatusBadRequest)

		return
	}

	queryToken := r.URL.Query().Get("token")
	if queryToken == "" {
		unauthorizedError(w)

		return
	}

	app.sessionManager.Put(r.Context(), verificationTokenSessionKey, queryToken)

	m, _ := app.contacts.Methods()
	d, _ := app.timeslots.Days()
	t, _ := app.timeslots.Times()

	exists := app.sessionManager.Exists(r.Context(), verificationEmailSessionKey)

	data := authRegisterData{
		HasSessionEmail: exists,
		ContactMethods:  m,
		Days:            d,
		Times:           t,
	}

	app.render(w, r, http.StatusOK, "auth-register.html", data)
}

type authRegisterForm struct {
	name          string
	email         string
	password      string
	contactMethod string
	contactValue  string
	validator.Validator
}

var ExpiredTokenFlash = FlashMessage{
	Type:    FlashError,
	Message: "Expired verification token.",
}

func (app *application) handleAuthRegisterPost(w http.ResponseWriter, r *http.Request) {
	if app.isAuthenticated(r) {
		clientError(w, http.StatusBadRequest)

		return
	}

	// Validate form
	err := r.ParseForm()
	if err != nil {
		clientError(w, http.StatusBadRequest)

		return
	}

	form := authRegisterForm{
		name:          r.Form.Get("name"),
		email:         r.Form.Get("email"),
		password:      r.Form.Get("password"),
		contactMethod: r.Form.Get("contact-method"),
		contactValue:  r.Form.Get("contact-value"),
	}

	email := app.sessionManager.GetString(r.Context(), verificationEmailSessionKey)
	if form.email != "" {
		email = form.email
	}

	form.Validate(validator.NotBlank(form.name), "invalid name: cannot be blank")
	form.Validate(validator.NotBlank(email), "invalid login email: cannot be blank")
	form.Validate(validator.Matches(email, validator.EmailRX), "invalid login email: must be a valid email address")
	form.Validate(validator.MaxChars(email, 254), "invalid login email: must be no more than 254 characters long")
	form.Validate(validator.PermittedEmailDomain(email, "oregonstate.edu"), "invalid login email: must be an OSU email address")
	form.Validate(validator.NotBlank(form.password), "invalid password: cannot be blank")
	form.Validate(validator.MinChars(form.password, 8), "invalid password: must be at least 8 characters long")
	form.Validate(validator.MaxChars(form.password, 72), "invalid password: must be no more than 72 characters long")
	form.Validate(validator.NotBlank(form.contactValue), "invalid contact value: cannot be blank")

	switch form.contactMethod {
	case "email":
		form.Validate(validator.Matches(form.contactValue, validator.EmailRX), "invalid contact email: must be a valid email address")
		form.Validate(validator.MaxChars(form.contactValue, 254), "invalid contact email: must be no more than 254 characters long")
	case "phone":
		form.Validate(validator.Matches(form.contactValue, validator.PhoneRX), "invalid contact phone: must be a valid phone number")
	case "other":
		// cleanse for anything malicous
	default:
		clientError(w, http.StatusBadRequest)

		return
	}

	if !form.IsValid() {
		validationError(w, form.Validator)

		return
	}

	// Verify token authentication
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
			app.flash(r, ExpiredTokenFlash)

			http.Redirect(w, r, "/", http.StatusSeeOther)
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

	// Insert data
	userID, err := app.users.Insert(form.name, email, form.password)
	if err != nil {
		if errors.Is(err, models.ErrDuplicateEmail) {
			unauthorizedError(w)
		} else {
			app.serverError(w, err)
		}

		return
	}

	methodID, err := app.contacts.MethodID(form.contactMethod)
	if err != nil {
		app.serverError(w, err)
	}

	err = app.contacts.Insert(form.contactValue, userID, methodID)
	if err != nil {
		app.serverError(w, err)
	}

	// Parse timetable and insert data
	days, _ := app.timeslots.Days()
	times, _ := app.timeslots.Times()
	for _, d := range days {
		for _, t := range times {
			key := fmt.Sprintf("%s-%s", d.Abbrev, t.Abbrev)
			if r.Form.Get(key) == "on" {
				err = app.timeslots.Insert(userID, d.ID, t.ID)
				if err != nil {
					app.serverError(w, err)
				}
			}
		}
	}

	// Login user
	app.sessionManager.Clear(r.Context())
	err = app.login(r, userID)
	if err != nil {
		app.serverError(w, err)

		return
	}

	f := FlashMessage{
		Type:    FlashSuccess,
		Message: "Successfully created account. Welcome!",
	}
	app.flash(r, f)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) handleAuthResetGet(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, http.StatusOK, "auth-reset.html", nil)
}

type authResetForm struct {
	email string
	validator.Validator
}

func (app *application) handleAuthResetPost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		clientError(w, http.StatusBadRequest)

		return
	}

	form := authResetForm{email: r.Form.Get("email")}

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

	f := FlashMessage{
		Type:    FlashInfo,
		Message: "A link to reset your password has been emailed to the address provided.",
	}

	if !exists {
		app.flash(r, f)

		refresh(w, r)

		return
	}

	token, err := crypto.GenerateRandomString(32)
	if err != nil {
		app.serverError(w, err)

		return
	}

	url := fmt.Sprintf("%s/auth/reset/update?token=%s", app.url, token)
	html := fmt.Sprintf("<p>Please follow the link below to reset your password:</p>"+
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

	app.flash(r, f)

	refresh(w, r)
}

type resetUpdateData struct {
	HasSessionEmail bool
}

func (app *application) handleAuthResetUpdateGet(w http.ResponseWriter, r *http.Request) {
	queryToken := r.URL.Query().Get("token")
	if queryToken == "" {
		unauthorizedError(w)

		return
	}

	app.sessionManager.Put(r.Context(), resetTokenSessionKey, queryToken)

	exists := app.sessionManager.Exists(r.Context(), resetEmailSessionKey)

	data := resetUpdateData{
		HasSessionEmail: exists,
	}

	app.render(w, r, http.StatusOK, "auth-reset-update.html", data)
}

type authResetUpdateForm struct {
	email    string
	password string
	validator.Validator
}

func (app *application) handleAuthResetUpdatePost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		clientError(w, http.StatusBadRequest)

		return
	}

	form := authResetUpdateForm{
		email:    r.Form.Get("email"),
		password: r.Form.Get("password"),
	}

	email := app.sessionManager.GetString(r.Context(), resetEmailSessionKey)
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
			app.flash(r, ExpiredTokenFlash)

			http.Redirect(w, r, "/auth/reset", http.StatusSeeOther)
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

	f := FlashMessage{
		Type:    FlashSuccess,
		Message: "Successfully updated password. Please login.",
	}
	app.flash(r, f)

	http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
}
