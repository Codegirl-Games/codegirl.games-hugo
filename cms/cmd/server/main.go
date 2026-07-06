package main

import (
	"log"
	"net/http"

	"github.com/codegirl-007/hugo-cms/internal/auth"
	"github.com/codegirl-007/hugo-cms/internal/config"
	"github.com/codegirl-007/hugo-cms/internal/github"
	"github.com/codegirl-007/hugo-cms/internal/handlers"
	"github.com/codegirl-007/hugo-cms/internal/media"
	"github.com/codegirl-007/hugo-cms/internal/posts"
	"github.com/codegirl-007/hugo-cms/internal/session"
	"github.com/codegirl-007/hugo-cms/internal/templates"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	gh := github.NewClient(cfg.GitHubOwner, cfg.GitHubRepo, cfg.GitHubBranch, cfg.GitHubToken)
	authenticator := auth.New(cfg.AdminUsername, cfg.AdminPasswordHash)
	sessions := session.NewStore(cfg.SessionSecret, cfg.CookieSecure)
	postService := posts.NewService(gh)
	mediaService := media.NewService(gh)

	renderer, err := templates.New()
	if err != nil {
		log.Fatalf("templates: %v", err)
	}

	h := handlers.New(authenticator, sessions, postService, mediaService, renderer)

	mux := http.NewServeMux()

	// Public routes
	mux.HandleFunc("GET /login", h.LoginPage)
	mux.HandleFunc("POST /login", h.Login)
	mux.HandleFunc("POST /logout", h.Logout)
	mux.Handle("GET /assets/", http.StripPrefix("/assets/", templates.StaticHandler()))

	// Protected admin routes
	mux.HandleFunc("GET /admin", h.Auth(h.Dashboard))
	mux.HandleFunc("GET /admin/posts", h.Auth(h.PostsList))
	mux.HandleFunc("GET /admin/posts/new", h.Auth(h.PostNew))
	mux.HandleFunc("GET /admin/posts/{slug}", h.Auth(h.PostEdit))
	mux.HandleFunc("GET /admin/media", h.Auth(h.MediaLibrary))

	// Protected API routes
	mux.HandleFunc("POST /api/posts/save", h.Auth(h.SavePost))
	mux.HandleFunc("POST /api/media/upload", h.Auth(h.UploadMedia))
	mux.HandleFunc("GET /api/media", h.Auth(h.ListMedia))

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
	})

	server := &http.Server{
		Addr:    cfg.Addr,
		Handler: mux,
	}

	log.Printf("CMS server listening on %s", cfg.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server: %v", err)
	}
}
