-- +goose Up
-- +goose StatementBegin

CREATE TYPE publishing_target_type AS ENUM ('twitter', 'notion', 'google_doc', 'blog', 'markdown');

CREATE TABLE publishing_targets (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type        publishing_target_type NOT NULL,
    name        VARCHAR(100) NOT NULL,
    config      JSONB NOT NULL DEFAULT '{}',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_publishing_targets_user_id ON publishing_targets (user_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS publishing_targets;
DROP TYPE IF EXISTS publishing_target_type;
-- +goose StatementEnd
