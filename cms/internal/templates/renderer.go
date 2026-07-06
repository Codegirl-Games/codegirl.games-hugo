package templates

import (
	"embed"
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"time"
)

//go:embed all:templates
var templateFS embed.FS

//go:embed all:static
var staticFS embed.FS

// Renderer renders HTML templates with shared layout and functions.
type Renderer struct {
	templates *template.Template
}

// New creates a template renderer.
func New() (*Renderer, error) {
	funcs := template.FuncMap{
		"formatDate": formatDate,
		"draftBadge": draftBadgeClass,
	}

	tmpl, err := template.New("").Funcs(funcs).ParseFS(templateFS, "templates/*.html")
	if err != nil {
		return nil, err
	}

	return &Renderer{templates: tmpl}, nil
}

// Render executes a named template with data.
func (r *Renderer) Render(w io.Writer, name string, data any) error {
	return r.templates.ExecuteTemplate(w, name, data)
}

// StaticHandler serves embedded static assets.
func StaticHandler() http.Handler {
	sub, err := fs.Sub(staticFS, "static")
	if err != nil {
		panic(err)
	}
	return http.FileServer(http.FS(sub))
}

func formatDate(t any) string {
	switch v := t.(type) {
	case time.Time:
		if v.IsZero() {
			return "—"
		}
		return v.Format("Jan 2, 2006")
	case string:
		return v
	default:
		return "—"
	}
}

func draftBadgeClass(draft bool) string {
	if draft {
		return "badge-draft"
	}
	return "badge-published"
}
