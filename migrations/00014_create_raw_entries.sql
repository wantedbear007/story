-- +goose Up
-- +goose StatementBegin

CREATE TYPE raw_entry_status AS ENUM ('raw', 'processing', 'structured', 'archived');
CREATE TYPE raw_entry_source AS ENUM ('cli', 'file', 'pipe', 'import', 'api');

CREATE TABLE raw_entries (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content    TEXT NOT NULL,
    status     raw_entry_status NOT NULL DEFAULT 'raw',
    source     raw_entry_source NOT NULL DEFAULT 'cli',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_raw_entries_user_id ON raw_entries (user_id);
CREATE INDEX idx_raw_entries_status ON raw_entries (user_id, status);
CREATE INDEX idx_raw_entries_created_at ON raw_entries (user_id, created_at DESC);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS raw_entries;
DROP TYPE IF EXISTS raw_entry_source;
DROP TYPE IF EXISTS raw_entry_status;
-- +goose StatementEnd
