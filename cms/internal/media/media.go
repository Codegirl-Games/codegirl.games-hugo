package media

import (
	"context"
	"fmt"
	"path"
	"sort"
	"strings"

	"github.com/codegirl-007/hugo-cms/internal/github"
)

const uploadsDir = "static/uploads"

// Item represents an uploaded media file.
type Item struct {
	Name string `json:"name"`
	Path string `json:"path"`
	URL  string `json:"url"`
	Size int    `json:"size"`
}

// Service manages media uploads via GitHub.
type Service struct {
	github github.Client
}

// NewService creates a media service.
func NewService(client github.Client) *Service {
	return &Service{github: client}
}

// List returns all files in the uploads directory.
func (s *Service) List(ctx context.Context) ([]Item, error) {
	items, err := s.github.ListDirectory(ctx, uploadsDir)
	if err != nil {
		// Directory may not exist yet.
		if strings.Contains(err.Error(), "404") {
			return []Item{}, nil
		}
		return nil, fmt.Errorf("list media: %w", err)
	}

	var media []Item
	for _, item := range items {
		if item.Type != "file" {
			continue
		}
		media = append(media, Item{
			Name: item.Name,
			Path: item.Path,
			URL:  "/uploads/" + item.Name,
			Size: item.Size,
		})
	}

	sort.Slice(media, func(i, j int) bool {
		return media[i].Name < media[j].Name
	})

	return media, nil
}

// Upload saves a binary file to static/uploads/ via the GitHub API.
func (s *Service) Upload(ctx context.Context, filename string, data []byte) (*Item, error) {
	filename = sanitizeFilename(filename)
	if filename == "" {
		return nil, fmt.Errorf("invalid filename")
	}

	filePath := path.Join(uploadsDir, filename)
	existingSHA := ""

	if existing, err := s.github.GetFile(ctx, filePath); err == nil {
		existingSHA = existing.SHA
	}

	message := fmt.Sprintf("Upload image: %s", filename)

	if err := s.github.CreateOrUpdateFile(ctx, filePath, message, data, existingSHA); err != nil {
		return nil, fmt.Errorf("upload media: %w", err)
	}

	return &Item{
		Name: filename,
		Path: filePath,
		URL:  "/uploads/" + filename,
		Size: len(data),
	}, nil
}

func sanitizeFilename(name string) string {
	name = path.Base(strings.TrimSpace(name))
	name = strings.ToLower(name)
	var b strings.Builder
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' || r == '.' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// IsImage returns true if the filename has a common image extension.
func IsImage(filename string) bool {
	lower := strings.ToLower(filename)
	for _, ext := range []string{".png", ".jpg", ".jpeg", ".gif", ".webp", ".svg"} {
		if strings.HasSuffix(lower, ext) {
			return true
		}
	}
	return false
}
