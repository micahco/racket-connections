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

type contextKey string

const (
	authenticatedUserIDSessionKey = "authenticatedUserID"
	verificationEmailSessionKey   = "verificationEmail"
	verificationTokenSessionKey   = "verificationToken"
	resetEmailSessionKey          = "resetEmail"
	resetTokenSessionKey          = "resetToken"
	isAuthenticatedContextKey     = contextKey("isAuthenticated")
)

func (app *application) login(r *http.Request, userID int) error {
	err := app.sessionManager.RenewToken(r.Context())
	if err != nil {
		return err
	}

	app.sessionManager.Put(r.Context(), string(authenticatedUserIDSessionKey), userID)

	return nil
}

func (app *application) logout(r *http.Request) error {
	err := app.sessionManager.RenewToken(r.Context())
	if err != nil {
		return err
	}

	app.sessionManager.Remove(r.Context(), string(authenticatedUserIDSessionKey))

	return nil
}

func (app *application) isAuthenticated(r *http.Request) bool {
	isAuthenticated, ok := r.Context().Value(isAuthenticatedContextKey).(bool)
	if !ok {
		return false
	}

	return isAuthenticated
}

func (app *application) getSessionUserID(r *http.Request) (int, error) {
	id, ok := app.sessionManager.Get(r.Context(), authenticatedUserIDSessionKey).(int)
	if !ok {
		return 0, fmt.Errorf("type assertion to int failed")
	}

	return id, nil
}

type authLoginForm struct {
	email    string
	password string
	validator.Validator
}

