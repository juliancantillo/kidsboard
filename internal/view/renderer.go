package view

import (
	"embed"
	"fmt"
	"html/template"
	"io"
	"path"
	"strings"
)

//go:embed templates
var templatesFS embed.FS

// Renderer wraps html/template with a layout/page convention. Each "page" is
// the body content; every render uses the single base layout in
// templates/layouts/base.html. Partials under templates/partials/ are
// available to all pages by name.
type Renderer struct {
	pages map[string]*template.Template
}

// NewRenderer parses every page in templates/pages/ paired with the base
// layout and all partials. Returns an error if any template fails to parse —
// fail-fast at boot so production renders never panic on bad templates.
func NewRenderer() (*Renderer, error) {
	pageFiles, err := listPageFiles()
	if err != nil {
		return nil, err
	}
	partialFiles, err := listFiles("templates/partials")
	if err != nil {
		return nil, err
	}

	r := &Renderer{pages: make(map[string]*template.Template, len(pageFiles))}
	for _, pageFile := range pageFiles {
		name := pageNameFromFile(pageFile)
		tmpl := template.New("base.html").Funcs(funcMap())
		files := append([]string{"templates/layouts/base.html", pageFile}, partialFiles...)
		parsed, err := tmpl.ParseFS(templatesFS, files...)
		if err != nil {
			return nil, fmt.Errorf("parse %q: %w", pageFile, err)
		}
		r.pages[name] = parsed
	}
	return r, nil
}

// Render writes the named page (e.g. "home", "profile") wrapped in the base
// layout. Data is the page-specific view model; the layout receives it as `.Data`.
func (r *Renderer) Render(w io.Writer, page string, data any) error {
	t, ok := r.pages[page]
	if !ok {
		return fmt.Errorf("unknown page %q", page)
	}
	return t.ExecuteTemplate(w, "base.html", pageContext{Page: page, Data: data})
}

type pageContext struct {
	Page string
	Data any
}

func listPageFiles() ([]string, error) {
	entries, err := templatesFS.ReadDir("templates/pages")
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".html") {
			continue
		}
		out = append(out, path.Join("templates/pages", e.Name()))
	}
	return out, nil
}

func listFiles(dir string) ([]string, error) {
	entries, err := templatesFS.ReadDir(dir)
	if err != nil {
		// Partials are optional.
		return nil, nil
	}
	out := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".html") {
			continue
		}
		out = append(out, path.Join(dir, e.Name()))
	}
	return out, nil
}

func pageNameFromFile(p string) string {
	base := path.Base(p)
	return strings.TrimSuffix(base, ".html")
}
