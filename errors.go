package main

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/micahco/racket-connections/internal/validator"
)

type errorPageData struct {
	Code       int
	StatusText string
	Message    string
}

func (app *application) renderError(w http.ResponseWriter, r *http.Request, status int, message string) {
	data := errorPageData{
		Code:       status,
		StatusText: http.StatusText(status),
		Message:    message,
	}

	app.render(w, r, status, "error.html", data)
}

func (app *application) serverError(w http.ResponseWriter, r *http.Request, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	app.errorLog.Output(2, trace)

	app.renderError(w, r, http.StatusInternalServerError, "")
}

func unauthorizedError(w http.ResponseWriter) {
	http.Error(w, "permission denied", http.StatusUnauthorized)
}

func validationError(w http.ResponseWriter, v validator.Validator) {
	http.Error(w, v.Errors(), http.StatusUnprocessableEntity)
}