func (app *application) handleAuthLoginPost(w http.ResponseWriter, r *http.Request) {
	if app.isAuthenticated(r) {
		app.renderError(w, r, http.StatusBadRequest, "")

		return
	}

	err := r.ParseForm()
	if err != nil {
		app.renderError(w, r, http.StatusBadRequest, "")

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

	id, err := app.models.User.Authenticate(form.email, form.password)
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			unauthorizedError(w)
		} else {
			app.serverError(w, r, err)
		}

		return
	}

	err = app.login(r, id)
	if err != nil {
		app.serverError(w, r, err)

		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) handleAuthLogoutPost(w http.ResponseWriter, r *http.Request) {
	err := app.logout(r)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

type authSignupForm struct {
	email string
	validator.Validator
}

func (app *application) handleAuthSignupPost(w http.ResponseWriter, r *http.Request) {
	if app.isAuthenticated(r) {
		app.renderError(w, r, http.StatusBadRequest, "")

		return
	}

	err := r.ParseForm()
	if err != nil {
		app.renderError(w, r, http.StatusBadRequest, "")

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

	// Consistent flash message
	f := FlashMessage{
		Type:    FlashInfo,
		Message: "A link to activate your account has been sent to the email address provided. Please check your junk folder.",
	}

	// Check if user with email already exists
	exists, err := app.models.User.ExistsEmail(form.email)
	if err != nil {
		app.serverError(w, r, err)

		return
	}

	// Don't send any email if user with email already exists
	if exists {
		app.flash(r, f)

		refresh(w, r)

		return
	}

	// Check if link verification has already been created
	v, err := app.models.Verification.Get(form.email)
	if err != nil && err != models.ErrNoRecord {
		app.serverError(w, r, err)

		return
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
		app.serverError(w, r, err)

		return
	}

	link := fmt.Sprintf("%s/auth/register?token=%s", app.baseURL, token)

	if app.isDevelopment {
		fmt.Println("Verification link:", link)
	}

	app.background(func() {
		err = app.mailer.Send(form.email, "email_verification.tmpl", link)
		if err != nil {
			app.errorLog.Println(err)
		}
	})

	err = app.models.Verification.Insert(token, form.email)
	if err != nil {
		app.serverError(w, r, err)

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
		app.renderError(w, r, http.StatusBadRequest, "")

		return
	}

	queryToken := r.URL.Query().Get("token")
	if queryToken == "" {
		unauthorizedError(w)

		return
	}

	app.sessionManager.Put(r.Context(), verificationTokenSessionKey, queryToken)

	m, _ := app.models.Contact.Methods()
	d, _ := app.models.Timeslot.Days()
	t, _ := app.models.Timeslot.Times()

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
		app.renderError(w, r, http.StatusBadRequest, "")

		return
	}

	// Validate form
	err := r.ParseForm()
	if err != nil {
		app.renderError(w, r, http.StatusBadRequest, "")

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
		app.renderError(w, r, http.StatusBadRequest, "")

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

	err = app.models.Verification.Verify(token, email)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			unauthorizedError(w)
		} else if errors.Is(err, models.ErrExpiredVerification) {
			app.flash(r, ExpiredTokenFlash)

			http.Redirect(w, r, "/", http.StatusSeeOther)
		} else {
			app.serverError(w, r, err)
		}

		return
	}

	err = app.models.Verification.Purge(email)
	if err != nil {
		app.serverError(w, r, err)

		return
	}

	// Insert data
	userID, err := app.models.User.Insert(form.name, email, form.password)
	if err != nil {
		if errors.Is(err, models.ErrDuplicateEmail) {
			unauthorizedError(w)
		} else {
			app.serverError(w, r, err)
		}

		return
	}

	methodID, err := app.models.Contact.MethodID(form.contactMethod)
	if err != nil {
		app.serverError(w, r, err)
	}

	err = app.models.Contact.Insert(form.contactValue, userID, methodID)
	if err != nil {
		app.serverError(w, r, err)
	}

	// Parse timetable and insert data
	days, _ := app.models.Timeslot.Days()
	times, _ := app.models.Timeslot.Times()
	for _, d := range days {
		for _, t := range times {
			key := fmt.Sprintf("%s-%s", d.Abbrev, t.Abbrev)
			if r.Form.Get(key) == "on" {
				err = app.models.Timeslot.Insert(userID, d.ID, t.ID)
				if err != nil {
					app.serverError(w, r, err)
				}
			}
		}
	}

	// Login user
	app.sessionManager.Clear(r.Context())
	err = app.login(r, userID)
	if err != nil {
		app.serverError(w, r, err)

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
		app.renderError(w, r, http.StatusBadRequest, "")

		return
	}

	form := authResetForm{email: r.Form.Get("email")}

	form.Validate(validator.NotBlank(form.email), "invalid email: cannot be blank")
	form.Validate(validator.MaxChars(form.email, 254), "invalid email: must be no more than 254 characters long")

	if !form.IsValid() {
		validationError(w, form.Validator)

		return
	}

	exists, err := app.models.User.ExistsEmail(form.email)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	f := FlashMessage{
		Type:    FlashInfo,
		Message: "A link to reset your password has been sent to the email address provided. Please check your junk folder.",
	}

	if !exists {
		app.flash(r, f)

		refresh(w, r)

		return
	}

	token, err := crypto.GenerateRandomString(32)
	if err != nil {
		app.serverError(w, r, err)

		return
	}

	link := fmt.Sprintf("%s/auth/reset/update?token=%s", app.baseURL, token)

	if app.isDevelopment {
		fmt.Println("Reset link:", link)
	}

	app.background(func() {
		err = app.mailer.Send(form.email, "reset_password.tmpl", link)
		if err != nil {
			app.errorLog.Println(err)
		}
	})

	err = app.models.Verification.Insert(token, form.email)
	if err != nil {
		app.serverError(w, r, err)

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
		app.renderError(w, r, http.StatusBadRequest, "")

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

	err = app.models.Verification.Verify(token, email)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			unauthorizedError(w)
		} else if errors.Is(err, models.ErrExpiredVerification) {
			app.flash(r, ExpiredTokenFlash)

			http.Redirect(w, r, "/auth/reset", http.StatusSeeOther)
		} else {
			app.serverError(w, r, err)
		}

		return
	}

	err = app.models.Verification.Purge(email)
	if err != nil {
		app.serverError(w, r, err)

		return
	}

	err = app.models.User.UpdatePassword(email, form.password)
	if err != nil {
		app.serverError(w, r, err)

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
