package posts

import (
	"context"
	"fmt"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/codegirl-007/hugo-cms/internal/github"
	"gopkg.in/yaml.v3"
)

const postsDir = "content/posts"

// Post represents a Hugo blog post with front matter and body.
type Post struct {
	Slug         string
	Title        string
	Date         time.Time
	Draft        bool
	Tags         []string
	Body         string
	SHA          string
	Extra        map[string]any
	LastModified time.Time
}

// ListItem is a summary of a post for dashboard display.
type ListItem struct {
	Slug         string
	Title        string
	Date         time.Time
	Draft        bool
	LastModified time.Time
}

// SaveInput contains fields submitted when saving a post.
type SaveInput struct {
	Slug     string
	Title    string
	Date     time.Time
	Draft    bool
	Tags     []string
	Body     string
	Original string
}

// Service manages Hugo posts via GitHub.
type Service struct {
	github github.Client
}

// NewService creates a post service.
func NewService(client github.Client) *Service {
	return &Service{github: client}
}

// List returns all posts sorted by date descending.
func (s *Service) List(ctx context.Context) ([]ListItem, error) {
	items, err := s.github.ListDirectory(ctx, postsDir)
	if err != nil {
		return nil, fmt.Errorf("list posts: %w", err)
	}

	var posts []ListItem
	for _, item := range items {
		if item.Type != "file" || !strings.HasSuffix(item.Name, ".md") {
			continue
		}
		if item.Name == "_index.md" {
			continue
		}

		file, err := s.github.GetFile(ctx, item.Path)
		if err != nil {
			continue
		}

		content, err := github.DecodeContent(file)
		if err != nil {
			continue
		}

		post, err := parsePost(item.Name, content)
		if err != nil {
			continue
		}

		posts = append(posts, ListItem{
			Slug:         post.Slug,
			Title:        post.Title,
			Date:         post.Date,
			Draft:        post.Draft,
			LastModified: post.LastModified,
		})
	}

	sort.Slice(posts, func(i, j int) bool {
		return posts[i].Date.After(posts[j].Date)
	})

	return posts, nil
}

// Get loads a single post by slug.
func (s *Service) Get(ctx context.Context, slug string) (*Post, error) {
	slug = sanitizeSlug(slug)
	if slug == "" {
		return nil, fmt.Errorf("invalid slug")
	}

	filePath := postPath(slug)
	file, err := s.github.GetFile(ctx, filePath)
	if err != nil {
		return nil, fmt.Errorf("get post: %w", err)
	}

	content, err := github.DecodeContent(file)
	if err != nil {
		return nil, fmt.Errorf("decode post: %w", err)
	}

	post, err := parsePost(path.Base(file.Path), content)
	if err != nil {
		return nil, err
	}
	post.SHA = file.SHA
	return post, nil
}

// Save creates or updates a post on the main branch.
func (s *Service) Save(ctx context.Context, input SaveInput) error {
	slug := sanitizeSlug(input.Slug)
	if slug == "" {
		return fmt.Errorf("slug is required")
	}
	if strings.TrimSpace(input.Title) == "" {
		return fmt.Errorf("title is required")
	}

	filePath := postPath(slug)
	existingSHA := ""

	// Check if we're renaming from an original slug.
	if input.Original != "" && input.Original != slug {
		originalPath := postPath(sanitizeSlug(input.Original))
		if originalFile, err := s.github.GetFile(ctx, originalPath); err == nil {
			// Delete old file by committing empty content is not ideal;
			// instead update old path only if slug changed - create new, delete old.
			if err := s.github.DeleteFile(ctx, originalPath,
				fmt.Sprintf("Remove post after rename: %s", input.Title), originalFile.SHA); err != nil {
				return fmt.Errorf("remove renamed post: %w", err)
			}
		}
	}

	if existing, err := s.github.GetFile(ctx, filePath); err == nil {
		existingSHA = existing.SHA
	}

	var extra map[string]any
	if input.Original != "" && input.Original == slug {
		if existing, err := s.Get(ctx, slug); err == nil {
			extra = existing.Extra
		}
	}

	content := []byte(renderPost(input, extra))
	action := "Create"
	if existingSHA != "" {
		action = "Update"
	}

	message := fmt.Sprintf("%s post: %s", action, input.Title)
	return s.github.CreateOrUpdateFile(ctx, filePath, message, content, existingSHA)
}

func postPath(slug string) string {
	return path.Join(postsDir, slug+".md")
}

// SanitizeSlug normalizes a post slug.
func SanitizeSlug(slug string) string {
	return sanitizeSlug(slug)
}

