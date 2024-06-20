package main

import (
	"fmt"
	"net/http"
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

func (app *application) flash(r *http.Request, message string) {
	app.sessionManager.Put(r.Context(), "flash", message)
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
