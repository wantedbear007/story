# Story

A CLI-first second brain for developers. Capture raw thoughts, learning, work logs, resources, and engineering notes — transform them into structured knowledge and publish to your favorite platforms.

## Features

- **Raw Capture** — Frictionless unstructured capture: type interactively, pipe (`git diff | story raw`), or read from a file. Content stored as-is, processed into structured knowledge later
- **Knowledge capture** — Log learning, work logs, resources, and engineering notes with tags and collections
- **Full-text search** — Search across all entries, tags, and resources (PostgreSQL `tsvector`)
- **Resource management** — Track URLs, GitHub repos, articles, YouTube videos, PDFs, and markdown with typed resources
- **Tag & collection organization** — Hierarchical collections, many-to-many tags and collections per entry
- **AI content generation** — Generate tweets, threads, and blog posts from entries using OpenAI, Gemini, Anthropic, or Ollama
- **Tweet lifecycle** — Draft → Review → Approve → Schedule → Post with full audit trail and version history
- **Prompt versioning** — Version-controlled prompt templates with semantic placeholders
- **Cost tracking** — Track input/output tokens and USD cost per generation
- **Web dashboard** — SPA for managing tweets, browsing entries, and viewing resources (`story web`)
- **Auto-login via URL** — Server generates a short code; browser opens with `?code=` and exchanges it for a JWT automatically
- **Publishing** — Publish entries to Twitter, blog, or markdown targets
- **Authentication** — Email/password registration, session management, email verification, password reset

## Quick Start

### Prerequisites

- Go 1.21+
- PostgreSQL 14+
- An LLM provider API key (optional, for tweet generation)

### Installation

```bash
# Clone the repository
git clone <repo-url> && cd story

# Copy and edit configuration
cp configs/config.example.yaml configs/config.yaml

# Set required environment variables
export STORY_DATABASE_PASSWORD=your_db_password
export STORY_AUTH_JWT_SECRET=your-32-char-jwt-secret-min
export STORY_LLM_OPENAI_API_KEY=sk-...  # or Gemini/Anthropic/Ollama

# Run database setup (migrations + schema)
go run ./cmd/story setup

# Build
go build ./cmd/story

# Register an account (you'll be prompted for email, password, and display name)
./story register

# Login (you'll be prompted for email and password)
./story login

# Quick raw capture (type notes, press Ctrl+D when done)
./story raw

# Pipe content directly
git diff | story raw
cat notes.txt | story raw

# Add your first structured entry
./story capture

# Generate a tweet from the entry
./story tweet generate <entry-id>

# Start the web dashboard
./story web
```

### Configuration

Config is loaded from `configs/config.yaml`, `~/.story/config.yaml`, or `$STORY_CONFIG_PATH`. Overridden by `STORY_*` environment variables:

| Variable | Description |
|----------|-------------|
| `STORY_DATABASE_PASSWORD` | PostgreSQL password |
| `STORY_AUTH_JWT_SECRET` | JWT signing secret (min 32 chars) |
| `STORY_LLM_OPENAI_API_KEY` | OpenAI API key |
| `STORY_LLM_GEMINI_API_KEY` | Google Gemini API key |
| `STORY_LLM_ANTHROPIC_API_KEY` | Anthropic API key |
| `STORY_LLM_OLLAMA_BASE_URL` | Ollama server URL |
| `STORY_SMTP_PASSWORD` | SMTP password |
| `STORY_APP_ENVIRONMENT` | development, staging, or production |
| `STORY_DATABASE_HOST` | Database host |
| `STORY_DATABASE_PORT` | Database port |
| `STORY_DATABASE_NAME` | Database name |
| `STORY_DATABASE_USER` | Database user |
| `STORY_SERVER_HOST` | Web dashboard host |
| `STORY_SERVER_PORT` | Web dashboard port |

## CLI Reference

### Onboarding
| Command | Description |
|---------|-------------|
| `story init` | Create initial `~/.story/config.yaml` |
| `story verify` | Check configuration and database connectivity |
| `story setup` | Run database migrations and create schema |

