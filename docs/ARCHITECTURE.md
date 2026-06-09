# Story Architecture

## Overview
CLI-first second brain for developers. Captures learning, work logs, resources, engineering notes, transforms them into structured knowledge, and publishes to platforms.

## Stack
- **Language:** Go 1.21+
- **Database:** PostgreSQL via pgx v5
- **Migrations:** goose
- **CLI:** cobra
- **Auth:** JWT (access) + opaque token (refresh/session)
- **LLM providers:** OpenAI, Gemini, Anthropic, Ollama
- **Config:** YAML + env var overrides (STORY_*)

## Codebase layout

```
cmd/story/main.go           — entrypoint, wires everything
internal/
  domain/                   — entities, repository interfaces, errors, time
  application/              — service layer (business logic + DTOs)
    auth/                    — auth: login, session mgmt, password reset
    user/                   — user: register, verify, profile, change password
    entry/                  — entries CRUD, search
    collection/             — collections/groups
    tag/                    — tags
    resource/               — resources CRUD, attach/detach
    publishing/             — publish to targets (Twitter, blog, etc.)
    content/                — content generation (tweet lifecycle)
  infrastructure/
    config/                 — YAML config loading + env overrides
    bootstrap/              — app initialization (DB pool, etc.)
    llm/                    — LLM provider abstraction + 4 implementations
    auth/                   — JWT service, password hashing
    email/                  — SMTP mailer
    repository/             — all pgx-based repository implementations
  interfaces/cli/           — cobra commands
  pkg/logger/               — structured logger
migrations/                 — SQL migrations (goose)
configs/                    — sample configs
docs/                       — architecture reference
```

## Core patterns

### Layer isolation
- `domain` has NO imports from other project packages. Only stdlib + uuid.
- `application` imports only `domain` + DTOs from sibling packages.
- `infrastructure` implements `domain` repository interfaces.
- `interfaces/cli` depends on `application` services via `Dependencies`.

### Repositories
- Named `*Repository` in `internal/infrastructure/repository/`.
- Implements `domain.*Repository` interface.
- Uses `pgxpool.Pool` for connections.
- Scan helpers use `scannable` interface (defined in `entry_repo.go`).
- Helper: `isUniqueViolation(err)` for duplicate detection.
- Error mapping: `pgx.ErrNoRows` -> `domain.ErrNotFound`.

### Services
- Named `*Service` in `internal/application/*/service.go`.
- Accepts domain repository interfaces, not concrete types.
- Returns DTOs, not domain entities.
- Errors wrap with context: `fmt.Errorf("doing thing: %w", err)`.
- Entry service auto-associates tags (create-or-get pattern).

### Config (yaml)
- Loaded from path, defaults hardcoded in `Load()`.
- Env overrides: `STORY_DATABASE_PASSWORD`, `STORY_AUTH_JWT_SECRET`, etc.
- `LLMConfig` has sub-configs per provider with API keys from env.
- Config file path: `STORY_CONFIG_PATH` env or `configs/config.yaml`.

### Auth
- Registration creates user + sends verification email.
- Login creates session + JWT (access token).
- Sessions tracked in `sessions` table with SHA-256 token hash.
- Refresh token rotation: old session revoked, new session created.
- CLI stores session in `~/.story/session.json`.

### CLI conventions
- Each command group: `newXCommand(deps *Dependencies)`, registered in `NewRootCommand`.
- All commands resolve authenticated user via `resolveCurrentUserID(deps)`.
- Flags for optional args, positional args for required (IDs).
- `uuidParse(s)` for UUID validation.

## Domain Entities
- User: id, email, password_hash, display_name, email_verified_at, timestamps, soft-delete
- Session: id, user_id, token_hash, device_info, ip, is_revoked, expires, last_used, created
- Entry: id, user_id, type (learning|work_log|resource|engineering_note), title, content, metadata (JSONB), timestamps, soft-delete
- Tag: id, user_id, name (unique per user)
- Collection: id, user_id, name, description
- Resource: id, user_id, type (url|github|article|youtube|pdf|markdown), title, url, description, metadata, content_hash
- PublishingTarget: id, user_id, type, name, config (JSONB)
- PublishedEntry: id, entry_id, target_id, external_url, status, error_message
- Tweet: id, entry_id, user_id, content, status (draft|reviewing|approved|scheduled|posted|archived), version, prompt_id, provider_name, model_name, token counts, cost, retry_count, latency_ms, error_message, scheduled_for, posted_at
- PromptTemplate: id, name, version (composite unique), template, description
- GenerationAudit: id, tweet_id, action, user_id, previous/new content/status, metadata (JSONB)

## Completed features

### Auth system (migration 00011)
- Email verification during registration
- Session management (login, logout, list, revoke)
- Token rotation on refresh
- Password change, forgot, reset (revokes all sessions)
- JWT with session_id claim
- SHA-256 token hashing for sessions + verifications + password resets

### Learning & Resources (migration 00012)
- Entry CRUD with full-text search, type/tag filtering
- Resource CRUD with 6 types, metadata extraction (GitHub owner/repo, YouTube ID)
- Entry-resource relationships (attach/detach)
- Timeline view, search across entries + resources

### Content Generation (migration 00013) — JUST ADDED
- Tweet lifecycle: DRAFT -> REVIEWING -> APPROVED -> SCHEDULED -> POSTED / ARCHIVED
- Prompt template versioning (seeded defaults: tweet-summarize, tweet-thread, blog-summarize)
- LLM generation with retry + exponential backoff + jitter
- Cost estimation per model (known pricing table, defaults to $0.01/$0.03 per 1K tokens)
- Audit trail: every status transition and generation recorded in `generation_audits`
- CLI: `story tweet generate|regenerate|list|get|approve|review|reject|schedule|archive|audit`
- Service: `internal/application/content/service.go`
- Uses existing `publishing.LLMProvider` interface (Complete + Name methods)
- Prompt rendering via Go `text/template` with {{.Title}} and {{.Content}}
- Model cost table in service.go, extensible

### LLM Provider Infrastructure
- Interface in `internal/infrastructure/llm/provider.go`:
  - `Complete(ctx, prompt, opts) (*Result, error)`
  - `Name() string`
- 4 implementations: OpenAI, Gemini (Google), Anthropic (Claude), Ollama
- Factory: `NewProvider(cfg)` switches on `cfg.Provider`
- `CompleteOptions`: Model, Temperature, MaxTokens, TopP
- `Result`: Content, Model, InputTokens, OutputTokens
- Publishing system defines its own narrow `LLMProvider` interface in application layer

## Key files
- `internal/infrastructure/llm/provider.go` — Provider interface + factory
- `internal/application/content/service.go` — Tweet generation with retry + cost
- `internal/application/publishing/llm.go` — LLM publisher wrapping provider
- `internal/interfaces/cli/tweet.go` — Tweet CLI commands
- `internal/infrastructure/repository/tweet_repo.go` — Tweet + audit persistence
- `internal/infrastructure/repository/prompt_repo.go` — Prompt template persistence
- `migrations/00013_content_generation.sql` — Schema + seed data

## Testing
- Tests in `*_test.go` alongside source files
- Domain `Now()` is a var for testability
- Repositories tested against real PostgreSQL via test helpers