func sanitizeSlug(slug string) string {
	slug = strings.TrimSpace(strings.ToLower(slug))
	slug = strings.ReplaceAll(slug, " ", "-")
	var b strings.Builder
	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			b.WriteRune(r)
		} else if r == ' ' || r == '_' {
			b.WriteRune('-')
		}
	}
	result := strings.Trim(b.String(), "-")
	for strings.Contains(result, "--") {
		result = strings.ReplaceAll(result, "--", "-")
	}
	return result
}

func parsePost(filename string, content []byte) (*Post, error) {
	slug := strings.TrimSuffix(filename, ".md")
	fm, body, err := splitFrontMatter(string(content))
	if err != nil {
		return &Post{Slug: slug, Title: slug, Body: string(content)}, nil
	}

	meta := map[string]any{}
	if err := yaml.Unmarshal([]byte(fm), &meta); err != nil {
		return nil, fmt.Errorf("parse front matter: %w", err)
	}

	post := &Post{
		Slug:  slug,
		Body:  strings.TrimLeft(body, "\n"),
		Extra: make(map[string]any),
	}

	if v, ok := meta["title"].(string); ok {
		post.Title = v
	}
	if v, ok := meta["draft"].(bool); ok {
		post.Draft = v
	}
	if tags, ok := meta["tags"].([]any); ok {
		for _, t := range tags {
			if s, ok := t.(string); ok {
				post.Tags = append(post.Tags, s)
			}
		}
	}

	post.Date = parseDate(meta["date"])
	post.LastModified = post.Date

	for k, v := range meta {
		switch k {
		case "title", "date", "draft", "tags":
			continue
		default:
			post.Extra[k] = v
		}
	}

	if post.Title == "" {
		post.Title = slug
	}

	return post, nil
}

func parseDate(v any) time.Time {
	switch d := v.(type) {
	case string:
		formats := []string{
			time.RFC3339,
			"2006-01-02T15:04:05Z07:00",
			"2006-01-02",
		}
		for _, f := range formats {
			if t, err := time.Parse(f, d); err == nil {
				return t
			}
		}
	case time.Time:
		return d
	}
	return time.Now().UTC()
}

func splitFrontMatter(content string) (string, string, error) {
	content = strings.TrimPrefix(content, "\ufeff")
	if !strings.HasPrefix(content, "---") {
		return "", content, fmt.Errorf("no front matter")
	}

	rest := content[3:]
	if strings.HasPrefix(rest, "\n") {
		rest = rest[1:]
	} else if strings.HasPrefix(rest, "\r\n") {
		rest = rest[2:]
	}

	idx := strings.Index(rest, "\n---")
	if idx < 0 {
		return "", content, fmt.Errorf("unclosed front matter")
	}

	fm := rest[:idx]
	body := rest[idx+4:]
	if strings.HasPrefix(body, "\n") {
		body = body[1:]
	} else if strings.HasPrefix(body, "\r\n") {
		body = body[2:]
	}

	return fm, body, nil
}

func renderPost(input SaveInput, extra map[string]any) string {
	meta := map[string]any{
		"title": input.Title,
		"date":  input.Date.UTC().Format(time.RFC3339),
		"draft": input.Draft,
	}
	if len(input.Tags) > 0 {
		meta["tags"] = input.Tags
	}
	for k, v := range extra {
		if _, exists := meta[k]; !exists {
			meta[k] = v
		}
	}

	fmBytes, _ := yaml.Marshal(meta)
	var b strings.Builder
	b.WriteString("---\n")
	b.Write(fmBytes)
	if !strings.HasSuffix(b.String(), "\n") {
		b.WriteString("\n")
	}
	b.WriteString("---\n\n")
	b.WriteString(strings.TrimRight(input.Body, "\n"))
	b.WriteString("\n")
	return b.String()
}

// Stats summarizes post counts for the dashboard.
type Stats struct {
	Total     int
	Drafts    int
	Published int
}

// ComputeStats returns counts from a post list.
func ComputeStats(posts []ListItem) Stats {
	var s Stats
	s.Total = len(posts)
	for _, p := range posts {
		if p.Draft {
			s.Drafts++
		} else {
			s.Published++
		}
	}
	return s
}

// FilterByTitle returns posts whose title contains the query (case-insensitive).
func FilterByTitle(posts []ListItem, query string) []ListItem {
	query = strings.TrimSpace(strings.ToLower(query))
	if query == "" {
		return posts
	}
	var filtered []ListItem
	for _, p := range posts {
		if strings.Contains(strings.ToLower(p.Title), query) {
			filtered = append(filtered, p)
		}
	}
	return filtered
}