### Auth
| Command | Description |
|---------|-------------|
| `story register` | Create a new account (top-level) |
| `story login` | Login and save session (top-level) |
| `story auth logout` | Logout and revoke current session |
| `story auth status` | Show current login status |
| `story auth sessions` | List active sessions |
| `story auth revoke <id>` | Revoke a specific session |
| `story password change` | Change password (top-level) |
| `story forgot-password` | Request password reset (top-level) |
| `story auth verify <token>` | Verify email address |

### Raw Capture
| Command | Description |
|---------|-------------|
| `story raw` | Interactive mode — type notes, press Ctrl+D to finish |
| `story raw --file <path>` | Read content from a file |
| `echo "note" \| story raw` | Pipe input from another command |
| `story process raw <id>` | Process a raw entry into structured knowledge (pending AI) |
| `story process raw --all` | Process all unprocessed raw entries (pending AI) |

### Entries
| Command | Description |
|---------|-------------|
| `story capture` | Interactive entry capture (type, title, tags) |
| `story entry add` | Add entry with full options |
| `story entry edit <id>` | Edit an existing entry |
| `story entry delete <id>` | Soft-delete an entry |
| `story timeline` | Show recent entries |
| `story query` | Search and list entries |
| `story search <query>` | Full-text search across entries |

### Organization
| Command | Description |
|---------|-------------|
| `story collection create` | Create a collection |
| `story collection list` | List collections |
| `story collection add <eid> <cid>` | Add entry to collection |
| `story collection remove <eid> <cid>` | Remove entry from collection |
| `story tag create` | Create a tag |
| `story tag list` | List tags |
| `story resource add` | Add a URL, GitHub repo, article, video, etc. |
| `story resource list` | List resources |
| `story resource search <query>` | Search resources |
| `story resource attach <rid> <eid>` | Attach resource to entry |

### Tweets (Content Generation)
| Command | Description |
|---------|-------------|
| `story tweet generate <eid>` | Generate tweet draft from entry |
| `story tweet regenerate <tid>` | Regenerate tweet (new version) |
| `story tweet list` | List tweets with status/entry filters |
| `story tweet get <tid>` | Show tweet details and metadata |
| `story tweet approve <tid>` | Approve tweet |
| `story tweet review <tid>` | Send to review |
| `story tweet reject <tid>` | Reject back to draft |
| `story tweet schedule <tid> <dt>` | Schedule for posting |
| `story tweet archive <tid>` | Archive tweet |
| `story tweet audit <tid>` | Show audit trail |

### Publishing
| Command | Description |
|---------|-------------|
| `story target add` | Add a publishing target (twitter, blog, markdown) |
| `story target list` | List publishing targets |
| `story publish entry <eid> <tid>` | Publish entry to target |

### System
| Command | Description |
|---------|-------------|
| `story web` | Start the web dashboard and open browser |
| `story config show` | View configuration |
| `story config validate` | Validate configuration |
| `story config smtp` | Configure SMTP settings |
| `story whoami` | Show current logged-in user |

## Web Dashboard

The dashboard is an embedded SPA for managing tweets, browsing entries, and viewing resources:

```
story web --port 8080
```

### Login Flow

1. Server generates a 6-character alphanumeric code and maps it to the session JWT
2. Terminal prints: `Open this URL: http://localhost:8080/?code=Qp9D73`
3. Browser opens with the code in the URL
4. The SPA exchanges the code for a JWT via `GET /api/exchange/{code}`
5. Code is single-use and expires after 5 minutes
6. JWT is stored in `localStorage` for subsequent page loads
7. The `?code=` parameter is removed from the URL after exchange

### Dashboard Features

- **Tweet drafts page** — View and filter tweets by status (draft, reviewing, approved, scheduled, posted, archived)
- **Tweet editor** — Edit content, see character count, regenerate, approve, schedule, archive
- **Entry viewer** — Browse entry details and attached resources
- **Copy button** — One-click copy tweet content to clipboard
- **Auto-disconnect** — Page pings `/api/ping` every 3 seconds; shows "Server Disconnected" when the server shuts down

### API Endpoints

