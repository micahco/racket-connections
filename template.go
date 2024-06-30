package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"net/url"
	"path/filepath"
	"time"

	"github.com/justinas/nosurf"
	"github.com/micahco/racket-connections/ui"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type templateData struct {
	CurrentYear     int
	Flash           FlashMessage
	IsAuthenticated bool
	CSRFToken       string
	Data            interface{}
}

func (app *application) render(w http.ResponseWriter, r *http.Request, status int, page string, data interface{}) {
	td := templateData{
		CurrentYear:     time.Now().Year(),
		Flash:           app.popFlash(r),
		IsAuthenticated: app.isAuthenticated(r),
		CSRFToken:       nosurf.Token(r),
		Data:            data,
	}

	if app.isProduction {
		err := app.renderFromCache(w, status, page, td)
		if err != nil {
			app.serverError(w, err)
		}

		return
	}

	t, err := template.ParseFiles("./ui/html/base.html")
	if err != nil {
		app.serverError(w, err)

		return
	}

	t, err = t.Funcs(functions).ParseGlob("./ui/html/partials/*.html")
	if err != nil {
		app.serverError(w, err)

		return
	}

	t, err = t.ParseFiles("./ui/html/pages/" + page)
	if err != nil {
		app.serverError(w, err)

		return
	}

	err = t.ExecuteTemplate(w, "base", td)
	if err != nil {
		app.serverError(w, err)

		return
	}
}

func (app *application) renderFromCache(w http.ResponseWriter, status int, page string, data templateData) error {
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

func dateFormat(t time.Time) string {
	return t.Format("02 Jan 2006")
}

func humanDate(t time.Time) string {
	days := int(time.Since(t).Hours() / 24)
	if days == 0 {
		return "today"
	} else if days == 1 {
		return "1 day ago"
	} else if days < 7 {
		return fmt.Sprintf("%d days ago", days)
	} else if days == 7 {
		return "1 week ago"
	} else if days < 30 {
		return fmt.Sprintf("%d weeks", days%7)
	} else if days < 365 {
		return t.Format("02 January")
	}
	return dateFormat(t)
}

func capitalize(s string) string {
	c := cases.Title(language.English)
	return c.String(s)
}

var functions = template.FuncMap{
	"humanDate":   humanDate,
	"dateFormat":  dateFormat,
	"capitalize":  capitalize,
	"queryEscape": url.QueryEscape,
}

func newTemplateCache() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	pages, err := fs.Glob(ui.Files, "html/pages/*.html")
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		name := filepath.Base(page)

		patterns := []string{
			"html/base.html",
			"html/partials/*.html",
			page,
		}

		ts, err := template.New(name).Funcs(functions).ParseFS(ui.Files, patterns...)
		if err != nil {
			return nil, err
		}

		cache[name] = ts
	}

	return cache, nil
}
