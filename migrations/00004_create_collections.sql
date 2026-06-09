-- +goose Up
-- +goose StatementBegin

CREATE TABLE collections (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name        VARCHAR(200) NOT NULL,
    description TEXT DEFAULT '',
    parent_id   UUID REFERENCES collections(id) ON DELETE SET NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

CREATE INDEX idx_collections_user_id ON collections (user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_collections_parent_id ON collections (parent_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_collections_deleted_at ON collections (deleted_at);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS collections;
-- +goose StatementEnd
