-- +goose Up
-- +goose StatementBegin

CREATE TYPE publish_status AS ENUM ('pending', 'published', 'failed');

CREATE TABLE published_entries (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entry_id      UUID NOT NULL REFERENCES entries(id) ON DELETE CASCADE,
    target_id     UUID NOT NULL REFERENCES publishing_targets(id) ON DELETE CASCADE,
    external_url  TEXT DEFAULT '',
    status        publish_status NOT NULL DEFAULT 'pending',
    error_message TEXT DEFAULT '',
    published_at  TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_published_entries_entry_id ON published_entries (entry_id);
CREATE INDEX idx_published_entries_target_id ON published_entries (target_id);
CREATE INDEX idx_published_entries_status ON published_entries (status);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS published_entries;
DROP TYPE IF EXISTS publish_status;
-- +goose StatementEnd
