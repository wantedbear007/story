-- +goose Up
-- +goose StatementBegin
ALTER TYPE raw_entry_source ADD VALUE 'notification_capture';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- PostgreSQL does not support removing values from an enum.
-- To reverse, you would need to create a new type, migrate data,
-- drop the old type, and rename. This is left as an exercise.
-- +goose StatementEnd
