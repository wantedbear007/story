-- +goose Up
-- +goose StatementBegin

CREATE TABLE entry_collections (
    collection_id UUID NOT NULL REFERENCES collections(id) ON DELETE CASCADE,
    entry_id      UUID NOT NULL REFERENCES entries(id) ON DELETE CASCADE,
    PRIMARY KEY (collection_id, entry_id)
);

CREATE INDEX idx_entry_collections_entry_id ON entry_collections (entry_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS entry_collections;
-- +goose StatementEnd
