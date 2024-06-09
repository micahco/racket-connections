package main

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func (app *application) handleGetIndex(w http.ResponseWriter, r *http.Request) {
	err := app.renderTemplate(w, "welcome.html", nil)
	if err != nil {
		panic(err)
	}
}

func (app *application) handleGetLogin(w http.ResponseWriter, r *http.Request) {
	err := app.renderTemplate(w, "login.html", nil)
	if err != nil {
		panic(err)
	}
}

type loginRequest struct {
	email    string
	password string
}

func (app *application) handlePostLogin(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		panic(httpError{err, http.StatusBadRequest})
	}

	var req loginRequest
	req.email = r.Form.Get("email")
	req.password = r.Form.Get("password")

	user, err := app.store.getUserByEmail(req.email)
	if err != nil {
		err = errors.New("permission denied")
		panic(httpError{err, http.StatusUnauthorized})
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.passwordHash), []byte(req.password))
	if err != nil {
		err = errors.New("permission denied")
		panic(httpError{err, http.StatusUnauthorized})
	}

	err = app.session.RenewToken(r.Context())
	if err != nil {
		panic(httpError{err, http.StatusInternalServerError})
	}

	app.session.Put(r.Context(), "userID", user.id)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

type signupRequest struct {
	name     string
	email    string
	password string
}

const (
	OSU_DOMAIN       = "oregonstate.edu"
	MIN_PASSWORD_LEN = 8
	MAX_PASSWORD_LEN = 72
)

func validateSignupRequest(req signupRequest) error {
	_, domain, found := strings.Cut(req.email, "@")
	if !found || domain != OSU_DOMAIN {
		return errors.New("invalid domain")
	}

	if len(req.password) < MIN_PASSWORD_LEN {
		return fmt.Errorf("password too short (min: %d)", MIN_PASSWORD_LEN)
	}

	if len(req.password) > MAX_PASSWORD_LEN {
		return fmt.Errorf("password too long (max: %d)", MAX_PASSWORD_LEN)
	}

	return nil
}

func (app *application) handleGetSignup(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func generateRandomToken(length int) (string, error) {
	b := make([]byte, length)

	// Read random bytes from the cryptographic random number generator
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	// Encode random bytes to base64 string
	token := base64.URLEncoding.EncodeToString(b)

	// Return the token
	return token, nil
}

func (app *application) handlePostSignup(w http.ResponseWriter, r *http.Request) {
	// Parse from to signup request
	err := r.ParseForm()
	if err != nil {
		panic(httpError{err, http.StatusBadRequest})
	}

	var req signupRequest
	req.name = r.Form.Get("name")
	req.email = r.Form.Get("email")
	req.password = r.Form.Get("password")

	if err := validateSignupRequest(req); err != nil {
		panic(httpError{err, http.StatusUnauthorized})
	}

	// Create user with signup request
	user, err := app.store.createUser(req)
	if err != nil {
		err = errors.New("permission denied")
		panic(httpError{err, http.StatusUnauthorized})
	}

	// Generate email verification token
	token, err := generateRandomToken(32)
	if err != nil {
		panic(httpError{err, http.StatusInternalServerError})
	}

	expiry := time.Now().Add(5 * time.Minute)
	v, err := app.store.createVerification(user.email, token, expiry)
	if err != nil {
		panic(httpError{err, http.StatusInternalServerError})
	}

	// send email to user with token
	fmt.Println("email:", v.email, "\ntoken:", v.token)

	app.session.Put(r.Context(), "userID", user.id)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) handleGetLogout(w http.ResponseWriter, r *http.Request) {
	app.session.Remove(r.Context(), "userID")

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) handleGetForgot(w http.ResponseWriter, r *http.Request) {
	err := app.renderTemplate(w, "forgot.html", nil)
	if err != nil {
		panic(err)
	}
}
