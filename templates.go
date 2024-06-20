package main

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"net/url"
	"path/filepath"
	"time"

	"github.com/justinas/nosurf"
	"github.com/micahco/racket-connections/internal/models"
)

type templateQueries struct {
	Sport string
	ID    int
}

type templateData struct {
	CurrentYear     int
	Flash           string
	IsAuthenticated bool
	CSRFToken       string
	Queries         *templateQueries
	HasSessionEmail bool
	Post            *models.PostDetails
	Posts           []*models.PostDetails
	Skills          []*models.SkillLevel
	Sports          []*models.Sport
	SportsPostsMap  map[int][]*models.PostDetails
}

func (app *application) newTemplateData(r *http.Request) *templateData {
	return &templateData{
		CurrentYear:     time.Now().Year(),
		Flash:           app.sessionManager.PopString(r.Context(), "flash"),
		IsAuthenticated: app.isAuthenticated(r),
		CSRFToken:       nosurf.Token(r),
		Queries:         &templateQueries{},
		HasSessionEmail: false,
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

func daysAgo(t time.Time) int {
	return int(time.Since(t).Hours() / 24)
}

func isToday(t time.Time) bool {
	return daysAgo(t) == 0
}

var functions = template.FuncMap{
	"daysAgo":     daysAgo,
	"isToday":     isToday,
	"queryEscape": url.QueryEscape,
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