All routes prefixed with `/api`. Authenticated endpoints require `Authorization: Bearer <jwt>`:

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/api/exchange/{code}` | No | Exchange login code for JWT |
| GET | `/api/ping` | No | Health check |
| GET | `/api/me` | JWT | Current user info |
| GET | `/api/tweets` | JWT | List tweets (filters: entry_id, status, limit, offset) |
| GET | `/api/tweets/{id}` | JWT | Get tweet details |
| POST | `/api/tweets/generate` | JWT | Generate tweet from entry |
| POST | `/api/tweets/{id}/regenerate` | JWT | Regenerate tweet (new version) |
| PUT | `/api/tweets/{id}` | JWT | Update tweet content |
| POST | `/api/tweets/{id}/approve` | JWT | Approve tweet |
| POST | `/api/tweets/{id}/review` | JWT | Send to review |
| POST | `/api/tweets/{id}/reject` | JWT | Reject to draft |
| POST | `/api/tweets/{id}/schedule` | JWT | Schedule for posting |
| POST | `/api/tweets/{id}/archive` | JWT | Archive tweet |
| GET | `/api/tweets/{id}/audits` | JWT | Audit trail |
| GET | `/api/entries` | JWT | List entries (filters: q, page, page_size) |
| GET | `/api/entries/{id}` | JWT | Get entry with tags and resources |
| GET | `/api/prompts` | JWT | List prompt templates |

## Architecture

```
cmd/story/main.go              — Entrypoint, dependency injection
internal/
  domain/                      — Entities, repository interfaces, sentinel errors
    raw_entry.go               — RawEntry model + RawEntryRepository interface
    entry.go                   — Entry model + EntryRepository interface
    tweet.go                   — Tweet, PromptTemplate, GenerationAudit models
    resource.go                — Resource model + ResourceRepository interface
    session.go, user.go        — Auth models
    tag.go, collection.go      — Organization models
    publishing.go              — Publishing models
    errors.go                  — ErrNotFound, ErrInvalidInput, etc.
  application/                 — Business logic services
    auth/                      — Login, register, session management, password reset
    user/                      — Registration, profile, email verification
    entry/                     — Entry CRUD, full-text search, tag/resource association
    collection/                — Hierarchical collection management
    tag/                       — Tag management
    resource/                  — Resource CRUD, search, entry attachment
    raw_entry/                 — Raw note capture (create, list, status transitions)
    content/                   — Tweet generation with LLM, lifecycle, audit trail
    publishing/                — Publish to external targets
  infrastructure/
    config/                    — YAML config + env variable overrides
    database/                  — PostgreSQL connection pool (pgx)
    auth/                      — JWT signing/validation (HS256), Argon2id hashing
    llm/                       — OpenAI, Gemini, Anthropic, Ollama adapters
    email/                     — SMTP mailer
    repository/                — PostgreSQL implementations of all domain repos
    setup/                     — Migration runner (embedded Goose-style SQL)
    bootstrap/                 — App lifecycle (config → logger → DB → signals)
  interfaces/
    cli/                       — 53 leaf commands across 20 Go files, with interactive prompts
    api/                       — REST API (15 authenticated + 2 public endpoints), auth middleware, CORS, login codes
web/                           — Embedded SPA (HTML/CSS/JS)
  index.html                   — App shell with login code input
  css/app.css                  — Dark theme
  js/
    api.js                     — API client (fetch wrapper, token management)
    app.js                     — Router, auth flow, heartbeat, toast notifications
    pages/
      drafts.js                — Tweet list with status filters
      edit.js                  — Tweet editor with lifecycle actions
      resources.js             — Entry/resource viewer
