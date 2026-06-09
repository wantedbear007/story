# Story

A CLI-first second brain for developers. Capture learning, work logs, resources, and engineering notes — transform them into structured knowledge and publish to your favorite platforms.

## Features

- **Knowledge capture** — CLI commands to quickly log learning, work, resources, and engineering notes
- **Full-text search** — Search across all entries, tags, and resources
- **Resource management** — Track URLs, GitHub repos, articles, YouTube videos, PDFs, and markdown
- **Tag & collection organization** — Organize entries with tags and collections
- **AI content generation** — Generate tweets from entries using LLMs (OpenAI, Gemini, Anthropic, Ollama)
- **Tweet lifecycle management** — Draft -> Review -> Approve -> Schedule -> Post, with audit trail
- **Prompt versioning** — Version-controlled prompt templates for content generation
- **Cost tracking** — Track token usage and cost per generation
- **Web dashboard** — Browser-based UI for managing tweets (`story web`)
- **Publishing** — Publish entries to Twitter, blog, or markdown
- **Auth** — Email/password registration, session management, email verification, password reset

## Quick Start

### Prerequisites

- Go 1.21+
- PostgreSQL 14+
- An LLM provider API key (OpenAI, Gemini, Anthropic, or Ollama)

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

# Run database migrations
goose -dir migrations postgres "$(go run ./cmd/story config dsn)" up

# Build and run
go build ./cmd/story
./story --help
```

### Configuration

Config is loaded from `configs/config.yaml` and overridden by `STORY_*` environment variables:

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

### First Run

```bash
# Register an account
./story auth register --email you@example.com --password "secure-pass" --display-name "Your Name"

# Login
./story auth login --email you@example.com --password "secure-pass"

# Add your first entry
./story entry add --type learning --title "Go Interfaces" --tags go,patterns
# (type/paste content, then Ctrl+D)

# Generate a tweet from the entry
./story tweet generate <entry-id>

# Start the web dashboard
./story web
```

## CLI Reference

### Auth
| Command | Description |
|---------|-------------|
| `story auth register` | Create a new account |
| `story auth login` | Login and save session |
| `story auth logout` | Logout and revoke session |
| `story auth status` | Show current login status |
| `story auth sessions` | List active sessions |
| `story auth revoke <id>` | Revoke a session |
| `story auth verify <token>` | Verify email address |
| `story auth password change` | Change password |
| `story auth password forgot` | Request password reset |
| `story auth password reset <token>` | Reset password |

### Entries
| Command | Description |
|---------|-------------|
| `story entry add` | Add a new entry (content from stdin) |
| `story entry edit <id>` | Edit an entry |
| `story entry delete <id>` | Soft-delete an entry |
| `story timeline` | Show recent entries |
| `story search <query>` | Full-text search entries |

### Resources
| Command | Description |
|---------|-------------|
| `story resource add` | Add a URL, GitHub repo, article, video, etc. |
| `story resource list` | List resources |
| `story resource search <query>` | Search resources |
| `story resource attach <rid> <eid>` | Attach resource to entry |

### Tweets (Content Generation)
| Command | Description |
|---------|-------------|
| `story tweet generate <eid>` | Generate tweet draft from entry |
| `story tweet regenerate <tid>` | Regenerate tweet (new version) |
| `story tweet list` | List tweets with filters |
| `story tweet get <tid>` | Show tweet details |
| `story tweet approve <tid>` | Approve tweet |
| `story tweet review <tid>` | Send to review |
| `story tweet reject <tid>` | Reject back to draft |
| `story tweet schedule <tid> <dt>` | Schedule for posting |
| `story tweet archive <tid>` | Archive tweet |
| `story tweet audit <tid>` | Show audit trail |

### Publishing
| Command | Description |
|---------|-------------|
| `story target add` | Add a publishing target |
| `story target list` | List targets |
| `story publish entry <eid> <tid>` | Publish entry to target |

### Web Dashboard
| Command | Description |
|---------|-------------|
| `story web` | Start the web dashboard and open browser |

## Web Dashboard

The web dashboard provides a browser-based interface for managing tweets:

```
story web --port 8080
```

- **Tweet drafts page** — View and filter all tweets by status
- **Tweet editor** — Edit content, see character count, regenerate, approve, archive
- **Resource viewer** — View entry details and attached resources
- **Copy button** — One-click copy tweet content to clipboard

The dashboard uses a REST API (`/api/*`) with JWT authentication and an embeddable static frontend for future React migration.

## Architecture

```
cmd/story/main.go              — Entrypoint, wires dependencies
internal/
  domain/                      — Entities, repository interfaces
  application/                 — Business logic services
    auth/                      — Authentication, sessions
    user/                      — Registration, profile
    entry/                     — Entry CRUD, search
    collection/                — Collection management
    tag/                       — Tag management
    resource/                  — Resource tracking
    publishing/                — Publishing pipeline
    content/                   — Tweet generation, lifecycle
  infrastructure/
    config/                    — YAML config + env overrides
    llm/                       — LLM provider (OpenAI/Gemini/Anthropic/Ollama)
    auth/                      — JWT service, password hashing
    email/                     — SMTP mailer
    repository/                — PostgreSQL persistence
    bootstrap/                 — App initialization
  interfaces/
    cli/                       — Cobra CLI commands
    api/                       — REST API + handlers
web/                           — Frontend (HTML/CSS/JS SPA)
migrations/                    — Goose SQL migrations
```

### Layer Isolation

- **Domain** — Zero external dependencies (stdlib + uuid only)
- **Application** — Depends only on domain interfaces
- **Infrastructure** — Implements domain interfaces
- **Interfaces** — CLI and API adapt to application services

### Content Generation Flow

1. User runs `story tweet generate <entry-id>` or clicks "Generate" in web UI
2. Service fetches the entry and the latest prompt template version
3. Prompt is rendered using Go `text/template` with entry data
4. LLM provider is called with retry logic (exponential backoff + jitter)
5. Cost is estimated based on token counts and model pricing table
6. Tweet is created in `draft` status with full generation metadata
7. Every action is recorded in the audit trail
8. Tweet lifecycle: draft -> reviewing -> approved -> scheduled -> posted

### Database

PostgreSQL with migrations managed by goose. Key tables:
- `users`, `sessions` — Auth
- `entries`, `entry_tags`, `entry_collections` — Knowledge base
- `tags`, `collections` — Organization
- `resources`, `entry_resources` — External resources
- `tweets`, `prompt_templates`, `generation_audits` — Content generation
- `publishing_targets`, `published_entries` — Publishing

## Development

```bash
# Build
go build ./cmd/story

# Test
go test ./...

# Run linter
golangci-lint run

# Add migration
goose -dir migrations postgres "$DSN" create add_some_feature sql

# Run migration
goose -dir migrations postgres "$DSN" up
```

### LLM Provider Configuration

Default providers and models:

| Provider | Config Provider | Default Model |
|----------|----------------|---------------|
| OpenAI | `openai` | gpt-4 |
| Google Gemini | `gemini` | gemini-pro |
| Anthropic | `anthropic` | claude-3-opus |
| Ollama | `ollama` | llama2 |

Switch providers by changing `llm.provider` in config or setting `STORY_LLM_*` env vars.

## License

MIT
