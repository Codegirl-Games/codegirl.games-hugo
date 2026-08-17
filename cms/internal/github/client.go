package github

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const apiBase = "https://api.github.com"

// ContentItem represents a file or directory entry from the GitHub Contents API.
type ContentItem struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	SHA         string `json:"sha"`
	Size        int    `json:"size"`
	Type        string `json:"type"`
	Content     string `json:"content"`
	Encoding    string `json:"encoding"`
	DownloadURL string `json:"download_url"`
}

// Client performs repository operations via the GitHub REST API.
type Client interface {
	GetFile(ctx context.Context, path string) (*ContentItem, error)
	ListDirectory(ctx context.Context, path string) ([]ContentItem, error)
	CreateOrUpdateFile(ctx context.Context, path, message string, content []byte, sha string) error
	DeleteFile(ctx context.Context, path, message, sha string) error
}

type client struct {
	owner  string
	repo   string
	branch string
	token  string
	http   *http.Client
}

// NewClient creates a GitHub API client.
func NewClient(owner, repo, branch, token string) Client {
	return &client{
		owner:  owner,
		repo:   repo,
		branch: branch,
		token:  token,
		http: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *client) GetFile(ctx context.Context, path string) (*ContentItem, error) {
	endpoint := fmt.Sprintf("%s/repos/%s/%s/contents/%s?ref=%s",
		apiBase, c.owner, c.repo, escapePath(path), url.QueryEscape(c.branch))

	var item ContentItem
	if err := c.doJSON(ctx, http.MethodGet, endpoint, nil, &item); err != nil {
		return nil, err
	}
	if item.Type != "file" {
		return nil, fmt.Errorf("path %q is not a file", path)
	}
	return &item, nil
}

func (c *client) ListDirectory(ctx context.Context, path string) ([]ContentItem, error) {
	endpoint := fmt.Sprintf("%s/repos/%s/%s/contents/%s?ref=%s",
		apiBase, c.owner, c.repo, escapePath(path), url.QueryEscape(c.branch))

	var items []ContentItem
	if err := c.doJSON(ctx, http.MethodGet, endpoint, nil, &items); err != nil {
		return nil, err
	}
	return items, nil
}

func (c *client) CreateOrUpdateFile(ctx context.Context, path, message string, content []byte, sha string) error {
	endpoint := fmt.Sprintf("%s/repos/%s/%s/contents/%s",
		apiBase, c.owner, c.repo, escapePath(path))

	body := map[string]any{
		"message": message,
		"content": base64.StdEncoding.EncodeToString(content),
		"branch":  c.branch,
	}
	if sha != "" {
		body["sha"] = sha
	}

	return c.doJSON(ctx, http.MethodPut, endpoint, body, nil)
}

func (c *client) DeleteFile(ctx context.Context, path, message, sha string) error {
	endpoint := fmt.Sprintf("%s/repos/%s/%s/contents/%s",
		apiBase, c.owner, c.repo, escapePath(path))

	body := map[string]any{
		"message": message,
		"sha":     sha,
		"branch":  c.branch,
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	req.Body = io.NopCloser(bytes.NewReader(data))
	req.ContentLength = int64(len(data))
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("github request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return parseAPIError(resp.StatusCode, respBody)
	}

	return nil
}

func (c *client) doJSON(ctx context.Context, method, endpoint string, body any, out any) error {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint, reqBody)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("github request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return parseAPIError(resp.StatusCode, respBody)
	}

	if out != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, out); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}

	return nil
}

func parseAPIError(status int, body []byte) error {
	var apiErr struct {
		Message string `json:"message"`
	}
	_ = json.Unmarshal(body, &apiErr)

	msg := strings.TrimSpace(apiErr.Message)
	if msg == "" {
		msg = string(body)
	}
	return fmt.Errorf("github api error (%d): %s", status, msg)
}

func escapePath(path string) string {
	parts := strings.Split(path, "/")
	for i, p := range parts {
		parts[i] = url.PathEscape(p)
	}
	return strings.Join(parts, "/")
}

// DecodeContent returns the decoded file content from a GitHub ContentItem.
func DecodeContent(item *ContentItem) ([]byte, error) {
	if item.Encoding != "base64" {
		return nil, fmt.Errorf("unsupported encoding: %s", item.Encoding)
	}
	content := strings.ReplaceAll(item.Content, "\n", "")
	return base64.StdEncoding.DecodeString(content)
}
