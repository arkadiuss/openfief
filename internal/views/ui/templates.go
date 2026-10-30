package ui

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

// webFS holds the page templates and the static files the UI serves:
// vendored htmx and Chart.js, its own scripts and its stylesheet.
//
//go:embed web
var webFS embed.FS

// errNoPage reports a page name with no template.
var errNoPage = errors.New("no such page template")

// pages holds one parsed template set per page, each the layout plus the
// page's own content and the shared fragments.
type pages struct {
	sets map[string]*template.Template
}

// page is what every page template gets: its title, the session (nil on
// the login and enrollment pages) and the page's own data.
type page struct {
	Title   string
	Session *session
	Data    any
}

// templateFuncs are the helpers templates format values with, amounts in
// the number format f (D-113).
func templateFuncs(f numberFormat) template.FuncMap {
	return template.FuncMap{
		"money":        f.money,
		"amount":       f.amount,
		"share":        share,
		"percent":      percent,
		"book":         bookLabel,
		"instrument":   instrumentLabel,
		"enum":         enumLabel,
		"balance":      f.balance,
		"missing":      missingLabel,
		"wrapper":      wrapperLabel,
		"feeRule":      f.feeRule,
		"trust":        trustLabel,
		"badge":        badgeOf,
		"roleFields":   roleFields,
		"sleeveSelect": sleeveSelect,
		"terms":        termsFields,
	}
}

// parsePages parses every page under web/templates/pages with the layout
// and the fragments.
func parsePages() (*pages, error) {
	base, err := template.New("layout.html").Funcs(templateFuncs(plainFormat())).ParseFS(webFS,
		"web/templates/layout.html", "web/templates/fragments/*.html")
	if err != nil {
		return nil, fmt.Errorf("parse layout: %w", err)
	}
	names, err := fs.Glob(webFS, "web/templates/pages/*.html")
	if err != nil {
		return nil, fmt.Errorf("list pages: %w", err)
	}
	p := &pages{sets: map[string]*template.Template{}}
	for _, name := range names {
		set, cloneErr := base.Clone()
		if cloneErr != nil {
			return nil, fmt.Errorf("clone layout: %w", cloneErr)
		}
		if _, parseErr := set.ParseFS(webFS, name); parseErr != nil {
			return nil, fmt.Errorf("parse %s: %w", name, parseErr)
		}
		p.sets[strings.TrimSuffix(path.Base(name), ".html")] = set
	}
	return p, nil
}

// render writes a template of a page with status, its amounts in the
// number format f, buffered so a template error never sends half a page:
// tmpl is "layout.html" for the whole page or a fragment's name for an
// htmx swap. Each render executes a clone of the parsed set with f's
// helpers, so the parsed set is never executed and stays clonable.
func (p *pages) render(w http.ResponseWriter, status int, name, tmpl string, data any, f numberFormat) error {
	parsed, ok := p.sets[name]
	if !ok {
		return fmt.Errorf("%w: %s", errNoPage, name)
	}
	set, err := parsed.Clone()
	if err != nil {
		return fmt.Errorf("clone %s: %w", name, err)
	}
	var buf bytes.Buffer
	if err := set.Funcs(templateFuncs(f)).ExecuteTemplate(&buf, tmpl, data); err != nil {
		return fmt.Errorf("render %s: %w", name, err)
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	if _, err := buf.WriteTo(w); err != nil {
		return fmt.Errorf("write %s: %w", name, err)
	}
	return nil
}

// show renders a whole page, or a plain 500 when the template fails.
func (s *server) show(w http.ResponseWriter, r *http.Request, status int, name string, data page) {
	s.showTemplate(w, r, status, name, "layout.html", data)
}

// showTemplate renders one template of a page, with amounts in the number
// format the request's session read (plainFormat() without one), or a plain
// 500 when it fails.
func (s *server) showTemplate(w http.ResponseWriter, r *http.Request, status int, name, tmpl string, data any) {
	if err := s.pages.render(w, status, name, tmpl, data, requestFormat(r)); err != nil {
		s.log.ErrorContext(r.Context(), "render page", "page", name, "error", err.Error())
		http.Error(w, "the page could not be rendered", http.StatusInternalServerError)
	}
}

// errorPage renders the error page with status and a message for the
// owner. It never names a token or an amount.
func (s *server) errorPage(w http.ResponseWriter, r *http.Request, status int, message string) {
	s.show(w, r, status, "error", page{Title: http.StatusText(status), Data: message})
}