migrations/                    — Goose SQL migrations (14 files)
```

### Layer Isolation

- **Domain** — Zero external dependencies (stdlib + uuid only). Defines entities, value objects, and repository interfaces
- **Application** — Depends only on domain interfaces. Orchestrates business logic, DTO conversion, cross-cutting concerns
- **Infrastructure** — Implements domain interfaces (PostgreSQL via pgx, LLM via HTTP, email via SMTP)
- **Interfaces** — CLI (Cobra) and API (net/http) adapt to application services

### Raw Capture Flow

1. User runs `story raw` (interactive), `story raw -f <file>` (file), or pipes input (`git diff | story raw`)
2. Service creates a `RawEntry` with status `raw` and source `cli`/`file`/`pipe`
3. Content is stored as-is — never modified, always the source of truth
4. Status lifecycle: `raw` → `processing` → `structured` → `archived`
5. `story process raw` (future) will convert raw entries into structured entries, topics, and tags via AI

### Content Generation Flow

1. User runs `story tweet generate <entry-id>` or clicks "Generate" in web UI
2. Service fetches the entry and the latest prompt template version
3. Prompt is rendered using Go `text/template` with entry data (`{{.Title}}`, `{{.Content}}`)
4. LLM provider is called with retry logic (exponential backoff + jitter, 3 retries)
5. Cost is estimated based on token counts and model-specific pricing table
6. Tweet is created in `draft` status with full generation metadata (provider, model, tokens, cost, latency)
7. Every action (generate, approve, reject, schedule) is recorded in `generation_audits`
8. Tweet lifecycle: `draft` → `reviewing` → `approved` → `scheduled` → `posted`

### Web Auth Flow

1. `story web` loads the session JWT from `~/.story/session.json`
2. Server creates a 6-character code → JWT mapping (expires 5 min, single-use)
3. Browser opens with `http://localhost:8080/?code=Qp9D73`
4. SPA extracts the code from URL, calls `GET /api/exchange/{code}`
5. Server validates the JWT before returning it (rejects expired tokens)
6. SPA stores JWT in `localStorage`, cleans the URL via `history.replaceState`
7. All subsequent API calls use `Authorization: Bearer <jwt>`
8. A heartbeat pings `/api/ping` every 3 seconds; page shows "Server Disconnected" on failure

### Database

PostgreSQL managed by Goose-style migrations (14 migrations, embedded in binary):

| Migration | Tables | Purpose |
|-----------|--------|---------|
| 00001 | `users` | Core user accounts with soft delete |
| 00002 | `entries` | Structured entries with type enum, full-text search, soft delete |
| 00003 | `tags` | User-scoped unique tags |
| 00004 | `collections` | Hierarchical collections (parent_id), soft delete |
| 00005 | `entry_collections` | Many-to-many entries ↔ collections |
| 00006 | `entry_tags` | Many-to-many entries ↔ tags |
| 00007 | `publishing_targets` | External platform configs (twitter, notion, blog, markdown) |
| 00008 | `published_entries` | Publication records with status tracking |
| 00009 | `refresh_tokens` | Legacy refresh tokens (replaced by sessions) |
| 00010 | `password_reset_tokens` | Password reset flow |
| 00011 | `sessions`, `email_verifications` | JWT session management, email verification tokens |
| 00012 | `resources`, `entry_resources` | Typed resources (url, github, article, youtube, pdf, markdown) with GIN indexes |
| 00013 | `prompt_templates`, `tweets`, `generation_audits` | AI content generation system with 3 seeded prompts |
| 00014 | `raw_entries` | Unstructured note capture with status/source enums |

## Development

```bash
# Build
go build ./cmd/story

# Test
go test ./...

# Run all domain tests
go test ./internal/domain/...

# Run linter
golangci-lint run

# Add migration (creates in both locations)
cp migrations/00015_*.sql internal/infrastructure/setup/migrations/

# Run migration
go run ./cmd/story setup
```

### LLM Provider Configuration

| Provider | Config Name | Default Model |
|----------|-------------|---------------|
| OpenAI | `openai` | gpt-4 |
| Google Gemini | `gemini` | gemini-pro |
| Anthropic | `anthropic` | claude-3-opus |
| Ollama | `ollama` | llama2 |

Switch by setting `llm.provider` in `config.yaml` or the corresponding `STORY_LLM_*` env var.

### Project Structure

```
├── cmd/story/main.go           — Entrypoint + DI wiring
├── internal/
│   ├── application/            — Business logic (10 service packages)
│   ├── infrastructure/         — Implementations (DB, LLM, auth, email, config)
│   └── interfaces/
│       ├── cli/                — 53 leaf commands
│       └── api/                — 17 REST endpoints + middleware
├── web/                        — Embedded SPA (5 JS files, CSS, HTML)
├── migrations/                 — 14 Goose SQL migrations
└── configs/                    — Config examples
```

## License

MIT
