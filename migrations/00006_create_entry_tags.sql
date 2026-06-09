-- +goose Up
-- +goose StatementBegin

CREATE TABLE entry_tags (
    entry_id  UUID NOT NULL REFERENCES entries(id) ON DELETE CASCADE,
    tag_id    UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (entry_id, tag_id)
);

CREATE INDEX idx_entry_tags_tag_id ON entry_tags (tag_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS entry_tags;
-- +goose StatementEnd
