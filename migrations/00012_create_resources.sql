-- +goose Up
-- +goose StatementBegin

CREATE TYPE resource_type AS ENUM ('url', 'github', 'article', 'youtube', 'pdf', 'markdown');

CREATE TABLE resources (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type        resource_type NOT NULL,
    title       VARCHAR(500) NOT NULL,
    url         TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    metadata    JSONB DEFAULT '{}',
    content_hash VARCHAR(64),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_resources_user_id ON resources (user_id);
CREATE INDEX idx_resources_type ON resources (user_id, type);
CREATE INDEX idx_resources_created_at ON resources (user_id, created_at DESC);
CREATE INDEX idx_resources_search ON resources USING GIN (to_tsvector('english', title || ' ' || description));
CREATE INDEX idx_resources_content_hash ON resources (content_hash) WHERE content_hash IS NOT NULL;
CREATE INDEX idx_resources_metadata ON resources USING GIN (metadata);

CREATE TABLE entry_resources (
    entry_id    UUID NOT NULL REFERENCES entries(id) ON DELETE CASCADE,
    resource_id UUID NOT NULL REFERENCES resources(id) ON DELETE CASCADE,
    PRIMARY KEY (entry_id, resource_id)
);

CREATE INDEX idx_entry_resources_resource_id ON entry_resources (resource_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS entry_resources;
DROP TABLE IF EXISTS resources;
DROP TYPE IF EXISTS resource_type;
-- +goose StatementEnd
