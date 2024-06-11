package main

import (
	"bytes"
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/micahco/racket-connections/internal/validator"
)

const (
	sessionAuthUserIDKey = "authenticatedUserID"
)

func (app *application) serverError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	app.errorLog.Output(2, trace)

	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func (app *application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

func (app *application) notFound(w http.ResponseWriter) {
	app.clientError(w, http.StatusNotFound)
}

func (app *application) permissionDenied(w http.ResponseWriter) {
	http.Error(w, "permission denied", http.StatusUnauthorized)
}

func (app *application) invalidRequest(w http.ResponseWriter, v validator.Validator) {
	http.Error(w, v.Errors(), http.StatusUnprocessableEntity)
}

// Data must be initialized with newTemplateData()
func (app *application) render(w http.ResponseWriter, status int, page string, data *templateData) {
	ts, ok := app.templateCache[page]
	if !ok {
		err := fmt.Errorf("template %s does not exist", page)
		app.serverError(w, err)
		return
	}

	buf := new(bytes.Buffer)

	err := ts.ExecuteTemplate(buf, "base", data)
	if err != nil {
		app.serverError(w, err)
		return
	}

	w.WriteHeader(status)

	if _, err := buf.WriteTo(w); err != nil {
		app.serverError(w, err)
	}
}

func (app *application) sessionLogin(r *http.Request, id int) error {
	err := app.sessionManager.RenewToken(r.Context())
	if err != nil {
		return err
	}

	app.sessionManager.Put(r.Context(), sessionAuthUserIDKey, id)

	return nil
}

func (app *application) sessionLogout(r *http.Request) error {
	err := app.sessionManager.RenewToken(r.Context())
	if err != nil {
		return err
	}

	app.sessionManager.Remove(r.Context(), sessionAuthUserIDKey)

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
