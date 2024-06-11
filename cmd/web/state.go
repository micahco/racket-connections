package main

import "net/http"

type contextKey string

const (
	authenticatedUserIDSessionKey = "authenticatedUserID"
	isAuthenticatedContextKey     = contextKey("isAuthenticated")
)

func (app *application) login(r *http.Request, id int) error {
	err := app.sessionManager.RenewToken(r.Context())
	if err != nil {
		return err
	}

	app.sessionManager.Put(r.Context(), string(authenticatedUserIDSessionKey), id)

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
