# Hugo CMS

A lightweight, self-hosted CMS for Hugo websites. Edit posts and upload images from desktop or mobile while keeping **Git as the single source of truth**. No database required.

All repository operations go through the **GitHub REST API** — the server never runs `git` commands.

## Architecture

```
Browser → Go Server → GitHub REST API → GitHub Repository → GitHub Action → Hugo Site
```

## Features

- Single admin authentication (bcrypt password hash + secure session cookie)
- Dashboard with post stats and recent posts
- Create and edit Hugo posts with automatic front matter generation
- EasyMDE markdown editor with split preview, toolbar, drag/drop images, and keyboard shortcuts
- Media library for uploading images to `static/uploads/`
- Client-side autosave (every 30 seconds) and unsaved changes warning
- Dark mode and responsive layout
- Auto-generated slugs from titles

## Project Structure

```
cms/
├── cmd/
│   ├── server/          # Main HTTP server
│   └── hashpassword/    # Utility to generate bcrypt hashes
├── internal/
│   ├── auth/            # Password verification
│   ├── config/          # Environment configuration
│   ├── github/          # GitHub REST API client
│   ├── handlers/        # HTTP route handlers
│   ├── media/           # Image upload service
│   ├── posts/           # Post parsing and saving
│   ├── session/         # Signed session cookies
│   └── templates/       # Embedded HTML templates and static assets
└── web/                 # (assets embedded in internal/templates/static)
```

## Configuration

Copy `.env.example` to `.env` and fill in the values:

| Variable | Description |
|----------|-------------|
| `GITHUB_TOKEN` | GitHub personal access token with `repo` scope |
| `GITHUB_OWNER` | Repository owner (user or org) |
| `GITHUB_REPO` | Repository name |
| `GITHUB_BRANCH` | Branch to commit to (default: `master`) |
| `SESSION_SECRET` | Random string for signing session cookies |
| `ADMIN_USERNAME` | Admin login username |
| `ADMIN_PASSWORD_HASH` | Bcrypt hash of admin password |
| `ADDR` | Listen address (default: `:8080`) |
| `COOKIE_SECURE` | Set `false` for local HTTP dev (default: `true`) |

### Generate a password hash

```bash
cd cms
go run ./cmd/hashpassword 'your-secure-password'
```

### Generate a session secret

```bash
openssl rand -hex 32
```

## Running Locally

```bash
cd cms
go mod tidy
export GITHUB_TOKEN=...
export GITHUB_OWNER=...
export GITHUB_REPO=...
export SESSION_SECRET=...
export ADMIN_USERNAME=admin
export ADMIN_PASSWORD_HASH=...
export COOKIE_SECURE=false
go run ./cmd/server
```

Open http://localhost:8080/admin

## HTTP Routes

| Method | Path | Description |
|--------|------|-------------|
| GET | `/login` | Login page |
| POST | `/login` | Authenticate |
| POST | `/logout` | End session |
| GET | `/admin` | Dashboard |
| GET | `/admin/posts` | Post listing |
| GET | `/admin/posts/new` | New post editor |
| GET | `/admin/posts/:slug` | Edit post |
| GET | `/admin/media` | Media library |
| POST | `/api/posts/save` | Save post to GitHub |
| POST | `/api/media/upload` | Upload image |
| GET | `/api/media` | List uploaded images |

## Deployment

Build a static binary and run behind a reverse proxy (nginx, Caddy, etc.) with HTTPS:

```bash
cd cms
CGO_ENABLED=0 go build -o hugo-cms ./cmd/server
```

Example Caddy config:

```
cms.example.com {
    reverse_proxy localhost:8080
}
```

Set `COOKIE_SECURE=true` in production so session cookies are only sent over HTTPS.

## Post Format

Posts are saved to `content/posts/<slug>.md` with Hugo front matter:

```yaml
---
title: "Hello World"
date: 2026-07-06T18:00:00Z
draft: false
tags:
  - hugo
  - programming
---

My markdown content.
```

Images are stored at `static/uploads/<filename>` and referenced as `![](/uploads/image.png)`.

## GitHub Token Permissions

The token needs write access to the repository contents. A fine-grained token with **Contents: Read and write** on the target repo, or a classic token with the `repo` scope, is sufficient.

## License

MIT
