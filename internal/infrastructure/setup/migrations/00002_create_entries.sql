-- +goose Up
-- +goose StatementBegin

CREATE TYPE entry_type AS ENUM ('learning', 'work_log', 'resource', 'engineering_note');

CREATE TABLE entries (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type        entry_type NOT NULL,
    title       VARCHAR(500) NOT NULL,
    content     TEXT NOT NULL,
    metadata    JSONB DEFAULT '{}',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

CREATE INDEX idx_entries_user_id ON entries (user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_entries_type ON entries (user_id, type) WHERE deleted_at IS NULL;
CREATE INDEX idx_entries_created_at ON entries (user_id, created_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_entries_search ON entries USING GIN (to_tsvector('english', title || ' ' || content)) WHERE deleted_at IS NULL;
CREATE INDEX idx_entries_deleted_at ON entries (deleted_at);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS entries;
DROP TYPE IF EXISTS entry_type;
-- +goose StatementEnd
