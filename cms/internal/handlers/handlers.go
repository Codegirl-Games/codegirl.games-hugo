package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/codegirl-007/hugo-cms/internal/auth"
	"github.com/codegirl-007/hugo-cms/internal/media"
	"github.com/codegirl-007/hugo-cms/internal/posts"
	"github.com/codegirl-007/hugo-cms/internal/session"
	"github.com/codegirl-007/hugo-cms/internal/templates"
)

// Handler holds HTTP dependencies.
type Handler struct {
	auth      *auth.Authenticator
	sessions  *session.Store
	posts     *posts.Service
	media     *media.Service
	renderer  *templates.Renderer
}

// New creates an HTTP handler.
func New(
	auth *auth.Authenticator,
	sessions *session.Store,
	posts *posts.Service,
	media *media.Service,
	renderer *templates.Renderer,
) *Handler {
	return &Handler{
		auth:     auth,
		sessions: sessions,
		posts:    posts,
		media:    media,
		renderer: renderer,
	}
}

// Auth wraps a handler with session authentication.
func (h *Handler) Auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !h.sessions.IsAuthenticated(r) {
			if strings.HasPrefix(r.URL.Path, "/api/") {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
				return
			}
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next(w, r)
	}
}

// RequireAuth redirects unauthenticated users to /login.
func (h *Handler) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !h.sessions.IsAuthenticated(r) {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (h *Handler) LoginPage(w http.ResponseWriter, r *http.Request) {
	if h.sessions.IsAuthenticated(r) {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	data := map[string]any{
		"Title": "Login",
		"Error": r.URL.Query().Get("error"),
	}
	h.render(w, "login.html", data)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/login?error=invalid", http.StatusSeeOther)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	if err := h.auth.Verify(username, password); err != nil {
		http.Redirect(w, r, "/login?error=invalid", http.StatusSeeOther)
		return
	}

	if err := h.sessions.Set(w, &session.Data{Username: username}); err != nil {
		http.Error(w, "session error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	h.sessions.Clear(w)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (h *Handler) Dashboard(w http.ResponseWriter, r *http.Request) {
	allPosts, err := h.posts.List(r.Context())
	if err != nil {
		h.renderError(w, "Failed to load posts", err)
		return
	}

	stats := posts.ComputeStats(allPosts)
	recent := allPosts
	if len(recent) > 5 {
		recent = recent[:5]
	}

	data := map[string]any{
		"Title":      "Dashboard",
		"Active":     "dashboard",
		"Stats":      stats,
		"RecentPosts": recent,
	}
	h.render(w, "dashboard.html", data)
}

func (h *Handler) PostsList(w http.ResponseWriter, r *http.Request) {
	allPosts, err := h.posts.List(r.Context())
	if err != nil {
		h.renderError(w, "Failed to load posts", err)
		return
	}

	query := r.URL.Query().Get("q")
	filtered := posts.FilterByTitle(allPosts, query)

	data := map[string]any{
		"Title":  "Posts",
		"Active": "posts",
		"Posts":  filtered,
		"Query":  query,
	}
	h.render(w, "posts_list.html", data)
}

func (h *Handler) PostNew(w http.ResponseWriter, r *http.Request) {
	now := time.Now().UTC().Format("2006-01-02T15:04")
	data := map[string]any{
		"Title":    "New Post",
		"Active":   "posts",
		"IsNew":    true,
		"Post":     map[string]any{},
		"Date":     now,
		"Draft":    true,
		"Tags":     "",
		"Body":     "",
		"Slug":     "",
		"Original": "",
	}
	h.render(w, "post_edit.html", data)
}

func (h *Handler) PostEdit(w http.ResponseWriter, r *http.Request) {
	slug := strings.TrimPrefix(r.URL.Path, "/admin/posts/")
	if slug == "" || slug == "new" {
		http.NotFound(w, r)
		return
	}

	post, err := h.posts.Get(r.Context(), slug)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	dateStr := post.Date.UTC().Format("2006-01-02T15:04")
	data := map[string]any{
		"Title":    post.Title,
		"Active":   "posts",
		"IsNew":    false,
		"Post":     post,
		"Date":     dateStr,
		"Draft":    post.Draft,
		"Tags":     strings.Join(post.Tags, ", "),
		"Body":     post.Body,
		"Slug":     post.Slug,
		"Original": post.Slug,
	}
	h.render(w, "post_edit.html", data)
}

func (h *Handler) SavePost(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title    string `json:"title"`
		Slug     string `json:"slug"`
		Date     string `json:"date"`
		Draft    bool   `json:"draft"`
		Tags     string `json:"tags"`
		Body     string `json:"body"`
		Original string `json:"original"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	date, err := time.Parse("2006-01-02T15:04", req.Date)
	if err != nil {
		date, err = time.Parse(time.RFC3339, req.Date)
		if err != nil {
			date = time.Now().UTC()
		}
	}

	tags := parseTags(req.Tags)
	input := posts.SaveInput{
		Slug:     req.Slug,
		Title:    req.Title,
		Date:     date.UTC(),
		Draft:    req.Draft,
		Tags:     tags,
		Body:     req.Body,
		Original: req.Original,
	}

	if err := h.posts.Save(r.Context(), input); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":   true,
		"slug": posts.SanitizeSlug(req.Slug),
	})
}

func (h *Handler) UploadMedia(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid upload"})
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "file required"})
		return
	}
	defer file.Close()

	if !media.IsImage(header.Filename) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "only image files are allowed"})
		return
	}

	data := make([]byte, header.Size)
	if _, err := file.Read(data); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "read file"})
		return
	}

	item, err := h.media.Upload(r.Context(), header.Filename, data)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":       true,
		"url":      item.URL,
		"markdown": "!(" + item.URL + ")",
		"name":     item.Name,
	})
}

func (h *Handler) ListMedia(w http.ResponseWriter, r *http.Request) {
	items, err := h.media.List(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *Handler) MediaLibrary(w http.ResponseWriter, r *http.Request) {
	items, err := h.media.List(r.Context())
	if err != nil {
		h.renderError(w, "Failed to load media", err)
		return
	}

	data := map[string]any{
		"Title":  "Media Library",
		"Active": "media",
		"Items":  items,
	}
	h.render(w, "media.html", data)
}

func (h *Handler) render(w http.ResponseWriter, name string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.renderer.Render(w, name, data); err != nil {
		http.Error(w, "template error", http.StatusInternalServerError)
	}
}

func (h *Handler) renderError(w http.ResponseWriter, msg string, err error) {
	data := map[string]any{
		"Title":   "Error",
		"Active":  "",
		"Message": msg,
		"Error":   err.Error(),
	}
	h.render(w, "error.html", data)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func parseTags(s string) []string {
	parts := strings.Split(s, ",")
	var tags []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			tags = append(tags, p)
		}
	}
	return tags
}

// NotFound handles unknown routes.
func NotFound(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/api/") {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}
	http.NotFound(w, r)
}

// MethodNotAllowed rejects unsupported HTTP methods.
func MethodNotAllowed(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/api/") {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}

// BadRequest is a helper for handler validation errors.
var ErrBadRequest = errors.New("bad request")
