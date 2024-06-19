package main

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"path/filepath"
	"time"

	"github.com/justinas/nosurf"
)

type templateData struct {
	CurrentYear     int
	Flash           string
	IsAuthenticated bool
	CSRFToken       string
	Page            any
}

func (app *application) newTemplateData(r *http.Request) *templateData {
	return &templateData{
		CurrentYear:     time.Now().Year(),
		Flash:           app.sessionManager.PopString(r.Context(), "flash"),
		IsAuthenticated: app.isAuthenticated(r),
		CSRFToken:       nosurf.Token(r),
	}
}

// Data must be initialized with newTemplateData()
func (app *application) render(w http.ResponseWriter, status int, page string, data *templateData) {
	if data == nil {
		app.serverError(w, fmt.Errorf("passed nil data into render"))
	}

	if app.isProduction {
		err := app.renderFromCache(w, status, page, data)
		if err != nil {
			app.serverError(w, err)
		}

		return
	}

	t, err := template.ParseFiles("./templates/base.html")
	if err != nil {
		app.serverError(w, err)

		return
	}

	t, err = t.Funcs(functions).ParseGlob("./templates/partials/*.html")
	if err != nil {
		app.serverError(w, err)

		return
	}

	t, err = t.ParseFiles("./templates/pages/" + page)
	if err != nil {
		app.serverError(w, err)

		return
	}

	err = t.ExecuteTemplate(w, "base", data)
	if err != nil {
		app.serverError(w, err)

		return
	}
}

func (app *application) renderFromCache(w http.ResponseWriter, status int, page string, data *templateData) error {
	ts, ok := app.templateCache[page]
	if !ok {
		err := fmt.Errorf("template %s does not exist", page)
		return err
	}

	buf := new(bytes.Buffer)

	err := ts.ExecuteTemplate(buf, "base", data)
	if err != nil {
		return err
	}

	w.WriteHeader(status)

	if _, err := buf.WriteTo(w); err != nil {
		return err
	}

	return nil
}

func humanDate(t time.Time) string {
	return t.Format("02 Jan 2006 at 15:04")
}

var functions = template.FuncMap{
	"humanDate": humanDate,
}

//go:embed templates
var fsys embed.FS

func newTemplateCache() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	pages, err := fs.Glob(fsys, "templates/pages/*.html")
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		name := filepath.Base(page)

		patterns := []string{
			"templates/base.html",
			"templates/partials/*.html",
			page,
		}

		ts, err := template.New(name).Funcs(functions).ParseFS(fsys, patterns...)
		if err != nil {
			return nil, err
		}

		cache[name] = ts
	}

	return cache, nil
}
