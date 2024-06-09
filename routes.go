package main

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const (
	BASE_TEMPLATE       = "base.html"
	MIN_PASSWORD_LEN    = 8
	MAX_PASSWORD_LEN    = 72
	SESSION_COOKIE_NAME = "session_token"
)

// render nested template (filename) in config.templateDir with base template
func (app *application) renderTemplate(w http.ResponseWriter, filename string, data any) error {
	base := filepath.Join(app.config.templatesDir, BASE_TEMPLATE)
	tmpl := filepath.Join(app.config.templatesDir, filename)

	t, err := template.ParseFiles(base, tmpl)
	if err != nil {
		return fmt.Errorf("unable to parse template: %w", err)
	}

	return t.Execute(w, data)
}

func (app *application) readSession(r *http.Request) error {
	// Read session id from cookie
	c, err := r.Cookie(SESSION_COOKIE_NAME)
	if err != nil {
		return err
	}
	sessionID := c.Value

	// Get session from database
	userSession, err := app.store.getSession(sessionID)
	if err != nil {
		return err
	}

	if userSession.isExpired() {
		return errors.New("session expired")
	}

	return nil
}

func (app *application) writeSession(w http.ResponseWriter, u *user) {
	// Create new session id and expiry
	sesionID := uuid.NewString()
	expiry := time.Now().Add(time.Hour)

	// Save session to database
	app.store.createSession(sesionID, u.id, expiry)

	// Add set cookie header with new session id
	http.SetCookie(w, &http.Cookie{
		Name:    SESSION_COOKIE_NAME,
		Value:   sesionID,
		Path:    "/",
		Expires: expiry,
	})
}

// GET /
func (app *application) handleGetIndex(w http.ResponseWriter, r *http.Request) {
	if err := app.readSession(r); err != nil {
		err := app.renderTemplate(w, "welcome.html", nil)
		if err != nil {
			panic(err)
		}

		return
	}

	// show posts
}

// GET /login
func (app *application) handleGetLogin(w http.ResponseWriter, r *http.Request) {
	if err := app.readSession(r); err != nil {
		err := app.renderTemplate(w, "login.html", nil)
		if err != nil {
			panic(err)
		}

		return
	}

	// Already authenticated. Redirect to index
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

type loginRequest struct {
	email    string
	password string
}

// POST /login
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

	app.writeSession(w, user)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

type signupRequest struct {
	name     string
	email    string
	password string
}

// POST /signup
func (app *application) handleSignupPost(w http.ResponseWriter, r *http.Request) {
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

	hash, err := bcrypt.GenerateFromPassword([]byte(req.password), bcrypt.DefaultCost)
	if err != nil {
		err = errors.New("permission denied")
		panic(httpError{err, http.StatusUnauthorized})
	}

	user, err := app.store.createUser(req, hash)
	if err != nil {
		err = errors.New("permission denied")
		panic(httpError{err, http.StatusUnauthorized})
	}

	app.writeSession(w, user)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func validateSignupRequest(req signupRequest) error {
	_, after, found := strings.Cut(req.email, "@")
	if !found || after != "oregonstate.edu" {
		return errors.New("permission denied")
	}

	if len(req.password) < MIN_PASSWORD_LEN {
		return fmt.Errorf("password too short (min: %d)", MIN_PASSWORD_LEN)
	}

	if len(req.password) > MIN_PASSWORD_LEN {
		return fmt.Errorf("password too long (max: %d)", MIN_PASSWORD_LEN)
	}

	return nil
}
