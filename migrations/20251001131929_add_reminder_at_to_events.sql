-- +goose Up
-- +goose StatementBegin
ALTER TABLE events
    ADD COLUMN reminder_at TIMESTAMP NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE events
    DROP COLUMN IF EXISTS reminder_at;
-- +goose StatementEnd
