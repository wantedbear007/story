-- +goose Up
-- +goose StatementBegin

CREATE TABLE tags (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name        VARCHAR(100) NOT NULL,
    color       VARCHAR(7) DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE (user_id, name)
);

CREATE INDEX idx_tags_user_id ON tags (user_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS tags;
-- +goose StatementEnd
